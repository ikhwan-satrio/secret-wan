package cmd

import (
	"fmt"

	"github.com/ikhwan-satrio/secret-wan/internal/buildinfo"
	"github.com/spf13/cobra"
)

// SetVersionInfo is called from main with values injected at build time via
// -ldflags "-X main.version=... -X main.commit=... -X main.date=...".
func SetVersionInfo(version, commit, date string) {
	buildinfo.Set(version, commit, date)

	rootCmd.Version = buildinfo.Display()
	rootCmd.SetVersionTemplate(buildinfo.Template())
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(buildinfo.Template())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
