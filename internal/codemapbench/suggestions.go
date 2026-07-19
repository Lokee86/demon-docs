package codemapbench

import (
	"fmt"
	"math"
	"path"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

// DefaultSuggestionLimitPerDocument bounds the ranked list returned per document.
// The benchmark is intended to measure useful surfaced suggestions, not every
// weak relationship that can be inferred from repository history.
const (
	DefaultSuggestionLimitPerDocument             = 30
	HardLinkSuggestionLimitPerDocument            = 5
	HardLinkDependencyMinimumScore                = 18
	HardLinkImplementationCounterpartMinimumScore = 20
	RepeatedMentionReservePerDocument             = 2
	RepeatedMentionMinimumCount                   = 2
)

type evidenceAtom struct {
	kind   evidence.Kind
	source string
	detail string
}

type rankedSuggestion struct {
	suggestion               Suggestion
	repeatedMentionCount     int
	hasExactPathMention      bool
	hasDeclaredSymbolMention bool
	hasTestCounterpart       bool
	hasDependencyNeighbor    bool
	hasRelatedDocumentTarget bool
	hasSiblingTarget         bool
	hasGitDocumentCoChange   bool
	targetIsTest             bool
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
			suggestion:   Suggestion{Link: Link{Document: document, Target: candidate.Path}},
			targetIsTest: isTestTarget(candidate.Path),
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
			switch item.Kind {
			case evidence.KindExactPathMention:
				itemResult.hasExactPathMention = true
			case evidence.KindDeclaredSymbolMention:
				itemResult.hasDeclaredSymbolMention = true
			case evidence.KindTestCounterpart:
				itemResult.hasTestCounterpart = true
			case evidence.KindDependencyNeighbor:
				itemResult.hasDependencyNeighbor = true
			case evidence.KindRelatedDocumentTarget:
				itemResult.hasRelatedDocumentTarget = true
			case evidence.KindSiblingTarget:
				itemResult.hasSiblingTarget = true
			case evidence.KindGitDocumentCoChange:
				itemResult.hasGitDocumentCoChange = true
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

	ordered := make([]rankedSuggestion, 0, len(selected))
	for _, item := range selected {
		ordered = append(ordered, item)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].suggestion.Score != ordered[j].suggestion.Score {
			return ordered[i].suggestion.Score > ordered[j].suggestion.Score
		}
		return ordered[i].suggestion.Target < ordered[j].suggestion.Target
	})

	result := make([]Suggestion, 0, len(ordered))
	hardLinks := 0
	for _, item := range ordered {
		suggestion := item.suggestion
		suggestion.Tier = SuggestionTierContext
		if hardLinks < HardLinkSuggestionLimitPerDocument && item.isHardLinkCandidate() {
			suggestion.Tier = SuggestionTierHardLink
			hardLinks++
		}
		result = append(result, suggestion)
	}
	return result
}

func (item rankedSuggestion) isHardLinkCandidate() bool {
	// A single exact path mention is explicit document context, but repeated
	// references plus independent semantic structure indicate that the document
	// relies on the target enough to justify a permanent link.
	if item.hasExactPathMention {
		return item.repeatedMentionCount >= RepeatedMentionMinimumCount &&
			(item.hasDeclaredSymbolMention || item.hasDependencyNeighbor)
	}
	if item.hasDeclaredSymbolMention {
		return true
	}
	// Filename-based test counterparts need independent semantic support so a
	// similarly named test in another service cannot qualify by structure alone.
	if item.hasTestCounterpart && (item.hasDependencyNeighbor || item.hasRelatedDocumentTarget || item.hasSiblingTarget) {
		return item.targetIsTest || item.suggestion.Score >= HardLinkImplementationCounterpartMinimumScore
	}
	if item.hasDependencyNeighbor && item.suggestion.Score >= HardLinkDependencyMinimumScore {
		return true
	}
	// A target inherited through a related document becomes link-worthy when it
	// also changed directly with the current document.
	return item.hasRelatedDocumentTarget && item.hasGitDocumentCoChange
}

func isTestTarget(value string) bool {
	value = strings.ReplaceAll(value, "\\", "/")
	for _, segment := range strings.Split(strings.ToLower(path.Dir(value)), "/") {
		if segment == "test" || segment == "tests" || segment == "spec" || segment == "specs" {
			return true
		}
	}
	base := strings.ToLower(strings.TrimSuffix(path.Base(value), path.Ext(value)))
	for _, prefix := range []string{"test_", "spec_"} {
		if strings.HasPrefix(base, prefix) {
			return true
		}
	}
	for _, suffix := range []string{"_test", "_spec", ".test", ".spec"} {
		if strings.HasSuffix(base, suffix) {
			return true
		}
	}
	return false
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
