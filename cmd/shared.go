package cmd

import (
	"fmt"
	"os"
	"sort"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

// appSession bundles the pieces every command needs: configuration, crypto,
// manifest parsing, and the selected storage provider.
type appSession struct {
	Config   config.AppConfig
	Crypto   crypto.CryptoModule
	Parser   manifest.ParserModule
	Manifest *manifest.Manifest
	Identity *age.X25519Identity
	Storage  provider.StorageProvider
}

func newSession() *appSession {
	appConfig := config.UseApp()

	return &appSession{
		Config: appConfig,
		Crypto: crypto.CryptoModule{AppConfig: appConfig},
		Parser: manifest.ParserModule{AppConfig: appConfig},
	}
}

// recipientsStore returns the store for the project's public recipient list.
func (s *appSession) recipientsStore() recipients.Store {
	return recipients.New(s.Config.RecipientsPath())
}

// recipientKeys loads the configured public recipients. A missing file means
// the project still uses the current key as its only recipient.
func (s *appSession) recipientKeys() ([]string, error) {
	return s.recipientsStore().Load()
}

// encryptionRecipients resolves the configured recipient list into age
// recipients and, unless allowSelfExclusion is set, refuses to lock the current
// key out of future writes.
func (s *appSession) encryptionRecipients(allowSelfExclusion bool) ([]age.Recipient, error) {
	return encryptionRecipients(s.Config, s.Identity, allowSelfExclusion)
}

// saveSecrets encrypts and saves the full secret map for all configured
// recipients.
func (s *appSession) saveSecrets(secrets map[string]string, allowSelfExclusion bool) error {
	recips, err := s.encryptionRecipients(allowSelfExclusion)
	if err != nil {
		return err
	}

	return s.Crypto.SaveSecrets(s.Storage, secrets, recips...)
}

// rekey decrypts the current store and writes it back for the configured
// recipients.
func (s *appSession) rekey(allowSelfExclusion bool) (int, error) {
	return rekeySecrets(s.Config, s.Crypto, s.Storage, s.Identity, allowSelfExclusion)
}

// encryptionRecipients is the package-level variant used by commands that do
// not build an appSession, such as sync.
func encryptionRecipients(appConfig config.AppConfig, identity *age.X25519Identity, allowSelfExclusion bool) ([]age.Recipient, error) {
	if identity == nil {
		return nil, fmt.Errorf("master key is not loaded")
	}

	keys, err := recipients.New(appConfig.RecipientsPath()).Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read recipients: %w", err)
	}

	self := identity.Recipient().String()
	if len(keys) == 0 {
		keys = []string{self}
	}

	if !allowSelfExclusion && !recipients.Contains(keys, self) {
		return nil, fmt.Errorf(
			"current key %s is not a recipient; run 'secwan recipient add %s' first",
			self, self,
		)
	}

	return recipients.Parse(keys)
}

// rekeySecrets re-encrypts the stored map for the configured recipients. A
// missing store is treated as "nothing to rekey".
func rekeySecrets(appConfig config.AppConfig, cryptoModule crypto.CryptoModule, storage provider.StorageProvider, identity *age.X25519Identity, allowSelfExclusion bool) (int, error) {
	if identity == nil {
		return 0, fmt.Errorf("master key is not loaded")
	}

	secrets, err := cryptoModule.LoadSecrets(storage, identity)
	if err != nil {
		if provider.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}

	recips, err := encryptionRecipients(appConfig, identity, allowSelfExclusion)
	if err != nil {
		return 0, err
	}

	if err := cryptoModule.SaveSecrets(storage, secrets, recips...); err != nil {
		return 0, err
	}

	return len(secrets), nil
}

// fail prints a styled error to stderr and terminates the process.
func fail(shortMsg, detailMsg string) {
	fmt.Fprintln(os.Stderr, ui.FormatError(shortMsg, detailMsg))
	os.Exit(1)
}

func failErr(shortMsg string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	fail(shortMsg, detail)
}

func printHeader(action string) {
	fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  "+action))
	fmt.Println(ui.Divider(44))
}

// requireManifest parses the manifest and resolves the storage provider.
func (s *appSession) requireManifest() {
	parsed, err := s.Parser.ParseManifest()
	if err != nil {
		failErr("Failed to parse manifest", err)
	}

	s.Manifest = parsed
	s.Storage = provider.GetStorageProvider(s.Config, parsed.Project.Provider, parsed.Project.Name)
}

// requireIdentity loads (or creates) the age identity used for encryption.
func (s *appSession) requireIdentity() {
	identity, err := s.Crypto.LoadOrGenerateKey()
	if err != nil {
		failErr("Failed to load master key", err)
	}

	s.Identity = identity
}

// requireExistingIdentity loads the age identity without creating a new one.
// Read-only commands use this so a missing key is reported instead of silently
// generating a key that cannot decrypt existing storage.
func (s *appSession) requireExistingIdentity() {
	identity, err := s.Crypto.LoadKey()
	if err != nil {
		if os.IsNotExist(err) {
			fail("Master key not found", "Run 'secwan init' or 'secwan set' first.")
		}
		failErr("Failed to load master key", err)
	}

	s.Identity = identity
}

func (s *appSession) requireProfile(profileName string) *manifest.Profile {
	profile, exists := s.Manifest.Profiles[profileName]
	if !exists {
		validProfiles := manifest.GetProfileNames(s.Manifest.Profiles)
		fail("Invalid profile", fmt.Sprintf("Profile '%s' not found. Available profiles: %v", profileName, validProfiles))
	}

	return profile
}

// requireSecrets loads stored secrets and exits with a helpful message when
// the storage has not been initialized yet.
func (s *appSession) requireSecrets() map[string]string {
	secrets, err := s.Crypto.LoadSecrets(s.Storage, s.Identity)
	if err != nil {
		if provider.IsNotFound(err) {
			fail("Secrets storage not found", "Run 'secwan sync' or 'secwan set <KEY>' to create it first.")
		}
		failErr("Failed to decrypt secrets", err)
	}

	if secrets == nil {
		secrets = make(map[string]string)
	}

	return secrets
}

// loadSecretsAllowMissing behaves like requireSecrets but returns an empty map
// when no storage exists yet. This lets `set` create the first secret.
func (s *appSession) loadSecretsAllowMissing() map[string]string {
	secrets, err := s.Crypto.LoadSecrets(s.Storage, s.Identity)
	if err == nil {
		if secrets == nil {
			secrets = make(map[string]string)
		}
		return secrets
	}

	if provider.IsNotFound(err) {
		fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning, ui.Warning.Render("No existing secrets found; a new storage will be created."))
		return make(map[string]string)
	}

	failErr("Failed to decrypt secrets", err)
	return nil
}

// profileKeys returns the de-duplicated, sorted list of keys managed by a
// profile. A key may be declared as both required and generated; it should
// still only appear once.
func profileKeys(profile *manifest.Profile) []string {
	set := make(map[string]bool, len(profile.Required)+len(profile.AutoGen))

	for _, key := range profile.Required {
		set[key] = true
	}
	for _, key := range profile.AutoGen {
		set[key] = true
	}

	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

// requiredKeysForPrompt returns required keys that are not auto-generated.
// A key declared as both should be generated, not prompted twice.
func requiredKeysForPrompt(profile *manifest.Profile) []string {
	generated := make(map[string]bool, len(profile.AutoGen))
	for _, key := range profile.AutoGen {
		generated[key] = true
	}

	keys := make([]string, 0, len(profile.Required))
	for _, key := range profile.Required {
		if !generated[key] {
			keys = append(keys, key)
		}
	}

	return keys
}

func keyInProfile(profile *manifest.Profile, key string) bool {
	if _, ok := profile.Rules[key]; ok {
		return true
	}

	for _, candidate := range profile.Required {
		if candidate == key {
			return true
		}
	}
	for _, candidate := range profile.AutoGen {
		if candidate == key {
			return true
		}
	}

	return false
}

func ruleFor(profile *manifest.Profile, key string) manifest.SecretRule {
	if profile.Rules == nil {
		return manifest.SecretRule{}
	}

	return profile.Rules[key]
}

// allDeclaredKeys returns every key mentioned by any profile. Stored keys that
// are not in this set are considered unmanaged.
func allDeclaredKeys(parsed *manifest.Manifest) map[string]bool {
	declared := make(map[string]bool)

	for _, profile := range parsed.Profiles {
		for _, key := range profile.Declared {
			declared[key] = true
		}
		// Defensive fallback for profiles constructed without Declared.
		for _, key := range profile.Required {
			declared[key] = true
		}
		for _, key := range profile.AutoGen {
			declared[key] = true
		}
	}

	return declared
}

func profilesContaining(parsed *manifest.Manifest, key string) []string {
	var names []string
	for name, profile := range parsed.Profiles {
		if keyInProfile(profile, key) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
}
