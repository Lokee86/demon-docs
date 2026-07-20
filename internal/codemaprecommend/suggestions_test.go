package codemaprecommend

import (
	"reflect"
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestSuggestionsFromEvidenceAssignsProductionTiers(t *testing.T) {
	items := SuggestionsFromEvidence("docs/runtime.md", []evidence.Candidate{
		{
			Path:     "src/hard.go",
			Evidence: []evidence.Evidence{{Kind: evidence.KindDeclaredSymbolMention, Detail: "Runtime", Count: 1}},
		},
		{
			Path:     "src/context.go",
			Evidence: []evidence.Evidence{{Kind: evidence.KindExactPathMention, Detail: "src/context.go", Count: 1}},
		},
	})
	if len(items) != 2 {
		t.Fatalf("suggestions=%#v", items)
	}
	byTarget := map[string]SuggestionTier{}
	for _, item := range items {
		byTarget[item.Target] = item.Tier
	}
	if byTarget["src/hard.go"] != SuggestionTierHardLink || byTarget["src/context.go"] != SuggestionTierContext {
		t.Fatalf("unexpected tiers: %#v", byTarget)
	}
}

func TestSuggestionsFromEvidenceFiltersIncidentalLockfile(t *testing.T) {
	items := SuggestionsFromEvidence("docs/runtime.md", []evidence.Candidate{{
		Path:     "package-lock.json",
		Evidence: []evidence.Evidence{{Kind: evidence.KindUniqueBasenameMention, Detail: "package-lock.json", Count: 1}},
	}})
	if len(items) != 0 {
		t.Fatalf("incidental lockfile was retained: %#v", items)
	}
}

func TestSuggestionsFromEvidenceOrderingIsDeterministic(t *testing.T) {
	candidates := []evidence.Candidate{
		{Path: "src/z.go", Evidence: []evidence.Evidence{{Kind: evidence.KindExactPathMention, Detail: "z", Count: 1}}},
		{Path: "src/a.go", Evidence: []evidence.Evidence{{Kind: evidence.KindExactPathMention, Detail: "a", Count: 1}}},
	}
	first := SuggestionsFromEvidence("docs/runtime.md", candidates)
	second := SuggestionsFromEvidence("docs/runtime.md", append([]evidence.Candidate(nil), candidates...))
	if !reflect.DeepEqual(first, second) || len(first) != 2 || first[0].Target != "src/a.go" || first[1].Target != "src/z.go" {
		t.Fatalf("nondeterministic ordering: first=%#v second=%#v", first, second)
	}
}

func TestIsTestTarget(t *testing.T) {
	for _, target := range []string{"internal/runtime_test.go", "spec/runtime_spec.rb", "tests/runtime.js"} {
		if !IsTestTarget(target) {
			t.Fatalf("expected test target: %s", target)
		}
	}
	if IsTestTarget("internal/runtime.go") {
		t.Fatal("production file classified as test")
	}
}
