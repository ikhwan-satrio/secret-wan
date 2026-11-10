package cmd

import (
	"fmt"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var unsetCmd = &cobra.Command{
	Use:     "unset <KEY>",
	Aliases: []string{"remove", "rm"},
	Short:   "Remove a single secret from storage",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.TrimSpace(args[0])
		if key == "" {
			cli.Fail("Invalid key", "Key must not be empty.")
		}

		session := cli.NewSession()
		session.RequireProject()
		session.RequireExistingIdentity()
		stored := session.RequireSecrets()

		if _, exists := stored[key]; !exists {
			cli.Fail("Secret not found", fmt.Sprintf("%s is not stored.", key))
		}

		cli.PrintHeader("unset")

		declaredIn := session.Manifest.ProfilesContaining(key)
		if len(declaredIn) > 0 {
			cli.Warn(fmt.Sprintf("%s is still declared in profile(s) %s.", key, strings.Join(declaredIn, ", ")))
		}

		delete(stored, key)

		session.SaveSecrets(stored, false)

		fmt.Printf("%s %s %s\n", ui.IconSuccess, ui.Success.Render("removed"), ui.Key.Render(key))
	},
}

func init() {
	rootCmd.AddCommand(unsetCmd)
}
