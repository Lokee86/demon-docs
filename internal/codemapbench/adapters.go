package codemapbench

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

// ResolvedLinksFromDataset converts exact, resolved authored codemap entries
// into benchmark links. Pattern families, unresolved symbols, and stale targets
// are intentionally excluded from exact-link recovery benchmarks.
func ResolvedLinksFromDataset(dataset codemap.Dataset) []Link {
	links := make([]Link, 0, len(dataset.Entries))
	for _, entry := range dataset.Entries {
		if entry.Resolution.Status != codemap.ResolutionResolved {
			continue
		}
		links = append(links, Link{Document: entry.Entry.DocumentPath, Target: entry.Entry.Target})
	}
	normalized, _ := normalizeKnownLinks(links)
	return normalized
}

// SuggestionsFromEvidence converts deterministic evidence candidates into the
// benchmark's suggestion shape. Scores are simple evidence occurrence totals;
// ranking policy can replace this adapter later without changing collection.
func SuggestionsFromEvidence(document string, candidates []evidence.Candidate) []Suggestion {
	result := make([]Suggestion, 0, len(candidates))
	for _, candidate := range candidates {
		suggestion := Suggestion{Link: Link{Document: document, Target: candidate.Path}}
		for _, item := range candidate.Evidence {
			suggestion.Score += float64(item.Count)
			detail := fmt.Sprintf("%s:%s", item.Kind, item.Detail)
			if item.Source != "" {
				detail = fmt.Sprintf("%s:%s:%s", item.Kind, item.Source, item.Detail)
			}
			suggestion.Evidence = append(suggestion.Evidence, detail)
		}
		sort.Strings(suggestion.Evidence)
		result = append(result, suggestion)
	}
	sortSuggestions(result)
	return result
}

// DecodeTrustedReviewLinks reads the conservative reviewed-link corpus used by
// the Space Rocks benchmark without coupling the benchmark to its prose fields.
func DecodeTrustedReviewLinks(reader io.Reader) ([]Link, error) {
	var review struct {
		Documents []struct {
			Document string `json:"document"`
			Links    []struct {
				Target string `json:"target"`
			} `json:"links"`
		} `json:"documents"`
	}
	if err := json.NewDecoder(reader).Decode(&review); err != nil {
		return nil, err
	}
	links := make([]Link, 0)
	for _, document := range review.Documents {
		for _, link := range document.Links {
			links = append(links, Link{Document: document.Document, Target: link.Target})
		}
	}
	normalized, err := normalizeKnownLinks(links)
	if err != nil {
		return nil, fmt.Errorf("trusted review links: %w", err)
	}
	return normalized, nil
}
