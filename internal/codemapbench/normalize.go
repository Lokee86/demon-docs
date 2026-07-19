package codemapbench

import (
	"errors"
	"sort"
	"strings"
)

func normalizeKnownLinks(links []Link) ([]Link, error) {
	unique := make(map[string]Link, len(links))
	for _, link := range links {
		normalized, err := normalizeLink(link)
		if err != nil {
			return nil, err
		}
		unique[linkKey(normalized)] = normalized
	}

	result := make([]Link, 0, len(unique))
	for _, link := range unique {
		result = append(result, link)
	}
	sortLinks(result)
	return result, nil
}

func normalizeLink(link Link) (Link, error) {
	link.Document = normalizeReference(link.Document)
	link.Target = normalizeReference(link.Target)
	if link.Document == "" {
		return Link{}, errors.New("link document cannot be empty")
	}
	if link.Target == "" {
		return Link{}, errors.New("link target cannot be empty")
	}
	return link, nil
}

func normalizeReference(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\", "/")
	for strings.HasPrefix(value, "./") {
		value = strings.TrimPrefix(value, "./")
	}
	return value
}

func linkKey(link Link) string {
	return link.Document + "\x00" + link.Target
}

func sortLinks(links []Link) {
	sort.Slice(links, func(i, j int) bool {
		return linkKey(links[i]) < linkKey(links[j])
	})
}

func sortSuggestions(suggestions []Suggestion) {
	sort.SliceStable(suggestions, func(i, j int) bool {
		return linkKey(suggestions[i].Link) < linkKey(suggestions[j].Link)
	})
}
