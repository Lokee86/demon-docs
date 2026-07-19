package codemapbench

import (
	"fmt"
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestSuggestionsFromEvidenceReservesRepeatedExplicitMentions(t *testing.T) {
	candidates := make([]evidence.Candidate, 0, DefaultSuggestionLimitPerDocument+1)
	for index := 0; index < DefaultSuggestionLimitPerDocument; index++ {
		path := fmt.Sprintf("src/a_%02d.go", index)
		candidates = append(candidates, evidence.Candidate{
			Path: path,
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindExactPathMention,
				Detail: path,
				Count:  1,
			}},
		})
	}
	repeatedPath := "src/z_repeated.go"
	candidates = append(candidates, evidence.Candidate{
		Path: repeatedPath,
		Evidence: []evidence.Evidence{{
			Kind:   evidence.KindExactPathMention,
			Detail: repeatedPath,
			Count:  2,
		}},
	})

	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != DefaultSuggestionLimitPerDocument+1 {
		t.Fatalf("got %d suggestions, want %d", len(suggestions), DefaultSuggestionLimitPerDocument+1)
	}
	for _, suggestion := range suggestions {
		if suggestion.Target == repeatedPath {
			return
		}
	}
	t.Fatalf("repeated explicit mention was not reserved: %#v", suggestions)
}
