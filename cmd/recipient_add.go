package cmd

import (
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var recipientAddCmd = &cobra.Command{
	Use:   "add <age1...>",
	Short: "Add a recipient and re-encrypt existing secrets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		session := newSession()

		change, err := session.AddRecipient(args[0])
		if err != nil {
			failRecipientChange(err)
		}

		printHeader("recipient add")
		if change.AlreadyPresent {
			fmt.Printf("%s %s\n", ui.IconInfo, ui.Muted.Render("Recipient is already configured."))
			return
		}

		fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("Added "+change.Key))
		printRekeyResult(&app.RekeyResult{Secrets: change.Secrets, Recipients: change.Recipients})
	},
}
