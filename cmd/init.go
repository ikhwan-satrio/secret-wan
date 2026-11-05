package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
	"github.com/ikhwan-satrio/secret-wan/internal/tui"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize master key and config template",
	Run: func(cmd *cobra.Command, args []string) {
		appConfig := config.UseApp()
		cryptoModule := crypto.CryptoModule{AppConfig: appConfig}

		fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  initialize"))
		fmt.Println(ui.Divider(44))
		fmt.Println()

		p := tea.NewProgram(tui.SelectModel{
			Title:   "Select storage provider:",
			Choices: []string{"dotfiles (Local encrypted file)", "keyring (OS Keychain)"},
		})

		finalModel, err := p.Run()
		if err != nil {
			fmt.Println(ui.FormatError("TUI crashed", err.Error()))
			os.Exit(1)
		}

		selModel := finalModel.(tui.SelectModel)
		if selModel.Cancel {
			fmt.Println(ui.Muted.Render("Initialization cancelled."))
			return
		}

		providerName := "dotfiles"
		if strings.HasPrefix(selModel.Result, "keyring") {
			providerName = "keyring"
		}

		identity, err := cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(ui.FormatError("Failed to initialize master key", err.Error()))
			os.Exit(1)
		}

		specTemplate := manifest.GetTemplate(providerName)

		if _, err := os.Stat(appConfig.SpecFile); err == nil {
			fmt.Printf("%s %s\n", ui.IconInfo, ui.Muted.Render(fmt.Sprintf("%s already exists, skipped creation.", appConfig.SpecFile)))
		} else {
			if err := os.WriteFile(appConfig.SpecFile, []byte(specTemplate), 0644); err != nil {
				fmt.Println(ui.FormatError("Failed to create config file", err.Error()))
				os.Exit(1)
			}

			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf("Created master key and %s", appConfig.SpecFile)))

			// A new project starts with the current key as its only recipient.
			// Existing projects keep their single-recipient behavior until the
			// recipient list is created explicitly.
			recipientStore := recipients.New(appConfig.RecipientsPath())
			if exists, statErr := recipientStore.Exists(); statErr != nil {
				fmt.Printf("%s %s\n", ui.IconWarning, ui.Warning.Render(fmt.Sprintf("Could not inspect %s: %v", appConfig.RecipientsPath(), statErr)))
			} else if !exists {
				if saveErr := recipientStore.Save([]string{identity.Recipient().String()}); saveErr != nil {
					fmt.Printf("%s %s\n", ui.IconWarning, ui.Warning.Render(fmt.Sprintf("Could not create %s: %v", appConfig.RecipientsPath(), saveErr)))
				} else {
					fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf("Created recipient list %s", appConfig.RecipientsPath())))
				}
			}
		}

		fmt.Println()
		fmt.Println(ui.Title.Render("Next steps"))
		fmt.Println(ui.Step(1, "secwan sync", "verify and set your secrets interactively"))
		fmt.Println(ui.Step(2, "secwan check", "verify all secrets are present and valid"))
		fmt.Println(ui.Step(3, "secwan exec -- <cmd>", "run your app with secrets injected"))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
