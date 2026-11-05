package cmd

import (
	"fmt"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var recipientCmd = &cobra.Command{
	Use:     "recipient",
	Aliases: []string{"recipients"},
	Short:   "Manage age recipients that can decrypt project secrets",
}

var recipientListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List configured age recipients",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		appConfig := config.UseApp()
		store := recipients.New(appConfig.RecipientsPath())

		keys, err := store.Load()
		printHeader("recipient list")
		if err != nil {
			failErr("Failed to read recipients", err)
		}

		cryptoModule := crypto.CryptoModule{AppConfig: appConfig}
		identity, keyErr := cryptoModule.LoadKey()
		self := ""
		if keyErr == nil {
			self = identity.Recipient().String()
		}

		fmt.Println(ui.Muted.Render("File: ") + ui.Accent.Render(appConfig.RecipientsPath()))
		fmt.Println()

		if len(keys) == 0 {
			fmt.Println(ui.Muted.Render("No recipient list yet; the project uses single-recipient mode."))
			if self != "" {
				fmt.Printf("  %s %s\n", ui.Key.Render(self), ui.Muted.Render("(you, implicit)"))
			} else {
				fmt.Println(ui.Muted.Render("No master key found. Run 'secwan init' first."))
			}
			return
		}

		for _, key := range keys {
			if key == self {
				fmt.Printf("  %s %s\n", ui.Key.Render(key), ui.Success.Render("(you)"))
				continue
			}
			fmt.Printf("  %s\n", ui.Key.Render(key))
		}

		fmt.Println()
		fmt.Println(ui.Muted.Render(fmt.Sprintf("%d recipient(s) configured.", len(keys))))

		if self != "" && !recipients.Contains(keys, self) {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning,
				ui.Warning.Render("Your key is not in this list. Future writes would not be decryptable by you."))
		}
	},
}

var recipientMeCmd = &cobra.Command{
	Use:     "me",
	Aliases: []string{"whoami"},
	Short:   "Print your public age recipient",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		appConfig := config.UseApp()
		cryptoModule := crypto.CryptoModule{AppConfig: appConfig}

		identity, err := cryptoModule.LoadKey()
		if err != nil {
			if os.IsNotExist(err) {
				fail("Master key not found", "Run 'secwan init' first.")
			}
			failErr("Failed to load master key", err)
		}

		self := identity.Recipient().String()
		fmt.Println(self)

		keys, loadErr := recipients.New(appConfig.RecipientsPath()).Load()
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning,
				ui.Warning.Render(fmt.Sprintf("Could not read recipients: %v", loadErr)))
			return
		}

		if len(keys) > 0 && !recipients.Contains(keys, self) {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning,
				ui.Warning.Render("Your key is not in the project's recipient list."))
		}
	},
}

var recipientAddCmd = &cobra.Command{
	Use:   "add <age1...>",
	Short: "Add a recipient and re-encrypt existing secrets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		newKey, err := normalizeRecipient(args[0])
		if err != nil {
			fail("Invalid recipient", err.Error())
		}

		session := newSession()
		session.requireManifest()
		session.requireIdentity()

		store := session.recipientsStore()
		keys, err := store.Load()
		if err != nil {
			failErr("Failed to read recipients", err)
		}

		self := session.Identity.Recipient().String()
		if len(keys) == 0 {
			keys = []string{self}
		}

		// Adding your own key is allowed when the list exists but does not
		// include it; this is the recovery path suggested by other commands.
		if !recipients.Contains(keys, self) && newKey != self {
			fail("Current key is not a recipient",
				fmt.Sprintf("Add your own key first: secwan recipient add %s", self))
		}

		printHeader("recipient add")
		if recipients.Contains(keys, newKey) {
			fmt.Printf("%s %s\n", ui.IconInfo, ui.Muted.Render("Recipient is already configured."))
			return
		}

		oldKeys := append([]string(nil), keys...)
		newKeys := append(append([]string(nil), keys...), newKey)
		if err := store.Save(newKeys); err != nil {
			failErr("Failed to save recipients", err)
		}

		fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("Added "+newKey))

		count, err := session.rekey(false)
		if err != nil {
			if rollbackErr := store.Save(oldKeys); rollbackErr != nil {
				failErr("Failed to re-encrypt secrets and could not restore recipients", err)
			}
			failErr("Failed to re-encrypt secrets; recipient list was restored", err)
		}

		printRekeyResult(count, len(newKeys))
	},
}

var recipientRemoveCmd = &cobra.Command{
	Use:     "remove <age1...>",
	Aliases: []string{"rm"},
	Short:   "Remove a recipient and re-encrypt existing secrets",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, err := normalizeRecipient(args[0])
		if err != nil {
			fail("Invalid recipient", err.Error())
		}

		force, _ := cmd.Flags().GetBool("force")

		session := newSession()
		session.requireManifest()
		session.requireExistingIdentity()

		store := session.recipientsStore()
		keys, err := store.Load()
		if err != nil {
			failErr("Failed to read recipients", err)
		}

		self := session.Identity.Recipient().String()
		if len(keys) == 0 {
			keys = []string{self}
		}

		if !recipients.Contains(keys, target) {
			fail("Recipient not found", fmt.Sprintf("%s is not in %s.", target, session.Config.RecipientsPath()))
		}
		if len(keys) == 1 {
			fail("Cannot remove the last recipient", "At least one recipient must remain.")
		}
		if !recipients.Contains(keys, self) && !force {
			fail("Current key is not a recipient", "Use --force only if you intend to lose access to future writes.")
		}

		selfRemoved := target == self
		if selfRemoved && !force {
			fail("Refusing to remove your own key",
				"Use --force to re-encrypt for the remaining recipients. This key will no longer decrypt new writes.")
		}

		oldKeys := append([]string(nil), keys...)
		newKeys := recipients.Remove(keys, target)
		if err := store.Save(newKeys); err != nil {
			failErr("Failed to save recipients", err)
		}

		printHeader("recipient remove")
		fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("Removed "+target))

		count, err := session.rekey(true)
		if err != nil {
			if rollbackErr := store.Save(oldKeys); rollbackErr != nil {
				failErr("Failed to re-encrypt secrets and could not restore recipients", err)
			}
			failErr("Failed to re-encrypt secrets; recipient list was restored", err)
		}

		printRekeyResult(count, len(newKeys))

		if selfRemoved {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning,
				ui.Warning.Render("Your key was removed. New writes cannot be decrypted with it."))
		}
	},
}

func normalizeRecipient(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("recipient must not be empty")
	}

	recipient, err := age.ParseX25519Recipient(raw)
	if err != nil {
		return "", err
	}

	return recipient.String(), nil
}

func printRekeyResult(count, recipientCount int) {
	if count == 0 {
		fmt.Println(ui.Muted.Render("No stored secrets to re-encrypt yet."))
		fmt.Println(ui.Muted.Render("Future saves will be encrypted to all configured recipients."))
		return
	}

	fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render(fmt.Sprintf(
		"Re-encrypted %d secret(s) for %d recipient(s).",
		count, recipientCount,
	)))
}

func init() {
	recipientRemoveCmd.Flags().Bool("force", false, "Allow removing your own key from the recipient list")

	recipientCmd.AddCommand(recipientListCmd)
	recipientCmd.AddCommand(recipientMeCmd)
	recipientCmd.AddCommand(recipientAddCmd)
	recipientCmd.AddCommand(recipientRemoveCmd)
	rootCmd.AddCommand(recipientCmd)
}
