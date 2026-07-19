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

func TestSuggestionsFromEvidenceKeepsExplicitPathMentionsAsContext(t *testing.T) {
	candidates := []evidence.Candidate{{
		Path: "src/runtime.go",
		Evidence: []evidence.Evidence{
			{Kind: evidence.KindExactPathMention, Detail: "mentioned in prose", Count: 2},
			{Kind: evidence.KindTestCounterpart, Source: "src/runtime_test.go", Detail: "test counterpart", Count: 1},
		},
	}}
	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != 1 || suggestions[0].Score != 12 || suggestions[0].Document != "docs/runtime.md" || len(suggestions[0].Evidence) != 2 || suggestions[0].Tier != SuggestionTierContext {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestSuggestionsFromEvidenceSeparatesHardLinksFromContext(t *testing.T) {
	dependencyEvidence := func(prefix string, count int) []evidence.Evidence {
		items := make([]evidence.Evidence, 0, count)
		for index := 0; index < count; index++ {
			items = append(items, evidence.Evidence{
				Kind:   evidence.KindDependencyNeighbor,
				Source: fmt.Sprintf("src/%s_source_%d.go", prefix, index),
				Detail: "outbound:go_import",
				Count:  1,
			})
		}
		return items
	}
	candidates := []evidence.Candidate{
		{Path: "src/dependency_hard.go", Evidence: dependencyEvidence("hard", 5)},
		{Path: "src/dependency_context.go", Evidence: dependencyEvidence("context", 4)},
		{Path: "src/explicit_context.go", Evidence: []evidence.Evidence{{Kind: evidence.KindExactPathMention, Count: 1}}},
		{Path: "src/explicit_symbol_context.go", Evidence: []evidence.Evidence{
			{Kind: evidence.KindExactPathMention, Count: 1},
			{Kind: evidence.KindDeclaredSymbolMention, Detail: "ExplicitSymbol", Count: 1},
		}},
		{Path: "src/test_only_context.go", Evidence: []evidence.Evidence{{Kind: evidence.KindTestCounterpart, Source: "src/test_only.go", Count: 1}}},
		{Path: "src/test_history_context.go", Evidence: []evidence.Evidence{
			{Kind: evidence.KindTestCounterpart, Source: "src/test_history.go", Count: 1},
			{Kind: evidence.KindGitTargetCoChange, Source: "src/owner.go", Count: 2},
		}},
		{Path: "src/test_supported_hard_test.go", Evidence: []evidence.Evidence{
			{Kind: evidence.KindTestCounterpart, Source: "src/test_supported_hard.go", Count: 1},
			{Kind: evidence.KindSiblingTarget, Source: "src/owner.go", Count: 1},
		}},
		{Path: "src/implementation_counterpart_context.go", Evidence: []evidence.Evidence{
			{Kind: evidence.KindTestCounterpart, Source: "src/implementation_counterpart_context_test.go", Count: 1},
			{Kind: evidence.KindSiblingTarget, Source: "src/owner.go", Count: 1},
		}},
		{Path: "src/implementation_counterpart_hard.go", Evidence: []evidence.Evidence{
			{Kind: evidence.KindTestCounterpart, Source: "src/implementation_counterpart_hard_test.go", Count: 1},
			{Kind: evidence.KindDependencyNeighbor, Source: "src/dependency_a.go", Detail: "outbound:go_import", Count: 1},
			{Kind: evidence.KindDependencyNeighbor, Source: "src/dependency_b.go", Detail: "outbound:go_import", Count: 1},
			{Kind: evidence.KindDependencyNeighbor, Source: "src/dependency_c.go", Detail: "outbound:go_import", Count: 1},
			{Kind: evidence.KindDependencyNeighbor, Source: "src/dependency_d.go", Detail: "outbound:go_import", Count: 1},
		}},
		{Path: "src/related_only_context.go", Evidence: []evidence.Evidence{{Kind: evidence.KindRelatedDocumentTarget, Source: "docs/related.md", Count: 1}}},
		{Path: "src/related_hard.go", Evidence: []evidence.Evidence{
			{Kind: evidence.KindRelatedDocumentTarget, Source: "docs/related.md", Count: 1},
			{Kind: evidence.KindGitDocumentCoChange, Source: "docs/runtime.md", Count: 1},
		}},
	}

	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != len(candidates) {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
	byTarget := make(map[string]Suggestion, len(suggestions))
	for _, suggestion := range suggestions {
		byTarget[suggestion.Target] = suggestion
	}
	if got := byTarget["src/dependency_hard.go"]; got.Score < HardLinkDependencyMinimumScore || got.Tier != SuggestionTierHardLink {
		t.Fatalf("dependency-backed hard link = %#v", got)
	}
	for _, target := range []string{"src/related_hard.go", "src/test_supported_hard_test.go", "src/implementation_counterpart_hard.go"} {
		if got := byTarget[target]; got.Tier != SuggestionTierHardLink {
			t.Fatalf("%s tier = %q, want hard link: %#v", target, got.Tier, suggestions)
		}
	}
	for _, target := range []string{
		"src/dependency_context.go",
		"src/explicit_context.go",
		"src/explicit_symbol_context.go",
		"src/test_only_context.go",
		"src/test_history_context.go",
		"src/implementation_counterpart_context.go",
		"src/related_only_context.go",
	} {
		if got := byTarget[target]; got.Tier != SuggestionTierContext {
			t.Fatalf("%s tier = %q, want context: %#v", target, got.Tier, suggestions)
		}
	}
}

func TestIsTestTargetRecognizesCommonConventions(t *testing.T) {
	for _, target := range []string{
		"src/runtime_test.go",
		"src/tests/runtime.go",
		"src/spec/runtime.rb",
		"src/test_runtime.py",
		"src/runtime.spec.ts",
	} {
		if !isTestTarget(target) {
			t.Fatalf("%q was not recognized as a test target", target)
		}
	}
	for _, target := range []string{"src/runtime.go", "src/testing/runtime.go", "src/contest/runtime.ts"} {
		if isTestTarget(target) {
			t.Fatalf("%q was incorrectly recognized as a test target", target)
		}
	}
}

func TestSuggestionsFromEvidenceCapsHardLinkSurface(t *testing.T) {
	candidates := make([]evidence.Candidate, 0, HardLinkSuggestionLimitPerDocument+1)
	for index := 0; index <= HardLinkSuggestionLimitPerDocument; index++ {
		candidates = append(candidates, evidence.Candidate{
			Path: fmt.Sprintf("src/symbol_%02d.go", index),
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindDeclaredSymbolMention,
				Detail: fmt.Sprintf("Symbol%d", index),
				Count:  1,
			}},
		})
	}

	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	for index, suggestion := range suggestions {
		want := SuggestionTierContext
		if index < HardLinkSuggestionLimitPerDocument {
			want = SuggestionTierHardLink
		}
		if suggestion.Tier != want {
			t.Fatalf("suggestion %d tier = %q, want %q: %#v", index, suggestion.Tier, want, suggestions)
		}
	}
}

func TestSuggestionsFromEvidenceFillsHardLinkSlotsPastContextCandidates(t *testing.T) {
	candidates := make([]evidence.Candidate, 0, HardLinkSuggestionLimitPerDocument+1)
	for index := 0; index < HardLinkSuggestionLimitPerDocument; index++ {
		candidates = append(candidates, evidence.Candidate{
			Path: fmt.Sprintf("src/explicit_%02d.go", index),
			Evidence: []evidence.Evidence{
				{Kind: evidence.KindExactPathMention, Count: 1},
				{Kind: evidence.KindGitTargetCoChange, Source: fmt.Sprintf("src/owner_%02d.go", index), Count: 8},
			},
		})
	}
	candidates = append(candidates, evidence.Candidate{
		Path: "src/lower_rank_symbol.go",
		Evidence: []evidence.Evidence{{
			Kind:   evidence.KindDeclaredSymbolMention,
			Detail: "LowerRankSymbol",
			Count:  1,
		}},
	})

	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	var lowerRank Suggestion
	for _, suggestion := range suggestions {
		if suggestion.Target == "src/lower_rank_symbol.go" {
			lowerRank = suggestion
		}
	}
	if lowerRank.Tier != SuggestionTierHardLink {
		t.Fatalf("lower-ranked qualifying candidate was not promoted: %#v", suggestions)
	}
}

func TestSuggestionsFromEvidenceDoesNotPromoteBeyondSuggestionLimit(t *testing.T) {
	candidates := make([]evidence.Candidate, 0, DefaultSuggestionLimitPerDocument+1)
	for index := 0; index < DefaultSuggestionLimitPerDocument; index++ {
		candidates = append(candidates, evidence.Candidate{
			Path: fmt.Sprintf("src/explicit_%02d.go", index),
			Evidence: []evidence.Evidence{
				{Kind: evidence.KindExactPathMention, Count: 1},
				{Kind: evidence.KindGitTargetCoChange, Source: fmt.Sprintf("src/owner_%02d.go", index), Count: 8},
			},
		})
	}
	candidates = append(candidates, evidence.Candidate{
		Path: "src/outside_limit_symbol.go",
		Evidence: []evidence.Evidence{{
			Kind:   evidence.KindDeclaredSymbolMention,
			Detail: "OutsideLimitSymbol",
			Count:  1,
		}},
	})

	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != DefaultSuggestionLimitPerDocument {
		t.Fatalf("got %d suggestions, want %d", len(suggestions), DefaultSuggestionLimitPerDocument)
	}
	for _, suggestion := range suggestions {
		if suggestion.Target == "src/outside_limit_symbol.go" {
			t.Fatalf("candidate outside output limit was promoted: %#v", suggestions)
		}
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
