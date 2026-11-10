package cmd

import (
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/recipient"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "secret-wan",
	Short: "Secure secret manager for your projects",
	Long: `secret-wan defines your secrets in a TOML manifest, stores them encrypted
in a local file or your OS keychain, and injects them into commands at runtime
without leaking them into your shell history.`,
	Example: `  secwan init
  secwan sync
  secwan list
  secwan set API_KEY
  secwan check --strict
  secwan doctor
  secwan exec -- npm start`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(recipient.NewCommand())
}
