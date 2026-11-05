package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
	"github.com/spf13/cobra"
)

type strictIssue struct {
	label  string
	detail string
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check secret completeness based on config",
	Run: func(cmd *cobra.Command, args []string) {
		profileName, _ := cmd.Flags().GetString("profile")
		strict, _ := cmd.Flags().GetBool("strict")

		session := newSession()
		session.requireManifest()
		profile := session.requireProfile(profileName)

		expectedSecrets := profileKeys(profile)
		if len(expectedSecrets) == 0 {
			fail("Failed to load configuration", fmt.Sprintf("Profile '%s' must define at least one secret", profile.Name))
		}

		session.requireExistingIdentity()
		actualSecrets := session.requireSecrets()

		printHeader("check")
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

		var issues []strictIssue
		if strict {
			issues = collectStrictIssues(session.Manifest, actualSecrets)
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
	},
}

func collectStrictIssues(parsed *manifest.Manifest, stored map[string]string) []strictIssue {
	var issues []strictIssue

	if strings.TrimSpace(parsed.Project.Name) == "" {
		issues = append(issues, strictIssue{"project.name", "project name is empty"})
	}

	switch parsed.Project.Provider {
	case "", "dotfiles", "keyring", "docker":
	default:
		issues = append(issues, strictIssue{
			"project.provider",
			fmt.Sprintf("unknown provider '%s' (expected dotfiles, keyring, or docker)", parsed.Project.Provider),
		})
	}

	declared := allDeclaredKeys(parsed)

	var extra []string
	for key := range stored {
		if !declared[key] {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	for _, key := range extra {
		issues = append(issues, strictIssue{key, "stored but not declared in any profile"})
	}

	for _, name := range manifest.GetProfileNames(parsed.Profiles) {
		profile := parsed.Profiles[name]

		if len(profileKeys(profile)) == 0 {
			issues = append(issues, strictIssue{name, "profile has no secrets defined"})
		}

		for _, key := range profile.Declared {
			rule := ruleFor(profile, key)
			switch {
			case !rule.Required && !rule.Generate:
				issues = append(issues, strictIssue{key, fmt.Sprintf("profile '%s' has no rule (use required or generate)", name)})
			case rule.Required && rule.Generate:
				issues = append(issues, strictIssue{key, fmt.Sprintf("profile '%s' marks the key as both required and generate", name)})
			}
		}
	}

	return issues
}

func printStrictIssues(issues []strictIssue) {
	fmt.Println()
	fmt.Println(ui.Bold.Render("Strict checks"))

	if len(issues) == 0 {
		fmt.Printf("  %s %s\n", ui.IconSuccess, ui.Success.Render("No inconsistencies found."))
		return
	}

	for _, issue := range issues {
		fmt.Printf("  %s %s %s\n",
			ui.IconWarning,
			ui.Key.Render(fmt.Sprintf("%-24s", issue.label)),
			ui.Warning.Render(issue.detail),
		)
	}
}

func init() {
	checkCmd.Flags().StringP("profile", "p", "default", "Profile to check")
	checkCmd.Flags().Bool("strict", false, "Also report unmanaged secrets and invalid manifest rules")
	rootCmd.AddCommand(checkCmd)
}
