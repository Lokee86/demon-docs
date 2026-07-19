package codemapbench

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// Run hides a deterministic subset of known links, invokes the generator with
// only the remaining links, and classifies every returned suggestion.
func Run(ctx context.Context, knownLinks []Link, generator Generator, config Config) (Report, error) {
	if generator == nil {
		return Report{}, errors.New("benchmark generator is required")
	}

	visible, hidden, seed, err := splitHoldout(knownLinks, config)
	if err != nil {
		return Report{}, err
	}
	known, _ := normalizeKnownLinks(knownLinks)
	request := Request{
		Documents:    documentPaths(known),
		VisibleLinks: append([]Link(nil), visible...),
	}

	raw, err := generator.Generate(ctx, request)
	if err != nil {
		return Report{}, fmt.Errorf("generate suggestions: %w", err)
	}

	report := Report{
		Seed:               seed,
		KnownLinks:         known,
		VisibleLinks:       visible,
		HiddenLinks:        hidden,
		RawSuggestionCount: len(raw),
	}
	classifySuggestions(&report, raw)
	calculateScores(&report)
	return report, nil
}

func classifySuggestions(report *Report, raw []Suggestion) {
	visible := linkSet(report.VisibleLinks)
	hidden := linkSet(report.HiddenLinks)
	recovered := make(map[string]Suggestion, len(report.HiddenLinks))
	seen := make(map[string]struct{}, len(raw))

	for index, suggestion := range raw {
		normalized, err := normalizeLink(suggestion.Link)
		if err != nil {
			report.InvalidSuggestions = append(report.InvalidSuggestions, InvalidSuggestion{
				Index: index, Suggestion: suggestion, Reason: err.Error(),
			})
			continue
		}
		suggestion.Link = normalized
		key := linkKey(normalized)
		if _, duplicate := seen[key]; duplicate {
			report.DuplicateSuggestions = append(report.DuplicateSuggestions, suggestion)
			continue
		}
		seen[key] = struct{}{}
		report.UniqueSuggestionCount++

		if _, ok := hidden[key]; ok {
			recovered[key] = suggestion
			continue
		}
		if _, ok := visible[key]; ok {
			report.AlreadyLinked = append(report.AlreadyLinked, suggestion)
			continue
		}
		report.UnmatchedSuggestions = append(report.UnmatchedSuggestions, suggestion)
	}

	for _, link := range report.HiddenLinks {
		if suggestion, ok := recovered[linkKey(link)]; ok {
			report.RecoveredLinks = append(report.RecoveredLinks, link)
			report.RecoveredSuggestions = append(report.RecoveredSuggestions, suggestion)
			continue
		}
		report.MissedLinks = append(report.MissedLinks, link)
	}
	sortSuggestions(report.UnmatchedSuggestions)
	sortSuggestions(report.AlreadyLinked)
	sortSuggestions(report.DuplicateSuggestions)
}

func calculateScores(report *Report) {
	predicted := len(report.RecoveredLinks) + len(report.UnmatchedSuggestions) + len(report.AlreadyLinked)
	if predicted > 0 {
		report.Precision = float64(len(report.RecoveredLinks)) / float64(predicted)
	}
	if len(report.HiddenLinks) > 0 {
		report.Recall = float64(len(report.RecoveredLinks)) / float64(len(report.HiddenLinks))
	}
}

func documentPaths(links []Link) []string {
	set := make(map[string]struct{}, len(links))
	for _, link := range links {
		set[link.Document] = struct{}{}
	}
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func linkSet(links []Link) map[string]struct{} {
	set := make(map[string]struct{}, len(links))
	for _, link := range links {
		set[linkKey(link)] = struct{}{}
	}
	return set
}
