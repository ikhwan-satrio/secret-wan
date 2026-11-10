package cmd

import (
	"fmt"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
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
			cli.Fail("Invalid key", "Key must not be empty.")
		}

		profileName, _ := cmd.Flags().GetString("profile")
		generate, _ := cmd.Flags().GetBool("generate")
		force, _ := cmd.Flags().GetBool("force")

		if len(args) == 2 && generate {
			cli.Fail("Conflicting options", "Provide either VALUE or --generate, not both.")
		}

		session := cli.NewSession()
		session.RequireProject()
		profile := session.RequireProfile(profileName)
		session.RequireIdentity()

		if !profile.HasKey(key) {
			declaredIn := session.Manifest.ProfilesContaining(key)
			detail := fmt.Sprintf("%s is not declared in profile '%s'. Add it to %s.", key, profile.Name, session.Config.SpecFile)
			if len(declaredIn) > 0 {
				detail = fmt.Sprintf("%s is declared in profile(s) %s, not '%s'.", key, strings.Join(declaredIn, ", "), profile.Name)
			}
			if !force {
				cli.Fail("Key not declared", detail+" Use --force to store it anyway.")
			}

			cli.Warn(detail)
		}

		cli.PrintHeader("set")
		fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(profile.Name))
		fmt.Println()

		stored, missing := session.LoadSecretsAllowMissing()
		if missing {
			cli.Warn("No existing secrets found; a new storage will be created.")
		}

		value, err := resolveSecretValue(session, key, args, generate)
		if err != nil {
			cli.FailErr("Failed to read secret value", err)
		}
		if value == "" {
			cli.Fail("Empty secret value", "Use 'secwan unset "+key+"' to remove a secret instead.")
		}

		_, existed := stored[key]
		stored[key] = value

		session.SaveSecrets(stored, false)

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

func resolveSecretValue(session *cli.Session, key string, args []string, generate bool) (string, error) {
	if generate {
		return session.GenerateSecret()
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
