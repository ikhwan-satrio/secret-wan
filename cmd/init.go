package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/internal/tui"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize master key and config template",
	Run: func(cmd *cobra.Command, args []string) {
		session := newSession()

		fmt.Println(ui.Title.Render("secret-wan") + ui.Muted.Render("  •  initialize"))
		fmt.Println(ui.Divider(44))
		fmt.Println()

		p := tea.NewProgram(tui.SelectModel{
			Title:   "Select storage provider:",
			Choices: []string{"dotfiles (Local encrypted file)", "keyring (OS Keychain)"},
		})

		finalModel, err := p.Run()
		if err != nil {
			failErr("TUI crashed", err)
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

		session.requireIdentity()

		created, err := session.CreateManifest(providerName)
		if err != nil {
			failErr("Failed to create config file", err)
		}

		if !created {
			fmt.Printf("%s %s\n", ui.IconInfo, ui.Muted.Render(fmt.Sprintf("%s already exists, skipped creation.", session.Config.SpecFile)))
		} else {
			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf("Created master key and %s", session.Config.SpecFile)))

			// A new project starts with the current key as its only recipient.
			// Existing projects keep their single-recipient behavior until the
			// recipient list is created explicitly.
			createdRecipients, recipientErr := session.CreateRecipientList()
			switch {
			case recipientErr != nil:
				warn(fmt.Sprintf("Could not create %s: %v", session.Config.RecipientsPath(), recipientErr))
			case createdRecipients:
				fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf("Created recipient list %s", session.Config.RecipientsPath())))
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
