package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/config"
	"github.com/ikhwan-satrio/secret-wan/utils"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [command] [args...]",
	Short: "Jalankan perintah dengan secret profile tertentu yang di-inject ke environment",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName, err := cmd.Flags().GetString("profile")
		if err != nil {
			fmt.Println(FormatError("Failed to read profile flag", err.Error()))
			os.Exit(1)
		}

		appConfig := config.UseApp()
		cryptoModule := utils.CryptoModule{AppConfig: appConfig}
		parserModule := utils.ParserModule{AppConfig: appConfig}

		identity, err := cryptoModule.LoadOrGenerateKey()
		if err != nil {
			fmt.Println(FormatError("Failed to load master key", err.Error()))
			os.Exit(1)
		}

		manifest, err := parserModule.ParseManifest()
		if err != nil {
			fmt.Println(FormatError("Failed to parse manifest", err.Error()))
			os.Exit(1)
		}

		targetProfile, exists := manifest.Profiles[profileName]
		if !exists {
			validProfiles := getProfileNames(manifest.Profiles)
			errMsg := fmt.Sprintf("Profile '%s' tidak ditemukan. Profile yang tersedia: %v", profileName, validProfiles)
			fmt.Println(FormatError("Invalid profile", errMsg))
			os.Exit(1)
		}

		// PERBAIKAN: Dapatkan provider terlebih dahulu
		provider := utils.GetStorageProvider(appConfig, manifest.Project.Provider, manifest.Project.Name)
		
		// PERBAIKAN: Panggil LoadSecrets dengan 2 argumen (provider, identity)
		secrets, err := cryptoModule.LoadSecrets(provider, identity)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(FormatError("Secrets storage not found",
					fmt.Sprintf("Jalankan 'secman sync -p %s' terlebih dahulu.", profileName)))
			} else {
				fmt.Println(FormatError("Failed to decrypt secrets", err.Error()))
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

		fmt.Printf("%s %s\n\n", StyleDim.Render("Running:"), strings.Join(append([]string{targetCmd}, cmdArgs...), " "))

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
			fmt.Println(FormatError("Failed to execute command", err.Error()))
			os.Exit(1)
		}
	},
}

func init() {
	execCmd.Flags().StringP("profile", "p", "default", "Profile environment yang ingin digunakan")
	rootCmd.AddCommand(execCmd)
}
