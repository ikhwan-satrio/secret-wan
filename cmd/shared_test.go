package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
)

func TestQuoteEnvValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`hello`, `"hello"`},
		{`a"b`, `"a\"b"`},
		{`$HOME`, `"\$HOME"`},
		{`a\b`, `"a\\b"`},
	}

	for _, test := range tests {
		if got := quoteEnvValue(test.input); got != test.want {
			t.Errorf("quoteEnvValue(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

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

	issues := collectStrictIssues(parsed, map[string]string{
		"A":     "value",
		"EXTRA": "value",
	})

	if len(issues) != 3 {
		t.Fatalf("collectStrictIssues() returned %d issues, want 3: %#v", len(issues), issues)
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

	issues := collectStrictIssues(parsed, map[string]string{"A": "value"})
	expected := fmt.Sprintf("unknown provider 'vault' (expected %s)", strings.Join(provider.Names(), ", "))
	if !hasIssue(issues, "project.provider", expected) {
		t.Fatalf("missing unknown provider issue: %#v", issues)
	}
}

func TestSecretType(t *testing.T) {
	parsed := &manifest.Manifest{
		Profiles: map[string]*manifest.Profile{
			"default": {
				Name:     "default",
				Required: []string{"A"},
				Declared: []string{"A"},
				Rules:    map[string]manifest.SecretRule{"A": {Required: true}},
			},
			"staging": {
				Name:     "staging",
				Required: []string{"B"},
				Declared: []string{"B"},
				Rules:    map[string]manifest.SecretRule{"B": {Required: true}},
			},
		},
	}

	if got := secretType(parsed, parsed.Profiles["default"], "A"); got != "required" {
		t.Errorf("secretType(A) = %q, want required", got)
	}
	if got := secretType(parsed, parsed.Profiles["default"], "B"); got != "other" {
		t.Errorf("secretType(B) = %q, want other", got)
	}
	if got := secretType(parsed, parsed.Profiles["default"], "Z"); got != "unmanaged" {
		t.Errorf("secretType(Z) = %q, want unmanaged", got)
	}
}

func hasIssue(issues []strictIssue, label, detail string) bool {
	for _, issue := range issues {
		if issue.label == label && issue.detail == detail {
			return true
		}
	}

	return false
}
