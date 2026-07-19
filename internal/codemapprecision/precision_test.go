package codemapprecision

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemapbench"
)

func TestCandidatesFromReportRanksByScoreAndTarget(t *testing.T) {
	report := codemapbench.Report{UnmatchedSuggestions: []codemapbench.Suggestion{
		{Link: codemapbench.Link{Document: "docs/b.md", Target: "z.go"}, Score: 1},
		{Link: codemapbench.Link{Document: "docs/a.md", Target: "b.go"}, Score: 2},
		{Link: codemapbench.Link{Document: "docs/a.md", Target: "a.go"}, Score: 2},
	}}
	candidates := CandidatesFromReport(report)
	if candidates[0].Document != "docs/a.md" || candidates[0].Target != "a.go" || candidates[0].Rank != 1 || candidates[1].Rank != 2 {
		t.Fatalf("unexpected ranking: %#v", candidates)
	}
}

func TestSampleIsDeterministicDeduplicatedAndKeepsCompleteTopFive(t *testing.T) {
	report := codemapbench.Report{}
	for document := range map[string]struct{}{"docs/a.md": {}, "docs/b.md": {}} {
		for index := 0; index < 8; index++ {
			report.UnmatchedSuggestions = append(report.UnmatchedSuggestions, codemapbench.Suggestion{
				Link:  codemapbench.Link{Document: document, Target: strings.ReplaceAll(document, ".md", "") + string(rune('a'+index)) + ".go"},
				Score: float64(8 - index), Evidence: []string{"exact_path_mention:source"},
			})
		}
	}
	candidates := CandidatesFromReport(report)
	config := SampleConfig{Seed: "sample-seed", RequestedCount: 12}
	first, err := Sample(candidates, config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Sample(candidates, config)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("sample changed between runs:\n%#v\n%#v", first, second)
	}
	seen := map[string]struct{}{}
	ranksByDocument := map[string]map[int]bool{}
	lowerRankCount := 0
	for _, item := range first {
		key := item.Document + "\x00" + item.Target
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate sample item %s", key)
		}
		seen[key] = struct{}{}
		if ranksByDocument[item.Document] == nil {
			ranksByDocument[item.Document] = map[int]bool{}
		}
		ranksByDocument[item.Document][item.Rank] = true
		if item.Rank > 5 {
			lowerRankCount++
		}
	}
	for document, ranks := range ranksByDocument {
		for rank := 1; rank <= 5; rank++ {
			if !ranks[rank] {
				t.Fatalf("%s is missing top-%d coverage: %#v", document, rank, ranks)
			}
		}
	}
	if lowerRankCount == 0 {
		t.Fatal("sample did not reserve any lower-ranked suggestions")
	}
}

func TestEvaluateCalculatesOverallAcceptanceAndPerDocumentAtK(t *testing.T) {
	benchmark := Benchmark{SchemaVersion: SchemaVersion, Suggestions: []LabeledSuggestion{
		labeled("docs/a.md", "a.go", 1, ValidMissingLink), labeled("docs/a.md", "b.go", 2, PlausibleButUnnecessary), labeled("docs/a.md", "c.go", 3, Incorrect),
		labeled("docs/b.md", "a.go", 1, Incorrect), labeled("docs/b.md", "b.go", 2, ValidMissingLink), labeled("docs/b.md", "c.go", 3, ValidMissingLink),
	}}
	report := codemapbench.Report{}
	for _, item := range benchmark.Suggestions {
		report.UnmatchedSuggestions = append(report.UnmatchedSuggestions, item.Suggestion)
	}
	evaluation, err := Evaluate(benchmark, report)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Overall.OverallPrecision != 0.5 || evaluation.Overall.AcceptancePrecision != 4.0/6.0 {
		t.Fatalf("unexpected overall metrics: %#v", evaluation.Overall)
	}
	if evaluation.PerDocument["docs/a.md"].PrecisionAt1.OverallPrecision != 1 || evaluation.PerDocument["docs/a.md"].PrecisionAt3.OverallPrecision != 1.0/3.0 {
		t.Fatalf("unexpected per-document metrics: %#v", evaluation.PerDocument["docs/a.md"])
	}
	if evaluation.PrecisionAt1 != 0.5 || evaluation.PrecisionAt3 != 0.5 {
		t.Fatalf("unexpected aggregate @k: %#v", evaluation)
	}
}

func labeled(document, target string, rank int, label Label) LabeledSuggestion {
	return LabeledSuggestion{Suggestion: codemapbench.Suggestion{Link: codemapbench.Link{Document: document, Target: target}, Score: float64(10 - rank), Evidence: []string{"exact_path_mention:source"}}, Rank: rank, Area: "test", Subsystem: "test", ScoreBucket: "1-<2", RankBucket: rankBucket(rank), PrimaryEvidenceKind: "exact_path_mention", EvidenceKinds: []string{"exact_path_mention"}, Label: label, Rationale: "reviewed", Audit: AuditMetadata{DocumentRef: "docs:1", DocumentExcerpt: "document", TargetRef: "src:1", TargetExcerpt: "target"}}
}

func TestLoadBenchmarkAllowsUnlabeledTemplateAndRejectsInvalidShape(t *testing.T) {
	benchmark := Benchmark{SchemaVersion: SchemaVersion, Suggestions: []LabeledSuggestion{
		{Suggestion: codemapbench.Suggestion{Link: codemapbench.Link{Document: "docs/a.md", Target: "src/a.go"}}, Rank: 1, Area: "a", Subsystem: "a", ScoreBucket: "1-<2", RankBucket: "1-5", PrimaryEvidenceKind: "exact_path_mention"},
	}}
	var encoded bytes.Buffer
	if err := WriteBenchmark(&encoded, benchmark); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBenchmark(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateLabeledBenchmark(loaded); err == nil {
		t.Fatal("expected unlabeled template to fail labeled validation")
	}
	benchmark.Suggestions = append(benchmark.Suggestions, benchmark.Suggestions[0])
	if err := ValidateBenchmark(benchmark); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestLoadersRejectWrongSchemaAndTrailingJSON(t *testing.T) {
	if _, err := LoadBenchmark(strings.NewReader(`{"schema_version":99}`)); err == nil {
		t.Fatal("expected benchmark schema error")
	}
	if _, err := LoadSuggestionReport(strings.NewReader(`{"schema_version":1,"unmatched_suggestions":[]} {}`)); err == nil {
		t.Fatal("expected trailing report JSON error")
	}
}

func TestCandidatesFromReportUseDeterministicDocumentRankingAndDecoration(t *testing.T) {
	report := codemapbench.Report{UnmatchedSuggestions: []codemapbench.Suggestion{
		{Link: codemapbench.Link{Document: "docs/a.md", Target: "z.go"}, Score: 2, Evidence: []string{"exact_path_mention:x"}},
		{Link: codemapbench.Link{Document: "docs/a.md", Target: "a.go"}, Score: 2, Evidence: []string{"declared_symbol_mention:x"}},
		{Link: codemapbench.Link{Document: "docs/a.md", Target: "m.go"}, Score: 3, Evidence: []string{"test_counterpart:x"}},
	}}
	candidates := CandidatesFromReport(report)
	got := []string{candidates[0].Target, candidates[1].Target, candidates[2].Target}
	if !reflect.DeepEqual(got, []string{"m.go", "a.go", "z.go"}) {
		t.Fatalf("ranked targets = %v", got)
	}
	if candidates[0].Rank != 1 || candidates[0].PrimaryEvidenceKind != "test_counterpart" || candidates[1].ScoreBucket != "2-<8" {
		t.Fatalf("candidate decoration = %#v", candidates)
	}
}

func TestSampleIsDeterministicAndStratifiedAcrossDimensions(t *testing.T) {
	var report codemapbench.Report
	for _, area := range []string{"alpha", "beta", "gamma"} {
		for rank := 1; rank <= 10; rank++ {
			kind := "unique_basename_mention"
			if rank%2 == 0 {
				kind = "declared_symbol_mention"
			}
			report.UnmatchedSuggestions = append(report.UnmatchedSuggestions, codemapbench.Suggestion{
				Link:  codemapbench.Link{Document: "docs/" + area + "/guide.md", Target: "src/" + area + "/target-" + string(rune('a'+rank-1)) + ".go"},
				Score: float64(rank), Evidence: []string{kind + ":source:detail"},
			})
		}
	}
	candidates := CandidatesFromReport(report)
	config := SampleConfig{Seed: "test-seed", RequestedCount: 18}
	first, err := Sample(candidates, config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Sample(candidates, config)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("sample changed between runs")
	}
	areas, ranks, scores, evidence := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, item := range first {
		areas[item.Area] = true
		ranks[item.RankBucket] = true
		scores[item.ScoreBucket] = true
		evidence[item.PrimaryEvidenceKind] = true
		if item.Label != "" || item.Rationale != "" || item.Audit.DocumentRef != "" {
			t.Fatal("sample item is not blank for curation")
		}
	}
	for name, values := range map[string]map[string]bool{"areas": areas, "ranks": ranks, "scores": scores, "evidence": evidence} {
		if len(values) < 2 {
			t.Fatalf("sample did not cover %s: %v", name, values)
		}
	}
}

func TestEvaluateIncludesOverallTopKAndBreakdowns(t *testing.T) {
	items := []LabeledSuggestion{
		labeled("docs/a.md", "a.go", 1, ValidMissingLink),
		labeled("docs/a.md", "b.go", 2, PlausibleButUnnecessary),
		labeled("docs/a.md", "mid.go", 3, Incorrect),
		labeled("docs/a.md", "c.go", 4, Incorrect),
		labeled("docs/b.md", "d.go", 1, Incorrect),
	}
	items[1].PrimaryEvidenceKind = "declared_symbol_mention"
	items[1].ScoreBucket = "2-<8"
	items[2].PrimaryEvidenceKind = "exact_path_mention"
	items[2].ScoreBucket = "2-<8"
	items[3].PrimaryEvidenceKind = "exact_path_mention"
	items[3].ScoreBucket = "2-<8"
	items[4].PrimaryEvidenceKind = "test_counterpart"
	benchmark := Benchmark{SchemaVersion: SchemaVersion, Suggestions: items}
	report := codemapbench.Report{}
	for _, item := range items {
		report.UnmatchedSuggestions = append(report.UnmatchedSuggestions, item.Suggestion)
	}
	evaluation, err := Evaluate(benchmark, report)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Overall.OverallPrecision != 0.2 || evaluation.Overall.AcceptancePrecision != 0.4 {
		t.Fatalf("overall metrics = %#v", evaluation.Overall)
	}
	if evaluation.PrecisionAt1 != 0.5 || evaluation.PrecisionAt3 != 0.25 || evaluation.PrecisionAt5 != 0.2 {
		t.Fatalf("top-k metrics = %v, %v, %v", evaluation.PrecisionAt1, evaluation.PrecisionAt3, evaluation.PrecisionAt5)
	}
	if got := evaluation.ByEvidenceKind["exact_path_mention"]; got.Total != 3 || got.Valid != 1 || got.Accepted != 1 {
		t.Fatalf("evidence breakdown = %#v", got)
	}
	if got := evaluation.ByScoreBucket["1-<2"]; got.Total != 2 || got.Accepted != 1 {
		t.Fatalf("score breakdown = %#v", got)
	}
	if got := evaluation.ByRankBucket["1-5"]; got.Total != 5 || got.Valid != 1 {
		t.Fatalf("rank breakdown = %#v", got)
	}
	if got := evaluation.PerDocument["docs/a.md"]; got.PrecisionAt3.OverallPrecision != 1.0/3.0 || got.PrecisionAt5.OverallPrecision != 0.25 {
		t.Fatalf("document metrics = %#v", got)
	}
}
