package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

type buildInfo struct {
	Version string
	Commit  string
	Date    string
}

var currentBuild = buildInfo{
	Version: "dev",
	Commit:  "none",
	Date:    "unknown",
}

// SetVersionInfo is called from main with values injected at build time via
// -ldflags "-X main.version=... -X main.commit=... -X main.date=...".
func SetVersionInfo(version, commit, date string) {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "none"
	}
	if date == "" {
		date = "unknown"
	}

	currentBuild = buildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	rootCmd.Version = displayVersion()
	rootCmd.SetVersionTemplate(versionTemplate())
}

func displayVersion() string {
	if currentBuild.Version == "dev" || len(currentBuild.Version) == 0 {
		return currentBuild.Version
	}

	if currentBuild.Version[0] == 'v' {
		return currentBuild.Version
	}

	return "v" + currentBuild.Version
}

func versionTemplate() string {
	return fmt.Sprintf(
		"secret-wan %s\n"+
			"  commit:   %s\n"+
			"  built:    %s\n"+
			"  go:       %s\n"+
			"  platform: %s/%s\n",
		displayVersion(),
		currentBuild.Commit,
		currentBuild.Date,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(versionTemplate())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
