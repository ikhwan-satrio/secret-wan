package cmd

import (
	"fmt"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

const maskedValue = "••••••••"

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List secrets for a profile (values masked by default)",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		profileName, _ := cmd.Flags().GetString("profile")
		reveal, _ := cmd.Flags().GetBool("reveal")
		all, _ := cmd.Flags().GetBool("all")

		session := cli.NewSession()
		session.RequireProject()
		profile := session.RequireProfile(profileName)
		session.RequireExistingIdentity()
		stored := session.RequireSecrets()

		keys := profile.Keys()
		if all {
			keys = profile.AllKeys(stored)
		}

		cli.PrintHeader("list")
		fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(profile.Name))
		if all {
			fmt.Println(ui.Muted.Render("Showing all stored secrets, including unmanaged keys."))
		}
		fmt.Println()

		if len(keys) == 0 {
			fmt.Println(ui.Muted.Render("No secrets defined for this profile."))
			return
		}

		fmt.Printf("  %s %s %s\n",
			ui.Muted.Render(fmt.Sprintf("%-28s", "KEY")),
			ui.Muted.Render(fmt.Sprintf("%-12s", "TYPE")),
			ui.Muted.Render("VALUE"),
		)

		for _, key := range keys {
			value, storedExists := stored[key]
			keyColumn := fmt.Sprintf("%-28s", key)
			typeColumn := fmt.Sprintf("%-12s", secretType(session.Manifest, profile, key))

			var valueColumn string
			switch {
			case !storedExists:
				valueColumn = ui.Warning.Render("not set")
			case reveal:
				valueColumn = strings.ReplaceAll(value, "\n", `\n`)
			default:
				valueColumn = ui.Muted.Render(maskedValue)
			}

			fmt.Printf("  %s %s %s\n", ui.Key.Render(keyColumn), ui.Muted.Render(typeColumn), valueColumn)
		}
	},
}

func secretType(parsed *manifest.Manifest, profile *manifest.Profile, key string) string {
	if !profile.HasKey(key) {
		if len(parsed.ProfilesContaining(key)) > 0 {
			return "other"
		}
		return "unmanaged"
	}

	rule := profile.Rule(key)
	switch {
	case rule.Required && rule.Generate:
		return "gen+req"
	case rule.Generate:
		return "generated"
	case rule.Required:
		return "required"
	default:
		return "declared"
	}
}

func init() {
	listCmd.Flags().StringP("profile", "p", "default", "Profile to list")
	listCmd.Flags().Bool("reveal", false, "Show secret values in plain text")
	listCmd.Flags().Bool("all", false, "Include stored keys that are not declared in the profile")
	rootCmd.AddCommand(listCmd)
}
