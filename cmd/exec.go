package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [command] [args...]",
	Short: "Run a command with a specific secret profile injected into the environment",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			fmt.Println(ui.FormatError("Failed to read profile flag", err.Error()))
			os.Exit(1)
		}

		appConfig := config.UseApp()
		cryptoModule := crypto.CryptoModule{AppConfig: appConfig}
		parserModule := manifest.ParserModule{AppConfig: appConfig}

		identity, err := cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(ui.FormatError("Failed to load master key", err.Error()))
			os.Exit(1)
		}

		parsedManifest, err := parserModule.ParseManifest()
		if err != nil {
			fmt.Println(ui.FormatError("Failed to parse manifest", err.Error()))
			os.Exit(1)
		}

		targetProfile, exists := parsedManifest.Profiles[profileName]
		if !exists {
			validProfiles := manifest.GetProfileNames(parsedManifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' not found. Available profiles: %v", profileName, validProfiles)
			fmt.Println(ui.FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		storage := provider.GetStorageProvider(appConfig, parsedManifest.Project.Provider, parsedManifest.Project.Name)

		secrets, err := cryptoModule.LoadSecrets(storage, identity)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.FormatError("Secrets storage not found",
					fmt.Sprintf("Run 'secwan sync -p %s' first.", profileName)))
			} else {
				fmt.Println(ui.FormatError("Failed to decrypt secrets", err.Error()))
			}
			os.Exit(1)
		}

		allowedKeys := make(map[string]bool)
		for _, k := range targetProfile.Required {
			allowedKeys[k] = true
		}
		for _, k := range targetProfile.AutoGen {
			allowedKeys[k] = true
		}

		filteredSecrets := make(map[string]string)
		for k, v := range secrets {
			if allowedKeys[k] {
				filteredSecrets[k] = v
			}
		}

		for k, v := range filteredSecrets {
			os.Setenv(k, v)
		}

		targetCmd := args[0]
		cmdArgs := args[1:]
		commandLine := strings.Join(append([]string{targetCmd}, cmdArgs...), " ")

		fmt.Printf("%s %s\n\n", ui.Muted.Render("Running:"), ui.Text.Render(commandLine))

		c := exec.Command(targetCmd, cmdArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = os.Environ()

		err = c.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				os.Exit(exitError.ExitCode())
			}
			fmt.Println(ui.FormatError("Failed to execute command", err.Error()))
			os.Exit(1)
		}
	},
}

func init() {
	execCmd.Flags().StringP("profile", "p", "default", "Profile environment to use")
	rootCmd.AddCommand(execCmd)
}
