package cmd

import (
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/config"
	"github.com/ikhwan-satrio/secret-wan/utils"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Memeriksa kelengkapan secret berdasarkan config",
	Run: func(cmd *cobra.Command, args []string) {
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			fmt.Println(FormatError("Failed to read profile flag", err.Error()))
			os.Exit(1)
		}

		appConfig := config.UseApp()
		cryptoModule := utils.CryptoModule{AppConfig: appConfig}
		parserModule := utils.ParserModule{AppConfig: appConfig}

		identity, err := cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(FormatError("Failed to load master key", err.Error()))
			os.Exit(1)
		}

		manifest, err := parserModule.ParseManifest()
		if err != nil {
			fmt.Println(FormatError("Failed to load configuration", fmt.Sprintf("Invalid manifest declaration: %v", err)))
			os.Exit(1)
		}

		targetProfile, exists := manifest.Profiles[profileName]
		if !exists {
			validProfiles := getProfileNames(manifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' tidak ditemukan. Profile yang tersedia: %v", profileName, validProfiles)
			fmt.Println(FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		expectedSecrets := append(targetProfile.Required, targetProfile.AutoGen...)
		if len(expectedSecrets) == 0 {
			fmt.Println(FormatError("Failed to load configuration", fmt.Sprintf("Profile '%s' must define at least one secret", targetProfile.Name)))
			os.Exit(1)
		}

		// Load menggunakan Storage Provider
		provider := utils.GetStorageProvider(appConfig, manifest.Project.Provider, manifest.Project.Name)
		actualSecrets, err := cryptoModule.LoadSecrets(provider, identity)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(FormatError("Failed to load secrets", fmt.Sprintf("Storage not found. Run 'secman sync -p %s' first.", profileName)))
			} else {
				fmt.Println(FormatError("Failed to decrypt secrets", err.Error()))
			}
			os.Exit(1)
		}

		fmt.Println(StyleSuccess.Render(fmt.Sprintf("\n✓ Profile '%s' Status", targetProfile.Name)))
		fmt.Println(StyleDim.Render("────────────────────────────────────────"))

		allPresent := true
		for _, key := range expectedSecrets {
			if val, exists := actualSecrets[key]; exists && val != "" {
				fmt.Printf("  %s %-25s %s\n", IconSuccess, key, StyleDim.Render("(configured)"))
			} else {
				fmt.Printf("  %s %-25s %s\n", IconError, key, StyleError.Render("(missing)"))
				allPresent = false
			}
		}

		fmt.Println(StyleDim.Render("────────────────────────────────────────"))

		if allPresent {
			fmt.Printf("\n%s %s\n", IconSuccess, StyleSuccess.Render("All secrets are properly configured!"))
		} else {
			fmt.Printf("\n%s %s\n", IconError, StyleError.Render("Some secrets are missing."))
			fmt.Println(StyleDim.Render(fmt.Sprintf("Run 'secman sync -p %s' to configure them.", profileName)))
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringP("profile", "p", "default", "Profile to check")
	rootCmd.AddCommand(checkCmd)
}
