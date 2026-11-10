package cmd

import (
	"testing"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
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
