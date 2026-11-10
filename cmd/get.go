package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <KEY>",
	Short: "Print a single secret (masked by default)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		profileName, _ := cmd.Flags().GetString("profile")
		reveal, _ := cmd.Flags().GetBool("reveal")

		session := cli.NewSession()
		session.RequireProject()
		profile := session.RequireProfile(profileName)
		session.RequireExistingIdentity()
		stored := session.RequireSecrets()

		value, exists := stored[key]
		if !exists {
			cli.Fail("Secret not found", fmt.Sprintf("%s is not stored. Run 'secwan set %s' to create it.", key, key))
		}

		if !profile.HasKey(key) {
			detail := fmt.Sprintf("%s is not declared in profile '%s'.", key, profile.Name)
			if others := session.Manifest.ProfilesContaining(key); len(others) > 0 {
				detail = fmt.Sprintf("%s is declared in profile(s) %s, not '%s'.", key, strings.Join(others, ", "), profile.Name)
			}
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning, ui.Warning.Render(detail))
		}

		if reveal {
			fmt.Println(value)
			return
		}

		fmt.Println(ui.Muted.Render(maskedValue))
	},
}

func init() {
	getCmd.Flags().StringP("profile", "p", "default", "Profile the key belongs to")
	getCmd.Flags().Bool("reveal", false, "Show the secret value in plain text")
	rootCmd.AddCommand(getCmd)
}
