package codemapbench

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestResolvedLinksFromDatasetFiltersAndNormalizes(t *testing.T) {
	dataset := codemap.Dataset{Entries: []codemap.DatasetEntry{
		{Entry: codemap.Entry{DocumentPath: "docs/b.md", Target: "src/b.go"}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved}},
		{Entry: codemap.Entry{DocumentPath: "docs/a.md", Target: "src/a.go"}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved}},
		{Entry: codemap.Entry{DocumentPath: "docs/a.md", Target: "src/a.go"}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved}},
		{Entry: codemap.Entry{DocumentPath: "docs/a.md", Target: "Handler"}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionUnsupported}},
	}}
	links := ResolvedLinksFromDataset(dataset)
	if len(links) != 2 || links[0] != (Link{Document: "docs/a.md", Target: "src/a.go"}) || links[1] != (Link{Document: "docs/b.md", Target: "src/b.go"}) {
		t.Fatalf("unexpected links: %#v", links)
	}
}

func TestSuggestionsFromEvidenceWeightsAndCapsOccurrences(t *testing.T) {
	candidates := []evidence.Candidate{{
		Path: "src/runtime.go",
		Evidence: []evidence.Evidence{
			{Kind: evidence.KindExactPathMention, Detail: "mentioned in prose", Count: 2},
			{Kind: evidence.KindTestCounterpart, Source: "src/runtime_test.go", Detail: "test counterpart", Count: 1},
		},
	}}
	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != 1 || suggestions[0].Score != 12 || suggestions[0].Document != "docs/runtime.md" || len(suggestions[0].Evidence) != 2 {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestSuggestionsFromEvidenceRejectsWeakSingleSignals(t *testing.T) {
	candidates := []evidence.Candidate{
		{Path: "src/history.go", Evidence: []evidence.Evidence{{Kind: evidence.KindGitDocumentCoChange, Source: "docs/runtime.md", Count: 10}}},
		{Path: "src/sibling.go", Evidence: []evidence.Evidence{{Kind: evidence.KindSiblingTarget, Source: "src/runtime.go", Count: 1}}},
	}
	if suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates); len(suggestions) != 0 {
		t.Fatalf("weak single-signal candidates were admitted: %#v", suggestions)
	}
}

func TestSuggestionsFromEvidenceDiscountsBroadFanout(t *testing.T) {
	candidates := []evidence.Candidate{{
		Path:     "src/narrow_test.go",
		Evidence: []evidence.Evidence{{Kind: evidence.KindTestCounterpart, Source: "src/narrow.go", Count: 1}},
	}}
	for index := 0; index < 4; index++ {
		candidates = append(candidates, evidence.Candidate{
			Path:     fmt.Sprintf("src/broad_%d.go", index),
			Evidence: []evidence.Evidence{{Kind: evidence.KindDependencyNeighbor, Source: "src/root.go", Detail: "outbound:imports", Count: 1}},
		})
	}
	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != 5 || suggestions[0].Target != "src/narrow_test.go" {
		t.Fatalf("broad fanout outranked narrow evidence: %#v", suggestions)
	}
}

func TestSuggestionsFromEvidenceBoundsEachDocument(t *testing.T) {
	candidates := make([]evidence.Candidate, 0, 40)
	for index := 0; index < 40; index++ {
		candidates = append(candidates, evidence.Candidate{
			Path:     fmt.Sprintf("src/file_%02d.go", index),
			Evidence: []evidence.Evidence{{Kind: evidence.KindExactPathMention, Count: 1}},
		})
	}
	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != DefaultSuggestionLimitPerDocument {
		t.Fatalf("got %d suggestions, want %d", len(suggestions), DefaultSuggestionLimitPerDocument)
	}
}

func TestTrustedReviewSetFeedsBenchmarkHarness(t *testing.T) {
	path := filepath.Join("..", "..", "research", "codemap-review", "space-rocks-trusted-links.json")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	links, err := DecodeTrustedReviewLinks(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 51 {
		t.Fatalf("got %d trusted links, want 51", len(links))
	}
	report, err := Run(context.Background(), links, GeneratorFunc(func(context.Context, Request) ([]Suggestion, error) {
		return nil, nil
	}), Config{HoldoutCount: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.HiddenLinks) != 10 || len(report.MissedLinks) != 10 || len(report.KnownLinks) != 51 {
		t.Fatalf("unexpected benchmark report: %#v", report)
	}
}
