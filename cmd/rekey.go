package cmd

import (
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
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
		session.requireManifest()
		session.requireExistingIdentity()

		keys, err := session.recipientKeys()
		if err != nil {
			failErr("Failed to read recipients", err)
		}

		self := session.Identity.Recipient().String()
		if len(keys) == 0 {
			keys = []string{self}
			fmt.Println(ui.Muted.Render("No recipient list found; using the current key only."))
		}

		if !recipients.Contains(keys, self) && !force {
			fail("Current key is not a recipient",
				fmt.Sprintf("Add it with 'secwan recipient add %s' or use --force to lock this key out.", self))
		}

		secrets, err := session.Crypto.LoadSecrets(session.Storage, session.Identity)
		if err != nil {
			if provider.IsNotFound(err) {
				printHeader("rekey")
				fmt.Println(ui.Muted.Render("No stored secrets to re-encrypt."))
				return
			}
			failErr("Failed to decrypt secrets", err)
		}

		ageRecipients, err := recipients.Parse(keys)
		if err != nil {
			failErr("Invalid recipients", err)
		}

		if err := session.Crypto.SaveSecrets(session.Storage, secrets, ageRecipients...); err != nil {
			failErr("Failed to save re-encrypted secrets", err)
		}

		printHeader("rekey")
		fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf(
			"Re-encrypted %d secret(s) for %d recipient(s).",
			len(secrets), len(ageRecipients),
		)))

		if !recipients.Contains(keys, self) {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning,
				ui.Warning.Render("Your key is not a recipient; this key cannot decrypt the new payload."))
		}
	},
}

func init() {
	rekeyCmd.Flags().Bool("force", false, "Re-encrypt even when your key is not a recipient")
	rootCmd.AddCommand(rekeyCmd)
}
