package cmd

import (
	"sort"

	"github.com/ikhwan-satrio/secret-wan/utils"
)

func getProfileNames(profiles map[string]*utils.Profile) []string {
	names := make([]string, 0, len(profiles))
	for k := range profiles {
		names = append(names, k)
	}
	
	// Urutkan agar tampilannya rapi di terminal
	sort.Strings(names)
	
	return names
}
