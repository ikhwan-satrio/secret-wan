package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ikhwan-satrio/secret-wan/internal/cli"
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
			cli.FailErr("Failed to read profile flag", err)
		}

		session := cli.NewSession()
		session.RequireIdentity()
		session.RequireProject()
		targetProfile := session.RequireProfile(profileName)

		cli.PrintHeader("sync")
		fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(targetProfile.Name))
		fmt.Println()

		// Preserve secrets from other profiles and avoid overwriting existing
		// generated values when sync is run more than once.
		stored, _ := session.LoadSecretsAllowMissing()

		changed := false

		for _, key := range targetProfile.AutoGen {
			if existing, ok := stored[key]; ok && existing != "" {
				fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Muted.Render("kept"), ui.Key.Render(key))
				continue
			}

			value, err := session.GenerateSecret()
			if err != nil {
				cli.Fail("Failed to generate secret", fmt.Sprintf("%s: %v", key, err))
			}

			stored[key] = value
			changed = true
			fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Muted.Render("generated"), ui.Key.Render(key))
		}

		promptKeys := targetProfile.PromptKeys()
		if len(promptKeys) > 0 {
			p := tea.NewProgram(tui.NewModel(promptKeys))
			finalModel, err := p.Run()
			if err != nil {
				cli.FailErr("TUI crashed", err)
			}

			m := finalModel.(tui.Model)
			if m.Cancelled {
				fmt.Println(ui.Muted.Render("Operation cancelled. No secrets were saved."))
				return
			}

			for key, value := range m.Choices {
				stored[key] = value
				changed = true
			}
		}

		if changed {
			recipientCount := session.SaveSecrets(stored, false)

			fmt.Println()
			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf(
				"Secrets encrypted and saved via %s (%d recipient(s))",
				session.Manifest.Project.Provider, recipientCount,
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
