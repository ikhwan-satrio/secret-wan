package recipients

import (
	"fmt"
	"sort"
	"strings"

	"filippo.io/age"
)

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

// ParseRecipient validates a single recipient string and returns its canonical
// form. Empty input is rejected with a user-facing message.
func ParseRecipient(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("recipient must not be empty")
	}

	recipient, err := age.ParseX25519Recipient(raw)
	if err != nil {
		return "", err
	}

	return recipient.String(), nil
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
