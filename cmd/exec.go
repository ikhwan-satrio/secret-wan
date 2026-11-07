package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
			failErr("Failed to read profile flag", err)
		}

		session := newSession()
		session.requireIdentity()
		session.requireProject()
		targetProfile := session.requireProfile(profileName)

		secrets, err := session.LoadSecrets()
		if err != nil {
			failApp("Failed to decrypt secrets", err)
		}

		allowedKeys := make(map[string]bool)
		for _, key := range targetProfile.Keys() {
			allowedKeys[key] = true
		}

		filteredSecrets := make(map[string]string)
		for key, value := range secrets {
			if allowedKeys[key] {
				filteredSecrets[key] = value
			}
		}

		for key, value := range filteredSecrets {
			os.Setenv(key, value)
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
			failErr("Failed to execute command", err)
		}
	},
}

func init() {
	execCmd.Flags().StringP("profile", "p", "default", "Profile environment to use")
	rootCmd.AddCommand(execCmd)
}
