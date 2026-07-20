package codemapbench

import (
	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

const (
	DefaultSuggestionLimitPerDocument             = codemaprecommend.DefaultSuggestionLimitPerDocument
	HardLinkSuggestionLimitPerDocument            = codemaprecommend.HardLinkSuggestionLimitPerDocument
	HardLinkDependencyMinimumScore                = codemaprecommend.HardLinkDependencyMinimumScore
	HardLinkImplementationCounterpartMinimumScore = codemaprecommend.HardLinkImplementationCounterpartMinimumScore
	RepeatedMentionReservePerDocument             = codemaprecommend.RepeatedMentionReservePerDocument
	RepeatedMentionMinimumCount                   = codemaprecommend.RepeatedMentionMinimumCount
)

// SuggestionsFromEvidence remains as a benchmark compatibility seam. The
// production implementation is owned by codemaprecommend.
func SuggestionsFromEvidence(document string, candidates []evidence.Candidate) []Suggestion {
	return codemaprecommend.SuggestionsFromEvidence(document, candidates)
}

func isTestTarget(target string) bool {
	return codemaprecommend.IsTestTarget(target)
}
