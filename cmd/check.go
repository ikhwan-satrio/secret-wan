package cmd

import (
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check secret completeness based on config",
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
			fmt.Println(ui.FormatError("Failed to load configuration", fmt.Sprintf("Invalid manifest declaration: %v", err)))
			os.Exit(1)
		}

		targetProfile, exists := parsedManifest.Profiles[profileName]
		if !exists {
			validProfiles := manifest.GetProfileNames(parsedManifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' not found. Available profiles: %v", profileName, validProfiles)
			fmt.Println(ui.FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		expectedSecrets := append(targetProfile.Required, targetProfile.AutoGen...)
		if len(expectedSecrets) == 0 {
			fmt.Println(ui.FormatError("Failed to load configuration", fmt.Sprintf("Profile '%s' must define at least one secret", targetProfile.Name)))
			os.Exit(1)
		}

		storage := provider.GetStorageProvider(appConfig, parsedManifest.Project.Provider, parsedManifest.Project.Name)
		actualSecrets, err := cryptoModule.LoadSecrets(storage, identity)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.FormatError("Failed to load secrets", fmt.Sprintf("Storage not found. Run 'secwan sync -p %s' first.", profileName)))
			} else {
				fmt.Println(ui.FormatError("Failed to decrypt secrets", err.Error()))
			}
			os.Exit(1)
		}

		fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  check"))
		fmt.Println(ui.Divider(44))
		fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(targetProfile.Name))
		fmt.Println()

		allPresent := true
		for _, key := range expectedSecrets {
			keyColumn := fmt.Sprintf("%-24s", key)

			if val, exists := actualSecrets[key]; exists && val != "" {
				fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Key.Render(keyColumn), ui.Muted.Render("configured"))
			} else {
				fmt.Printf("  %s %s %s\n", ui.IconError, ui.Key.Render(keyColumn), ui.Error.Render("missing"))
				allPresent = false
			}
		}

		fmt.Println()
		fmt.Println(ui.Divider(44))

		if allPresent {
			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("All secrets are properly configured!"))
			return
		}

		fmt.Printf("%s %s\n", ui.IconError, ui.Error.Render("Some secrets are missing."))
		fmt.Println(ui.Muted.Render(fmt.Sprintf("Run 'secwan sync -p %s' to configure them.", profileName)))
		os.Exit(1)
	},
}

func init() {
	checkCmd.Flags().StringP("profile", "p", "default", "Profile to check")
	rootCmd.AddCommand(checkCmd)
}
