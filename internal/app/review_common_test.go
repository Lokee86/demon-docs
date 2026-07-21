package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/review"
)

func TestMergeReviewSuggestionsKeepsLinkSuggestionsWhenCodemapGenerationFails(t *testing.T) {
	linkSuggestion := review.Suggestion{
		ID:         "sg-link",
		Kind:       review.SuggestionLinkRepair,
		SourcePath: "docs/example.md",
	}
	var warnings bytes.Buffer

	got := mergeReviewSuggestions(
		[]review.Suggestion{linkSuggestion},
		[]review.Suggestion{{ID: "sg-codemap", Kind: review.SuggestionCodemap}},
		errors.New("reference not found"),
		&warnings,
	)

	if len(got) != 1 || got[0].ID != linkSuggestion.ID {
		t.Fatalf("suggestions = %#v, want retained link suggestion only", got)
	}
	if !strings.Contains(warnings.String(), "warning: codemap suggestions unavailable: reference not found") {
		t.Fatalf("warning = %q", warnings.String())
	}
}
