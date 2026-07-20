package documentpolicy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Fingerprint(schema Schema) string {
	canonical := schema
	canonical.Description = ""
	data, _ := json.Marshal(canonical)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func Similarity(previous, current Schema) float64 {
	if len(previous.Sections) == 0 && len(current.Sections) == 0 {
		return 1
	}
	before := sectionsByID(previous.Sections)
	after := sectionsByID(current.Sections)
	denominator := len(before)
	if len(after) > denominator {
		denominator = len(after)
	}
	if denominator == 0 {
		return 1
	}
	unchanged := 0
	for id, section := range before {
		candidate, exists := after[id]
		if exists && canonicalSection(section) == canonicalSection(candidate) {
			unchanged++
		}
	}
	return float64(unchanged) / float64(denominator)
}

func sectionsByID(sections []Section) map[string]Section {
	result := make(map[string]Section, len(sections))
	for _, section := range sections {
		result[section.ID] = section
	}
	return result
}

func canonicalSection(section Section) string {
	aliases := append([]string(nil), section.Aliases...)
	for index := range aliases {
		aliases[index] = strings.ToLower(strings.TrimSpace(aliases[index]))
	}
	sort.Strings(aliases)
	return strings.Join([]string{
		strings.TrimSpace(section.ID),
		strings.TrimSpace(section.Heading),
		strings.TrimSpace(section.Parent),
		strings.TrimSpace(section.After),
		section.Placeholder,
		strings.Join(aliases, "|"),
		fmt.Sprint(section.Optional),
		fmt.Sprint(section.AllowDuplicates),
		fmt.Sprint(section.CanonicalizeAliases),
	}, "\x00")
}
