package check

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
)

func TestCollectStrictIssues(t *testing.T) {
	parsed := &manifest.Manifest{
		Project: manifest.ProjectInfo{Name: "demo", Provider: "dotfiles"},
		Profiles: map[string]*manifest.Profile{
			"default": {
				Name:     "default",
				Required: []string{"A"},
				Declared: []string{"A", "B"},
				Rules: map[string]manifest.SecretRule{
					"A": {Required: true},
					"B": {},
				},
			},
			"empty": {
				Name:  "empty",
				Rules: map[string]manifest.SecretRule{},
			},
		},
	}

	issues := CollectStrictIssues(parsed, map[string]string{
		"A":     "value",
		"EXTRA": "value",
	})

	if len(issues) != 3 {
		t.Fatalf("CollectStrictIssues() returned %d issues, want 3: %#v", len(issues), issues)
	}

	if !hasIssue(issues, "EXTRA", "stored but not declared in any profile") {
		t.Error("missing issue for unmanaged key EXTRA")
	}
	if !hasIssue(issues, "empty", "profile has no secrets defined") {
		t.Error("missing issue for empty profile")
	}
	if !hasIssue(issues, "B", "profile 'default' has no rule (use required or generate)") {
		t.Error("missing issue for rule without action")
	}
}

func TestCollectStrictIssuesFlagsUnknownProvider(t *testing.T) {
	parsed := &manifest.Manifest{
		Project: manifest.ProjectInfo{Name: "demo", Provider: "vault"},
		Profiles: map[string]*manifest.Profile{
			"default": {Name: "default", Required: []string{"A"}, Declared: []string{"A"}},
		},
	}

	issues := CollectStrictIssues(parsed, map[string]string{"A": "value"})
	expected := fmt.Sprintf("unknown provider 'vault' (expected %s)", strings.Join(provider.Names(), ", "))
	if !hasIssue(issues, "project.provider", expected) {
		t.Fatalf("missing unknown provider issue: %#v", issues)
	}
}

func hasIssue(issues []Issue, label, detail string) bool {
	for _, issue := range issues {
		if issue.Label == label && issue.Detail == detail {
			return true
		}
	}

	return false
}
