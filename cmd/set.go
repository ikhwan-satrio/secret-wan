package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/secrets"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <KEY> [VALUE]",
	Short: "Create or update a single secret",
	Long: `Create or update a secret without opening the interactive sync screen.

When VALUE is omitted, the value is read from the terminal without echoing it.
Use --generate to create a cryptographically secure random value instead.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.TrimSpace(args[0])
		if key == "" {
			fail("Invalid key", "Key must not be empty.")
		}

		profileName, _ := cmd.Flags().GetString("profile")
		generate, _ := cmd.Flags().GetBool("generate")
		force, _ := cmd.Flags().GetBool("force")

		if len(args) == 2 && generate {
			fail("Conflicting options", "Provide either VALUE or --generate, not both.")
		}

		session := newSession()
		session.requireManifest()
		profile := session.requireProfile(profileName)
		session.requireIdentity()

		if !keyInProfile(profile, key) {
			declaredIn := profilesContaining(session.Manifest, key)
			detail := fmt.Sprintf("%s is not declared in profile '%s'. Add it to %s.", key, profile.Name, session.Config.SpecFile)
			if len(declaredIn) > 0 {
				detail = fmt.Sprintf("%s is declared in profile(s) %s, not '%s'.", key, strings.Join(declaredIn, ", "), profile.Name)
			}
			if !force {
				fail("Key not declared", detail+" Use --force to store it anyway.")
			}

			fmt.Fprintf(os.Stderr, "%s %s\n", ui.IconWarning, ui.Warning.Render(detail))
		}

		printHeader("set")
		fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(profile.Name))
		fmt.Println()

		stored := session.loadSecretsAllowMissing()

		value, err := resolveSecretValue(session, key, args, generate)
		if err != nil {
			failErr("Failed to read secret value", err)
		}
		if value == "" {
			fail("Empty secret value", "Use 'secwan unset "+key+"' to remove a secret instead.")
		}

		_, existed := stored[key]
		stored[key] = value

		if err := session.saveSecrets(stored, false); err != nil {
			failErr("Failed to save secrets", err)
		}

		action := "saved"
		if existed {
			action = "updated"
		}

		fmt.Printf("%s %s %s\n", ui.IconSuccess, ui.Success.Render(action), ui.Key.Render(key))

		if generate {
			fmt.Println(ui.Muted.Render(fmt.Sprintf("Run 'secwan get %s --reveal' to view the generated value.", key)))
		}
	},
}

func resolveSecretValue(session *appSession, key string, args []string, generate bool) (string, error) {
	if generate {
		return session.Crypto.GenerateSecurePassword(session.Config.SecretLength)
	}

	if len(args) == 2 {
		return args[1], nil
	}

	module := secrets.SecretModule{AppConfig: session.Config}
	return module.PromptUserInput(key, true)
}

func init() {
	setCmd.Flags().StringP("profile", "p", "default", "Profile the key belongs to")
	setCmd.Flags().Bool("generate", false, "Generate a secure random value")
	setCmd.Flags().Bool("force", false, "Store the key even if it is not declared in the profile")
	rootCmd.AddCommand(setCmd)
}
