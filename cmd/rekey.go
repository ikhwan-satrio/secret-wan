package cmd

import (
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var rekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Re-encrypt all secrets for the configured recipients",
	Long: `Re-encrypt the stored secret map for every key listed in the project's
recipient file. Use this after editing the recipient list manually or after
adding a teammate's key.

If no recipient file exists, the current key remains the only recipient.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		session := newSession()
		session.requireProject()
		session.requireExistingIdentity()

		result, err := session.Rekey(force)
		if err != nil {
			failApp("Failed to re-encrypt secrets", err)
		}

		if result.UsedImplicit {
			fmt.Println(ui.Muted.Render("No recipient list found; using the current key only."))
		}

		printHeader("rekey")
		if result.Missing {
			fmt.Println(ui.Muted.Render("No stored secrets to re-encrypt."))
			return
		}

		fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf(
			"Re-encrypted %d secret(s) for %d recipient(s).",
			result.Secrets, result.Recipients,
		)))

		if !result.SelfListed {
			warn("Your key is not a recipient; this key cannot decrypt the new payload.")
		}
	},
}

func init() {
	rekeyCmd.Flags().Bool("force", false, "Re-encrypt even when your key is not a recipient")
	rootCmd.AddCommand(rekeyCmd)
}
