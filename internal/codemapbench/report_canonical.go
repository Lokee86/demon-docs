package codemapbench

import (
	"sort"
	"strconv"
	"strings"
)

func canonicalReport(report Report) Report {
	result := report
	result.KnownLinks = canonicalLinks(report.KnownLinks)
	result.VisibleLinks = canonicalLinks(report.VisibleLinks)
	result.HiddenLinks = canonicalLinks(report.HiddenLinks)
	result.RecoveredLinks = canonicalLinks(report.RecoveredLinks)
	result.MissedLinks = canonicalLinks(report.MissedLinks)
	result.RecoveredSuggestions = canonicalSuggestions(report.RecoveredSuggestions)
	result.UnmatchedSuggestions = canonicalSuggestions(report.UnmatchedSuggestions)
	result.AlreadyLinked = canonicalSuggestions(report.AlreadyLinked)
	result.DuplicateSuggestions = canonicalSuggestions(report.DuplicateSuggestions)
	result.InvalidSuggestions = canonicalInvalidSuggestions(report.InvalidSuggestions)
	return result
}

func canonicalLinks(links []Link) []Link {
	result := append([]Link{}, links...)
	sortLinks(result)
	return result
}

func canonicalSuggestions(suggestions []Suggestion) []Suggestion {
	result := make([]Suggestion, len(suggestions))
	for index, suggestion := range suggestions {
		result[index] = canonicalSuggestion(suggestion)
	}
	sort.SliceStable(result, func(left, right int) bool {
		return suggestionSortKey(result[left]) < suggestionSortKey(result[right])
	})
	return result
}

func canonicalInvalidSuggestions(suggestions []InvalidSuggestion) []InvalidSuggestion {
	result := make([]InvalidSuggestion, len(suggestions))
	for index, invalid := range suggestions {
		invalid.Suggestion = canonicalSuggestion(invalid.Suggestion)
		result[index] = invalid
	}
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].Index != result[right].Index {
			return result[left].Index < result[right].Index
		}
		leftKey := suggestionSortKey(result[left].Suggestion) + "\x00" + result[left].Reason
		rightKey := suggestionSortKey(result[right].Suggestion) + "\x00" + result[right].Reason
		return leftKey < rightKey
	})
	return result
}

func canonicalSuggestion(suggestion Suggestion) Suggestion {
	suggestion.Evidence = append([]string{}, suggestion.Evidence...)
	sort.Strings(suggestion.Evidence)
	return suggestion
}

func suggestionSortKey(suggestion Suggestion) string {
	return linkKey(suggestion.Link) + "\x00" +
		strconv.FormatFloat(suggestion.Score, 'g', -1, 64) + "\x00" +
		strings.Join(suggestion.Evidence, "\x00")
}
