package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/tui"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize secrets based on config",
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

		fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  sync"))
		fmt.Println(ui.Divider(44))
		fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(targetProfile.Name))
		fmt.Println()

		existingSecrets := make(map[string]string)

		for _, key := range targetProfile.AutoGen {
			existingSecrets[key], _ = cryptoModule.GenerateSecurePassword(100)
			fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Muted.Render("generated"), ui.Key.Render(key))
		}

		if len(targetProfile.Required) > 0 {
			p := tea.NewProgram(tui.NewModel(targetProfile.Required))
			finalModel, err := p.Run()
			if err != nil {
				fmt.Println(ui.FormatError("TUI crashed", err.Error()))
				os.Exit(1)
			}

			m := finalModel.(tui.Model)
			if m.Cancelled {
				fmt.Println(ui.Muted.Render("Operation cancelled. No secrets were saved."))
				return
			}

			for k, v := range m.Choices {
				existingSecrets[k] = v
			}
		}

		if len(existingSecrets) > 0 {
			storage := provider.GetStorageProvider(appConfig, parsedManifest.Project.Provider, parsedManifest.Project.Name)

			err = cryptoModule.SaveSecrets(storage, existingSecrets, identity.Recipient())
			if err != nil {
				fmt.Println(ui.FormatError("Failed to save secrets", err.Error()))
				os.Exit(1)
			}

			fmt.Println()
			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf("Secrets encrypted and saved via %s", parsedManifest.Project.Provider)))
			return
		}

		fmt.Println(ui.Muted.Render("No secrets to configure."))
	},
}

func init() {
	syncCmd.Flags().StringP("profile", "p", "default", "Profile to select")
	rootCmd.AddCommand(syncCmd)
}
