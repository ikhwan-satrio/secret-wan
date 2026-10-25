package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/config"
	"github.com/ikhwan-satrio/secret-wan/tui"
	"github.com/ikhwan-satrio/secret-wan/utils"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sinkronisasi secret berdasarkan config",
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
			fmt.Println(FormatError("Failed to parse manifest", err.Error()))
			os.Exit(1)
		}

		// Direct lookup, tidak perlu loop
		targetProfile, exists := manifest.Profiles[profileName]
		if !exists {
			validProfiles := getProfileNames(manifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' tidak ditemukan. Profile yang tersedia: %v", profileName, validProfiles)
			fmt.Println(FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		fmt.Println(StyleDim.Render(fmt.Sprintf("Configuring profile: %s", targetProfile.Name)))
		existingSecrets := make(map[string]string)

		// 1. Auto-generate
		for _, key := range targetProfile.AutoGen {
			existingSecrets[key], _ = cryptoModule.GenerateSecurePassword(100)
			fmt.Printf("  %s %s\n", IconSuccess, StyleDim.Render(fmt.Sprintf("Generated: %s", key)))
		}

		// 2. Manual input via TUI
		if len(targetProfile.Required) > 0 {
			p := tea.NewProgram(tui.NewModel(targetProfile.Required))
			finalModel, err := p.Run()
			if err != nil {
				fmt.Println(FormatError("TUI crashed", err.Error()))
				os.Exit(1)
			}

			m := finalModel.(tui.Model)
			if m.Cancelled {
				fmt.Println("\n" + StyleDim.Render("  Operation cancelled. No secrets were saved."))
				return
			}

			for k, v := range m.Choices {
				existingSecrets[k] = v
			}
		}

		// 3. Save menggunakan Storage Provider yang benar
		if len(existingSecrets) > 0 {
			provider := utils.GetStorageProvider(appConfig, manifest.Project.Provider, manifest.Project.Name)

			err = cryptoModule.SaveSecrets(provider, existingSecrets, identity.Recipient())
			if err != nil {
				fmt.Println(FormatError("Failed to save secrets", err.Error()))
				os.Exit(1)
			}

			fmt.Printf("\n  %s %s\n", IconSuccess, StyleDim.Render(fmt.Sprintf("Encrypted and saved via %s", manifest.Project.Provider)))
		} else {
			fmt.Println("\n" + StyleDim.Render("  No secrets to configure."))
		}
	},
}

func init() {
	syncCmd.Flags().StringP("profile", "p", "default", "Profile to select")
	rootCmd.AddCommand(syncCmd)
}
