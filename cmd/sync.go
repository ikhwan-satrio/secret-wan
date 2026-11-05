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

		storage := provider.GetStorageProvider(appConfig, parsedManifest.Project.Provider, parsedManifest.Project.Name)

		// Preserve secrets from other profiles and avoid overwriting existing
		// generated values when sync is run more than once.
		stored := make(map[string]string)
		current, loadErr := cryptoModule.LoadSecrets(storage, identity)
		switch {
		case loadErr == nil:
			stored = current
		case provider.IsNotFound(loadErr):
			// First sync: start with an empty store.
		default:
			fmt.Println(ui.FormatError("Failed to decrypt existing secrets", loadErr.Error()))
			os.Exit(1)
		}

		changed := false

		for _, key := range targetProfile.AutoGen {
			if existing, ok := stored[key]; ok && existing != "" {
				fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Muted.Render("kept"), ui.Key.Render(key))
				continue
			}

			value, err := cryptoModule.GenerateSecurePassword(appConfig.SecretLength)
			if err != nil {
				fmt.Println(ui.FormatError("Failed to generate secret", fmt.Sprintf("%s: %v", key, err)))
				os.Exit(1)
			}

			stored[key] = value
			changed = true
			fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Muted.Render("generated"), ui.Key.Render(key))
		}

		promptKeys := requiredKeysForPrompt(targetProfile)
		if len(promptKeys) > 0 {
			p := tea.NewProgram(tui.NewModel(promptKeys))
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
				stored[k] = v
				changed = true
			}
		}

		if changed {
			recips, err := encryptionRecipients(appConfig, identity, false)
			if err != nil {
				fmt.Println(ui.FormatError("Failed to resolve recipients", err.Error()))
				os.Exit(1)
			}

			err = cryptoModule.SaveSecrets(storage, stored, recips...)
			if err != nil {
				fmt.Println(ui.FormatError("Failed to save secrets", err.Error()))
				os.Exit(1)
			}

			fmt.Println()
			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf(
				"Secrets encrypted and saved via %s (%d recipient(s))",
				parsedManifest.Project.Provider, len(recips),
			)))
			return
		}

		fmt.Println(ui.Muted.Render("No changes to save."))
	},
}

func init() {
	syncCmd.Flags().StringP("profile", "p", "default", "Profile to select")
	rootCmd.AddCommand(syncCmd)
}
