// Package buildinfo holds the version metadata injected at build time and the
// rendering used by both the version command and diagnostics.
package buildinfo

import (
	"fmt"
	"runtime"
)

// Info describes the build metadata for this binary.
type Info struct {
	Version string
	Commit  string
	Date    string
}

var current = Info{
	Version: "dev",
	Commit:  "none",
	Date:    "unknown",
}

// Set replaces the build metadata, filling defaults for empty values.
func Set(version, commit, date string) {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "none"
	}
	if date == "" {
		date = "unknown"
	}

	current = Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

// Current returns the active build metadata.
func Current() Info {
	return current
}

// Display returns the version with a leading "v" when it is a release version.
func Display() string {
	if current.Version == "dev" || len(current.Version) == 0 {
		return current.Version
	}

	if current.Version[0] == 'v' {
		return current.Version
	}

	return "v" + current.Version
}

// Template renders the multi-line version block used by `version` and
// `--version`.
func Template() string {
	return fmt.Sprintf(
		"secret-wan %s\n"+
			"  commit:   %s\n"+
			"  built:    %s\n"+
			"  go:       %s\n"+
			"  platform: %s/%s\n",
		Display(),
		current.Commit,
		current.Date,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
