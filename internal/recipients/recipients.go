// Package recipients manages the public age recipients that a project's
// encrypted secrets are addressed to.
//
// The recipient list is public information and is intentionally stored next to
// the manifest in a plain-text file (secret-wan.recipients by default). Private
// keys never belong in this file.
package recipients

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"filippo.io/age"
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

// ParseLines converts file lines into normalized recipient strings. Empty
// lines, comments, and inline comments are ignored.
func ParseLines(lines []string) ([]string, error) {
	keys := make([]string, 0, len(lines))

	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if comment := strings.Index(line, "#"); comment >= 0 {
			line = strings.TrimSpace(line[:comment])
		}
		if line == "" {
			continue
		}

		recipient, err := age.ParseX25519Recipient(line)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient on line %d: %w", index+1, err)
		}

		keys = append(keys, recipient.String())
	}

	return Normalize(keys)
}

// Normalize validates, de-duplicates, and sorts recipient strings.
func Normalize(keys []string) ([]string, error) {
	seen := make(map[string]bool, len(keys))
	normalized := make([]string, 0, len(keys))

	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		recipient, err := age.ParseX25519Recipient(key)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient %q: %w", key, err)
		}

		value := recipient.String()
		if seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}

	sort.Strings(normalized)

	return normalized, nil
}

// Parse converts recipient strings into age recipients. Callers should pass a
// normalized list from Load or Normalize.
func Parse(keys []string) ([]age.Recipient, error) {
	result := make([]age.Recipient, 0, len(keys))

	for _, key := range keys {
		recipient, err := age.ParseX25519Recipient(key)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient %q: %w", key, err)
		}
		result = append(result, recipient)
	}

	return result, nil
}

// Contains reports whether key is present in keys.
func Contains(keys []string, key string) bool {
	for _, candidate := range keys {
		if candidate == key {
			return true
		}
	}

	return false
}

// Remove returns a copy of keys without key.
func Remove(keys []string, key string) []string {
	result := make([]string, 0, len(keys))
	for _, candidate := range keys {
		if candidate != key {
			result = append(result, candidate)
		}
	}

	return result
}
