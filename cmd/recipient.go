package cmd

import (
	"errors"
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var recipientCmd = &cobra.Command{
	Use:     "recipient",
	Aliases: []string{"recipients"},
	Short:   "Manage age recipients that can decrypt project secrets",
}

// printRekeyResult reports the outcome shared by recipient add/remove and the
// standalone rekey command.
func printRekeyResult(result *app.RekeyResult) {
	if result.Secrets == 0 {
		fmt.Println(ui.Muted.Render("No stored secrets to re-encrypt yet."))
		fmt.Println(ui.Muted.Render("Future saves will be encrypted to all configured recipients."))
		return
	}

	fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf(
		"Re-encrypted %d secret(s) for %d recipient(s).",
		result.Secrets, result.Recipients,
	)))
}

// failRecipientChange renders the detailed errors produced when a recipient
// change and its rekey could not be kept consistent.
func failRecipientChange(err error) {
	var changeErr *app.RecipientChangeError
	if errors.As(err, &changeErr) {
		if changeErr.Restored {
			failErr("Failed to re-encrypt secrets; recipient list was restored", changeErr.Err)
		}

		detail := changeErr.Err.Error()
		if changeErr.RollbackErr != nil {
			detail += "; rollback failed: " + changeErr.RollbackErr.Error()
		}
		fail("Failed to re-encrypt secrets and could not restore recipients", detail)
	}

	failApp("Failed to update recipients", err)
}

func init() {
	recipientRemoveCmd.Flags().Bool("force", false, "Allow removing your own key from the recipient list")

	recipientCmd.AddCommand(recipientListCmd)
	recipientCmd.AddCommand(recipientMeCmd)
	recipientCmd.AddCommand(recipientAddCmd)
	recipientCmd.AddCommand(recipientRemoveCmd)
	rootCmd.AddCommand(recipientCmd)
}
