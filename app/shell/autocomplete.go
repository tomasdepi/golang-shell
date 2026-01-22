package shell

import (
	"os"
	"slices"
	"strings"
)

type Autocomplete struct {
	Entries []string
}

func (au *Autocomplete) DiscoverEntiesInPath() {

	pathEnvVar := os.Getenv("PATH")
	if pathEnvVar == "" {
		//empty
		return
	}

	paths := strings.Split(pathEnvVar, ":")

	for _, path := range paths {
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {

			if !slices.Contains(au.Entries, entry.Name()) {
				au.Entries = append(au.Entries, entry.Name())
			}

		}
	}

}

func (au *Autocomplete) GetEntriesMatching(prefix string) []string {

	matches := []string{}

	for _, entry := range au.Entries {
		if strings.HasPrefix(entry, prefix) {
			matches = append(matches, entry)
		}
	}

	return matches
}

func (au *Autocomplete) CommonPrefix(strs []string) string {

	prefix := strs[0]

	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			if len(prefix) == 0 {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix
}
