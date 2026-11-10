package recipient

import (
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <age1...>",
		Aliases: []string{"rm"},
		Short:   "Remove a recipient and re-encrypt existing secrets",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")

			session := cli.NewSession()

			change, err := session.RemoveRecipient(args[0], force)
			if err != nil {
				failRecipientChange(err)
			}

			cli.PrintHeader("recipient remove")
			fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("Removed "+change.Key))
			printRekeyResult(&app.RekeyResult{Secrets: change.Secrets, Recipients: change.Recipients})

			if change.SelfRemoved {
				cli.Warn("Your key was removed. New writes cannot be decrypted with it.")
			}
		},
	}

	cmd.Flags().Bool("force", false, "Allow removing your own key from the recipient list")

	return cmd
}
