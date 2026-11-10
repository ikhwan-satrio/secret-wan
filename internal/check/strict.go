package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/ui"
)

// Issue is one strict-check finding.
type Issue struct {
	Label  string
	Detail string
}

// CollectStrictIssues reports manifest inconsistencies and stored keys that no
// profile declares.
func CollectStrictIssues(parsed *manifest.Manifest, stored map[string]string) []Issue {
	var issues []Issue

	if strings.TrimSpace(parsed.Project.Name) == "" {
		issues = append(issues, Issue{"project.name", "project name is empty"})
	}

	if !provider.IsKnown(parsed.Project.Provider) {
		issues = append(issues, Issue{
			"project.provider",
			fmt.Sprintf("unknown provider '%s' (expected %s)", parsed.Project.Provider, strings.Join(provider.Names(), ", ")),
		})
	}

	declared := parsed.DeclaredKeys()

	var extra []string
	for key := range stored {
		if !declared[key] {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	for _, key := range extra {
		issues = append(issues, Issue{key, "stored but not declared in any profile"})
	}

	for _, name := range manifest.GetProfileNames(parsed.Profiles) {
		profile := parsed.Profiles[name]

		if len(profile.Keys()) == 0 {
			issues = append(issues, Issue{name, "profile has no secrets defined"})
		}

		for _, key := range profile.Declared {
			rule := profile.Rule(key)
			switch {
			case !rule.Required && !rule.Generate:
				issues = append(issues, Issue{key, fmt.Sprintf("profile '%s' has no rule (use required or generate)", name)})
			case rule.Required && rule.Generate:
				issues = append(issues, Issue{key, fmt.Sprintf("profile '%s' marks the key as both required and generate", name)})
			}
		}
	}

	return issues
}

func printStrictIssues(issues []Issue) {
	fmt.Println()
	fmt.Println(ui.Bold.Render("Strict checks"))

	if len(issues) == 0 {
		fmt.Printf("  %s %s\n", ui.IconSuccess, ui.Success.Render("No inconsistencies found."))
		return
	}

	for _, issue := range issues {
		fmt.Printf("  %s %s %s\n",
			ui.IconWarning,
			ui.Key.Render(fmt.Sprintf("%-24s", issue.Label)),
			ui.Warning.Render(issue.Detail),
		)
	}
}
