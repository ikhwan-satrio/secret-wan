package manifest

import (
	"sort"
)

func GetProfileNames(profiles map[string]*Profile) []string {
	names := make([]string, 0, len(profiles))
	for k := range profiles {
		names = append(names, k)
	}

	// Urutkan agar tampilannya rapi di terminal
	sort.Strings(names)

	return names
}
