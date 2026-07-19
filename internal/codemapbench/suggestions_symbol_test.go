package codemapbench

import (
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestSuggestionsFromEvidenceAdmitsDeclaredSymbolMentions(t *testing.T) {
	candidate := evidence.Candidate{
		Path: "services/player-data/playerdata/runtime.go",
		Evidence: []evidence.Evidence{{
			Kind:   evidence.KindDeclaredSymbolMention,
			Detail: "Runtime.LoadStats",
			Count:  2,
		}},
	}

	suggestions := SuggestionsFromEvidence("docs/profile.md", []evidence.Candidate{candidate})
	if len(suggestions) != 1 || suggestions[0].Target != candidate.Path || suggestions[0].Score <= 7 {
		t.Fatalf("unexpected symbol suggestion: %#v", suggestions)
	}
}
