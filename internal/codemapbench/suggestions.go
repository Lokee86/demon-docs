package codemapbench

import (
	"fmt"
	"math"
	"sort"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

// DefaultSuggestionLimitPerDocument bounds the ranked list returned per document.
// The benchmark is intended to measure useful surfaced suggestions, not every
// weak relationship that can be inferred from repository history.
const (
	DefaultSuggestionLimitPerDocument = 30
	RepeatedMentionReservePerDocument = 2
	RepeatedMentionMinimumCount       = 2
)

type evidenceAtom struct {
	kind   evidence.Kind
	source string
	detail string
}

type rankedSuggestion struct {
	suggestion           Suggestion
	repeatedMentionCount int
}

// SuggestionsFromEvidence ranks deterministic evidence candidates. Repeated
// evidence is logarithmically capped, and evidence sources with broad fanout
// are discounted so one large commit or directory does not dominate the list.
func SuggestionsFromEvidence(document string, candidates []evidence.Candidate) []Suggestion {
	fanout := make(map[evidenceAtom]int)
	for _, candidate := range candidates {
		for _, item := range candidate.Evidence {
			fanout[suggestionEvidenceAtom(item)]++
		}
	}

	ranked := make([]rankedSuggestion, 0, len(candidates))
	for _, candidate := range candidates {
		if !admitSuggestionCandidate(candidate) {
			continue
		}
		itemResult := rankedSuggestion{
			suggestion: Suggestion{Link: Link{Document: document, Target: candidate.Path}},
		}
		for _, item := range candidate.Evidence {
			atom := suggestionEvidenceAtom(item)
			occurrences := item.Count
			if occurrences < 1 {
				occurrences = 1
			}
			occurrenceFactor := 1 + math.Log2(float64(occurrences))
			if item.Kind == evidence.KindExactPathMention || item.Kind == evidence.KindUniqueBasenameMention {
				occurrenceFactor = 1
			}
			breadth := fanout[atom]
			if breadth < 1 {
				breadth = 1
			}
			itemResult.suggestion.Score += evidenceWeight(item.Kind) *
				occurrenceFactor /
				math.Log2(float64(breadth+1))

			detail := fmt.Sprintf("%s:%s", item.Kind, item.Detail)
			if item.Source != "" {
				detail = fmt.Sprintf("%s:%s:%s", item.Kind, item.Source, item.Detail)
			}
			if item.Count > 1 {
				detail = fmt.Sprintf("%s:x%d", detail, item.Count)
			}
			itemResult.suggestion.Evidence = append(itemResult.suggestion.Evidence, detail)
			if item.Kind == evidence.KindExactPathMention && item.Count >= RepeatedMentionMinimumCount {
				itemResult.repeatedMentionCount = item.Count
			}
		}
		sort.Strings(itemResult.suggestion.Evidence)
		ranked = append(ranked, itemResult)
	}

	return selectRankedSuggestions(ranked)
}

func selectRankedSuggestions(ranked []rankedSuggestion) []Suggestion {
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].suggestion.Score != ranked[j].suggestion.Score {
			return ranked[i].suggestion.Score > ranked[j].suggestion.Score
		}
		return ranked[i].suggestion.Target < ranked[j].suggestion.Target
	})

	selected := make(map[string]rankedSuggestion)
	limit := min(DefaultSuggestionLimitPerDocument, len(ranked))
	for _, item := range ranked[:limit] {
		selected[item.suggestion.Target] = item
	}

	repeated := append([]rankedSuggestion(nil), ranked...)
	sort.Slice(repeated, func(i, j int) bool {
		if repeated[i].repeatedMentionCount != repeated[j].repeatedMentionCount {
			return repeated[i].repeatedMentionCount > repeated[j].repeatedMentionCount
		}
		if repeated[i].suggestion.Score != repeated[j].suggestion.Score {
			return repeated[i].suggestion.Score > repeated[j].suggestion.Score
		}
		return repeated[i].suggestion.Target < repeated[j].suggestion.Target
	})
	reserved := 0
	for _, item := range repeated {
		if item.repeatedMentionCount < RepeatedMentionMinimumCount || reserved >= RepeatedMentionReservePerDocument {
			break
		}
		if _, exists := selected[item.suggestion.Target]; exists {
			continue
		}
		selected[item.suggestion.Target] = item
		reserved++
	}

	result := make([]Suggestion, 0, len(selected))
	for _, item := range selected {
		result = append(result, item.suggestion)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		return result[i].Target < result[j].Target
	})
	return result
}

func suggestionEvidenceAtom(item evidence.Evidence) evidenceAtom {
	source := item.Source
	if source == "" {
		source = string(item.Kind)
	}
	detail := ""
	if item.Kind == evidence.KindDependencyNeighbor || item.Kind == evidence.KindDeclaredSymbolMention {
		detail = item.Detail
	}
	return evidenceAtom{kind: item.Kind, source: source, detail: detail}
}

func admitSuggestionCandidate(candidate evidence.Candidate) bool {
	kinds := make(map[evidence.Kind]struct{}, len(candidate.Evidence))
	for _, item := range candidate.Evidence {
		kinds[item.Kind] = struct{}{}
	}
	if len(kinds) >= 2 {
		return true
	}
	for kind := range kinds {
		switch kind {
		case evidence.KindExactPathMention,
			evidence.KindUniqueBasenameMention,
			evidence.KindDeclaredSymbolMention,
			evidence.KindTestCounterpart,
			evidence.KindDependencyNeighbor,
			evidence.KindRelatedDocumentTarget:
			return true
		}
	}
	return false
}

func evidenceWeight(kind evidence.Kind) float64 {
	switch kind {
	case evidence.KindExactPathMention:
		return 6
	case evidence.KindUniqueBasenameMention:
		return 4
	case evidence.KindDeclaredSymbolMention:
		return 7
	case evidence.KindTestCounterpart:
		return 6
	case evidence.KindDependencyNeighbor:
		return 4
	case evidence.KindRelatedDocumentTarget:
		return 4
	case evidence.KindSiblingTarget:
		return 2
	case evidence.KindGitTargetCoChange:
		return 1.5
	case evidence.KindGitDocumentCoChange:
		return 1
	default:
		return 1
	}
}
