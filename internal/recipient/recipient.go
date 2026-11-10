// Package recipient implements the recipient command family: listing the
// project's age recipients, identifying the current key, and adding or removing
// recipients with an automatic rekey.
package recipient

import (
	"errors"
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

// NewCommand builds the `recipient` command family.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recipient",
		Aliases: []string{"recipients"},
		Short:   "Manage age recipients that can decrypt project secrets",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newMeCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newRemoveCommand())

	return cmd
}

// printRekeyResult reports the outcome shared by recipient add/remove.
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
			cli.FailErr("Failed to re-encrypt secrets; recipient list was restored", changeErr.Err)
		}

		detail := changeErr.Err.Error()
		if changeErr.RollbackErr != nil {
			detail += "; rollback failed: " + changeErr.RollbackErr.Error()
		}
		cli.Fail("Failed to re-encrypt secrets and could not restore recipients", detail)
	}

	cli.FailApp("Failed to update recipients", err)
}
