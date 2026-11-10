package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/buildinfo"
	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

func run() {
	appConfig := config.UseApp()
	cryptoModule := crypto.CryptoModule{AppConfig: appConfig}
	parserModule := manifest.ParserModule{AppConfig: appConfig}

	cli.PrintHeader("doctor")
	fmt.Println(ui.Muted.Render("Checking your secret-wan setup..."))
	fmt.Println()

	report := &doctorReport{}
	build := buildinfo.Current()

	environment := report.section("Environment")
	report.ok(environment, "Platform", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	report.ok(environment, "Go runtime", runtime.Version())
	report.ok(environment, "Version", fmt.Sprintf("%s (commit %s, built %s)",
		buildinfo.Display(), build.Commit, build.Date))

	parsed, err := parserModule.ParseManifest()
	manifestSection := report.section("Manifest")
	if err != nil {
		report.fail(manifestSection, appConfig.SpecFile, err.Error())
		report.finish()
		return
	}

	report.ok(manifestSection, "File", appConfig.SpecFile)

	if strings.TrimSpace(parsed.Project.Name) == "" {
		report.fail(manifestSection, "Project name", "is empty")
	} else {
		report.ok(manifestSection, "Project name", parsed.Project.Name)
	}

	providerName := parsed.Project.Provider
	if providerName == "" {
		providerName = "dotfiles"
	}
	if provider.IsKnown(providerName) {
		report.ok(manifestSection, "Provider", providerName)
		if provider.IsReadOnly(providerName) {
			report.ok(manifestSection, "Provider mode", "read-only")
		}
	} else {
		report.fail(manifestSection, "Provider",
			fmt.Sprintf("unknown provider '%s' (expected %s)", providerName, strings.Join(provider.Names(), ", ")))
	}

	profileNames := manifest.GetProfileNames(parsed.Profiles)
	report.ok(manifestSection, "Profiles",
		fmt.Sprintf("%d defined: %s", len(profileNames), strings.Join(profileNames, ", ")))

	keySection := report.section("Master key")
	keyPath, err := crypto.KeyPath(appConfig.AppName)
	if err != nil {
		report.fail(keySection, "Location", err.Error())
		report.finish()
		return
	}

	report.ok(keySection, "Location", keyPath)

	if dirInfo, statErr := os.Stat(filepath.Dir(keyPath)); statErr == nil && runtime.GOOS != "windows" {
		if dirInfo.Mode().Perm()&0077 != 0 {
			report.warn(keySection, "Directory permissions",
				fmt.Sprintf("%04o (recommended 0700)", dirInfo.Mode().Perm()))
		}
	}

	var identity *age.X25519Identity
	info, statErr := os.Stat(keyPath)
	switch {
	case statErr != nil:
		if os.IsNotExist(statErr) {
			report.fail(keySection, "File", "not found; run 'secwan init' or 'secwan set' first")
		} else {
			report.fail(keySection, "File", statErr.Error())
		}
	case info.IsDir():
		report.fail(keySection, "File", "is a directory")
	default:
		if runtime.GOOS != "windows" {
			permissions := info.Mode().Perm()
			if permissions&0077 != 0 {
				report.warn(keySection, "Permissions",
					fmt.Sprintf("%04o (recommended 0600)", permissions))
			} else {
				report.ok(keySection, "Permissions", fmt.Sprintf("%04o", permissions))
			}
		}

		identity = readIdentity(report, keySection, keyPath)
	}

	recipientsSection := report.section("Recipients")
	recipientKeys, recipientErr := recipients.New(appConfig.RecipientsPath()).Load()
	switch {
	case recipientErr != nil:
		report.fail(recipientsSection, "File", recipientErr.Error())
	case len(recipientKeys) == 0:
		report.ok(recipientsSection, "Recipients",
			"single-recipient mode (no recipients configured)")
	default:
		report.ok(recipientsSection, "Recipients",
			fmt.Sprintf("%d configured", len(recipientKeys)))
		if identity != nil {
			self := identity.Recipient().String()
			if recipients.Contains(recipientKeys, self) {
				report.ok(recipientsSection, "Your key", "listed")
			} else {
				report.warn(recipientsSection, "Your key",
					"not listed; run 'secwan recipient add' before the next rekey")
			}
		}
	}

	storageSection := report.section("Storage")
	storage := provider.GetStorageProvider(appConfig, parsed.Project.Provider, parsed.Project.Name)
	report.ok(storageSection, "Provider", providerName)

	storageReady := true
	if checker, ok := storage.(interface{ Check() error }); ok {
		if checkErr := checker.Check(); checkErr != nil {
			storageReady = false
			if provider.IsNotFound(checkErr) {
				report.warn(storageSection, "Storage", "not initialized; run 'secwan sync' or 'secwan set' first")
			} else {
				report.fail(storageSection, "Storage", checkErr.Error())
			}
		} else {
			report.ok(storageSection, "Storage", "available")
		}
	}

	var stored map[string]string
	if identity == nil {
		report.warn(storageSection, "Secrets", "skipped because the master key is unavailable")
	} else if storageReady {
		secrets, loadErr := cryptoModule.LoadSecrets(storage, identity)
		switch {
		case loadErr != nil && provider.IsNotFound(loadErr):
			report.warn(storageSection, "Secrets", "no secrets stored yet")
		case loadErr != nil:
			report.fail(storageSection, "Decryption", loadErr.Error())
		default:
			stored = secrets
			report.ok(storageSection, "Secrets", fmt.Sprintf("%d stored", len(secrets)))
		}
	}

	checkProfiles(report, parsed, profileNames, stored)

	report.finish()
}

func readIdentity(report *doctorReport, section *doctorSection, keyPath string) *age.X25519Identity {
	file, err := os.Open(keyPath)
	if err != nil {
		report.fail(section, "Read", err.Error())
		return nil
	}
	defer file.Close()

	identities, err := age.ParseIdentities(file)
	if err != nil {
		report.fail(section, "Parse", err.Error())
		return nil
	}
	if len(identities) == 0 {
		report.fail(section, "Parse", "key file is empty")
		return nil
	}

	identity, supported := identities[0].(*age.X25519Identity)
	if !supported {
		report.fail(section, "Type", "unsupported key type")
		return nil
	}

	report.ok(section, "Recipient", identity.Recipient().String())
	return identity
}

func checkProfiles(report *doctorReport, parsed *manifest.Manifest, profileNames []string, stored map[string]string) {
	profileSection := report.section("Profiles")

	for _, name := range profileNames {
		profile := parsed.Profiles[name]
		expected := profile.Keys()

		if len(expected) == 0 {
			report.warn(profileSection, name, "no secrets defined")
			continue
		}
		if stored == nil {
			report.warn(profileSection, name, "skipped because secrets are unavailable")
			continue
		}

		var missing []string
		for _, key := range expected {
			if strings.TrimSpace(stored[key]) == "" {
				missing = append(missing, key)
			}
		}

		if len(missing) == 0 {
			report.ok(profileSection, name,
				fmt.Sprintf("%d/%d secrets configured", len(expected), len(expected)))
		} else {
			report.fail(profileSection, name,
				fmt.Sprintf("%d/%d configured; missing: %s",
					len(expected)-len(missing), len(expected), strings.Join(missing, ", ")))
		}
	}

	for _, name := range profileNames {
		profile := parsed.Profiles[name]
		for _, key := range profile.Declared {
			rule := profile.Rule(key)
			switch {
			case !rule.Required && !rule.Generate:
				report.warn(profileSection, key,
					fmt.Sprintf("profile '%s' has no rule (use required or generate)", name))
			case rule.Required && rule.Generate:
				report.warn(profileSection, key,
					fmt.Sprintf("profile '%s' marks this key as both required and generate", name))
			}
		}
	}

	consistency := report.section("Consistency")
	if stored == nil {
		report.warn(consistency, "Unmanaged secrets", "skipped because secrets are unavailable")
		return
	}

	declared := parsed.DeclaredKeys()

	var extra []string
	for key := range stored {
		if !declared[key] {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)

	if len(extra) == 0 {
		report.ok(consistency, "Unmanaged secrets", "none")
	} else {
		report.warn(consistency, "Unmanaged secrets", strings.Join(extra, ", "))
	}

	var empty []string
	for key, value := range stored {
		if strings.TrimSpace(value) == "" {
			empty = append(empty, key)
		}
	}
	sort.Strings(empty)
	if len(empty) > 0 {
		report.warn(consistency, "Empty values", strings.Join(empty, ", "))
	}
}
