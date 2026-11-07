// Package recipients manages the public age recipients that a project's
// encrypted secrets are addressed to.
//
// The recipient list is public information and is intentionally stored next to
// the manifest in a plain-text file (secret-wan.recipients by default). Private
// keys never belong in this file.
package recipients

import (
	"os"
	"strings"
)

// DefaultFile is used when no explicit recipients path is configured.
const DefaultFile = "secret-wan.recipients"

// Store reads and writes a recipients file.
type Store struct {
	Path string
}

// New returns a Store for path, falling back to DefaultFile when path is empty.
func New(path string) Store {
	if strings.TrimSpace(path) == "" {
		path = DefaultFile
	}

	return Store{Path: path}
}

// Exists reports whether the recipients file is present.
func (s Store) Exists() (bool, error) {
	_, err := os.Stat(s.Path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// Load reads, validates, normalizes, de-duplicates, and sorts the recipients.
// A missing file is not an error; it means the project is still in
// single-recipient mode.
func (s Store) Load() ([]string, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return ParseLines(strings.Split(string(data), "\n"))
}

// Save writes the normalized recipient list with a short header. The file is
// public key material, so it is written with mode 0644.
func (s Store) Save(keys []string) error {
	normalized, err := Normalize(keys)
	if err != nil {
		return err
	}

	var builder strings.Builder
	builder.WriteString("# secret-wan recipients\n")
	builder.WriteString("# Public age keys, one per line. This file is safe to commit.\n")
	builder.WriteString("# Add a teammate with: secwan recipient add <age1...>\n")

	if len(normalized) > 0 {
		builder.WriteString("\n")
		for _, key := range normalized {
			builder.WriteString(key)
			builder.WriteString("\n")
		}
	}

	return os.WriteFile(s.Path, []byte(builder.String()), 0644)
}
