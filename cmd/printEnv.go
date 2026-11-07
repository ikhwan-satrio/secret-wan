package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var printEnvCmd = &cobra.Command{
	Use:   "print-env",
	Short: "Print decrypted secrets to stdout in KEY=VALUE format",
	Long: `Print decrypted secrets for a profile to stdout.

All headers and warnings are written to stderr, so the output can safely be
used in scripts:

  eval "$(secwan print-env)"

Use --json for a structured JSON object instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		profileName, _ := cmd.Flags().GetString("profile")
		asJSON, _ := cmd.Flags().GetBool("json")

		session := newSession()
		session.requireProject()
		profile := session.requireProfile(profileName)
		session.requireExistingIdentity()
		stored := session.requireSecrets()

		active := make(map[string]string)
		for _, key := range profile.Keys() {
			if value, exists := stored[key]; exists {
				active[key] = value
			}
		}

		keys := make([]string, 0, len(active))
		for key := range active {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		if !asJSON {
			fmt.Fprintf(os.Stderr, "%s %s\n",
				ui.Muted.Render("# Secrets for profile:"),
				ui.Accent.Render(profile.Name),
			)
			fmt.Fprintln(os.Stderr, ui.Warning.Render("# WARNING: This output contains sensitive data. Do not share screenshots."))
		}

		if len(keys) == 0 {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning,
				ui.Warning.Render(fmt.Sprintf("No secrets configured for profile '%s'.", profile.Name)))
			if asJSON {
				fmt.Println("{}")
			}
			return
		}

		if asJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(active); err != nil {
				failErr("Failed to encode JSON", err)
			}
			return
		}

		for _, key := range keys {
			fmt.Printf("%s=%s\n", key, quoteEnvValue(active[key]))
		}
	},
}

// quoteEnvValue renders a value as a double-quoted environment assignment.
// Shell-sensitive characters are escaped so the result is safe to eval.
func quoteEnvValue(value string) string {
	escaped := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`$`, `\$`,
		"`", "\\`",
	).Replace(value)

	return `"` + escaped + `"`
}

func init() {
	printEnvCmd.Flags().StringP("profile", "p", "default", "Profile environment to display")
	printEnvCmd.Flags().Bool("json", false, "Print secrets as a JSON object")
	rootCmd.AddCommand(printEnvCmd)
}
