package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var printEnvCmd = &cobra.Command{
	Use:   "print-env",
	Short: "Display decrypted secret values in KEY=VALUE format",
	Run: func(cmd *cobra.Command, args []string) {
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			fmt.Println(ui.FormatError("Failed to read profile flag", err.Error()))
			os.Exit(1)
		}

		appConfig := config.UseApp()
		cryptoModule := crypto.CryptoModule{AppConfig: appConfig}
		parserModule := manifest.ParserModule{AppConfig: appConfig}

		identity, err := cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(ui.FormatError("Failed to load master key", err.Error()))
			os.Exit(1)
		}

		parsedManifest, err := parserModule.ParseManifest()
		if err != nil {
			fmt.Println(ui.FormatError("Failed to parse manifest", err.Error()))
			os.Exit(1)
		}

		targetProfile, exists := parsedManifest.Profiles[profileName]
		if !exists {
			validProfiles := manifest.GetProfileNames(parsedManifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' not found. Available profiles: %v", profileName, validProfiles)
			fmt.Println(ui.FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		storage := provider.GetStorageProvider(appConfig, parsedManifest.Project.Provider, parsedManifest.Project.Name)
		secrets, err := cryptoModule.LoadSecrets(storage, identity)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.FormatError("Secrets storage not found",
					fmt.Sprintf("Run 'secwan sync -p %s' first.", profileName)))
			} else {
				fmt.Println(ui.FormatError("Failed to decrypt secrets", err.Error()))
			}
			os.Exit(1)
		}

		allowedKeys := make(map[string]bool)
		for _, k := range targetProfile.Required {
			allowedKeys[k] = true
		}
		for _, k := range targetProfile.AutoGen {
			allowedKeys[k] = true
		}

		var activeKeys []string
		for k := range allowedKeys {
			if _, exists := secrets[k]; exists {
				activeKeys = append(activeKeys, k)
			}
		}
		sort.Strings(activeKeys)

		if len(activeKeys) == 0 {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("# No secrets configured for profile '%s'", targetProfile.Name)))
			return
		}

		fmt.Printf("%s %s\n", ui.Muted.Render("# Secrets for profile:"), ui.Accent.Render(targetProfile.Name))
		fmt.Printf("%s\n", ui.Warning.Render("# WARNING: This output contains sensitive data. Do not share screenshots."))
		fmt.Println()

		for _, k := range activeKeys {
			fmt.Printf("%s=\"%s\"\n", k, secrets[k])
		}
	},
}

func init() {
	printEnvCmd.Flags().StringP("profile", "p", "default", "Profile environment to display")
	rootCmd.AddCommand(printEnvCmd)
}
