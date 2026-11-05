package cmd

import (
	"path/filepath"
	"reflect"
	"testing"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
)

func TestProfileKeysDeduplicatesAndSorts(t *testing.T) {
	profile := &manifest.Profile{
		Required: []string{"B", "A"},
		AutoGen:  []string{"A", "C"},
	}

	got := profileKeys(profile)
	want := []string{"A", "B", "C"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profileKeys() = %v, want %v", got, want)
	}
}

func TestRequiredKeysForPromptSkipsGenerated(t *testing.T) {
	profile := &manifest.Profile{
		Required: []string{"A", "B"},
		AutoGen:  []string{"B", "C"},
	}

	got := requiredKeysForPrompt(profile)
	want := []string{"A"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("requiredKeysForPrompt() = %v, want %v", got, want)
	}
}

func TestAllDeclaredKeysIncludesEveryProfile(t *testing.T) {
	parsed := &manifest.Manifest{
		Profiles: map[string]*manifest.Profile{
			"default": {Declared: []string{"A", "B"}},
			"staging": {Declared: []string{"C"}},
		},
	}

	got := allDeclaredKeys(parsed)
	for _, key := range []string{"A", "B", "C"} {
		if !got[key] {
			t.Errorf("allDeclaredKeys() missing %q", key)
		}
	}
	if got["D"] {
		t.Error("allDeclaredKeys() unexpectedly contains D")
	}
}

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
	if !hasIssue(issues, "project.provider", "unknown provider 'vault' (expected dotfiles, keyring, or docker)") {
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

func TestEncryptionRecipientsUsesCurrentKeyWhenNoFile(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.AppConfig{RecipientsFile: filepath.Join(t.TempDir(), "secret-wan.recipients")}
	got, err := encryptionRecipients(cfg, identity, false)
	if err != nil {
		t.Fatalf("encryptionRecipients() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("encryptionRecipients() returned %d recipients, want 1", len(got))
	}
}

func TestEncryptionRecipientsRequiresCurrentKey(t *testing.T) {
	self, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	other, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.AppConfig{RecipientsFile: filepath.Join(t.TempDir(), "secret-wan.recipients")}
	store := recipients.New(cfg.RecipientsPath())
	if err := store.Save([]string{other.Recipient().String()}); err != nil {
		t.Fatal(err)
	}

	if _, err := encryptionRecipients(cfg, self, false); err == nil {
		t.Fatal("encryptionRecipients() error = nil, want a lockout error")
	}

	got, err := encryptionRecipients(cfg, self, true)
	if err != nil {
		t.Fatalf("encryptionRecipients(allowSelfExclusion) error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("encryptionRecipients(allowSelfExclusion) returned %d recipients, want 1", len(got))
	}
}

func TestEncryptionRecipientsKeepsConfiguredKeys(t *testing.T) {
	self, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	other, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.AppConfig{RecipientsFile: filepath.Join(t.TempDir(), "secret-wan.recipients")}
	store := recipients.New(cfg.RecipientsPath())
	if err := store.Save([]string{self.Recipient().String(), other.Recipient().String()}); err != nil {
		t.Fatal(err)
	}

	got, err := encryptionRecipients(cfg, self, false)
	if err != nil {
		t.Fatalf("encryptionRecipients() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("encryptionRecipients() returned %d recipients, want 2", len(got))
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
