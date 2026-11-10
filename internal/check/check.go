// Package check implements the strict consistency rules behind `secwan check
// --strict`, which report unmanaged secrets and invalid manifest rules.
package check

import (
	"fmt"
	"os"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/cli"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

// Run reports missing secrets for a profile and, when strict is set, manifest
// inconsistencies.
func Run(profileName string, strict bool) {
	session := cli.NewSession()
	session.RequireProject()
	profile := session.RequireProfile(profileName)

	expectedSecrets := profile.Keys()
	if len(expectedSecrets) == 0 {
		cli.Fail("Failed to load configuration", fmt.Sprintf("Profile '%s' must define at least one secret", profile.Name))
	}

	session.RequireExistingIdentity()
	actualSecrets := session.RequireSecrets()

	cli.PrintHeader("check")
	fmt.Println(ui.Muted.Render("Profile: ") + ui.Accent.Render(profile.Name))
	fmt.Println()

	allPresent := true
	for _, key := range expectedSecrets {
		keyColumn := fmt.Sprintf("%-24s", key)

		if value, exists := actualSecrets[key]; exists && strings.TrimSpace(value) != "" {
			fmt.Printf("  %s %s %s\n", ui.IconSuccess, ui.Key.Render(keyColumn), ui.Muted.Render("configured"))
		} else {
			fmt.Printf("  %s %s %s\n", ui.IconError, ui.Key.Render(keyColumn), ui.Error.Render("missing"))
			allPresent = false
		}
	}

	var issues []Issue
	if strict {
		issues = CollectStrictIssues(session.Manifest, actualSecrets)
		printStrictIssues(issues)
	}

	fmt.Println()
	fmt.Println(ui.Divider(44))

	if !allPresent {
		fmt.Printf("%s %s\n", ui.IconError, ui.Error.Render("Some secrets are missing."))
		fmt.Println(ui.Muted.Render(fmt.Sprintf("Run 'secwan sync -p %s' to configure them.", profileName)))
	}

	if strict && len(issues) > 0 {
		fmt.Printf("%s %s\n", ui.IconWarning,
			ui.Warning.Render(fmt.Sprintf("Strict check found %d issue(s).", len(issues))))
	}

	if !allPresent || (strict && len(issues) > 0) {
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", ui.IconSuccess, ui.Success.Render("All secrets are properly configured!"))
}
