package codemapbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMarshalJSONReportIsCanonicalAndVersioned(t *testing.T) {
	first := exportFixtureReport()
	second := exportFixtureReport()
	reverseLinks(second.KnownLinks)
	reverseLinks(second.HiddenLinks)
	reverseSuggestions(second.RecoveredSuggestions)
	reverseSuggestions(second.UnmatchedSuggestions)
	reverseStrings(second.RecoveredSuggestions[0].Evidence)

	before := append([]string(nil), first.RecoveredSuggestions[0].Evidence...)
	firstJSON, err := MarshalJSONReport(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := MarshalJSONReport(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("canonical JSON changed with input ordering:\n%s\n%s", firstJSON, secondJSON)
	}
	if !reflect.DeepEqual(first.RecoveredSuggestions[0].Evidence, before) {
		t.Fatalf("export mutated report evidence: %#v", first.RecoveredSuggestions[0].Evidence)
	}

	var decoded map[string]any
	if err := json.Unmarshal(firstJSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["schema_version"] != float64(ReportSchemaVersion) {
		t.Fatalf("unexpected schema version: %#v", decoded["schema_version"])
	}
	if decoded["seed"] != "report-seed" {
		t.Fatalf("report fields were not emitted at top level: %#v", decoded)
	}
	if !strings.HasSuffix(string(firstJSON), "\n") {
		t.Fatal("JSON report must end with a newline")
	}
}

func TestFormatTextReportIncludesScoresEvidenceAndClassifications(t *testing.T) {
	got := FormatTextReport(exportFixtureReport())
	want := `Codemap benchmark report
Schema: 1
Seed: report-seed

Known links: 3
Visible links: 1
Hidden links: 2
Recovered links: 1
Missed links: 1
Unmatched suggestions: 1
Already-linked suggestions: 1
Duplicate suggestions: 1
Invalid suggestions: 1
Raw suggestions: 5
Unique suggestions: 3
Precision: 33.33%
Recall: 50.00%

Recovered:
- docs/a.md -> src/a.go (score 0.9000)
  evidence: direct mention
  evidence: git co-change

Missed:
- docs/b.md -> src/b.go

Unmatched:
- docs/a.md -> src/extra.go (score 0.4000)
  evidence: sibling file

Already linked:
- docs/c.md -> src/c.go (score 0.3000)

Duplicates:
- docs/a.md -> src/a.go (score 0.2000)

Invalid:
- suggestion 4: link document cannot be empty ( -> src/invalid.go)
`
	if got != want {
		t.Fatalf("unexpected text report:\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}

func exportFixtureReport() Report {
	return Report{
		Seed: "report-seed",
		KnownLinks: []Link{
			{Document: "docs/c.md", Target: "src/c.go"},
			{Document: "docs/a.md", Target: "src/a.go"},
			{Document: "docs/b.md", Target: "src/b.go"},
		},
		VisibleLinks: []Link{{Document: "docs/c.md", Target: "src/c.go"}},
		HiddenLinks: []Link{
			{Document: "docs/b.md", Target: "src/b.go"},
			{Document: "docs/a.md", Target: "src/a.go"},
		},
		RecoveredLinks: []Link{{Document: "docs/a.md", Target: "src/a.go"}},
		RecoveredSuggestions: []Suggestion{{
			Link:     Link{Document: "docs/a.md", Target: "src/a.go"},
			Score:    0.9,
			Evidence: []string{"git co-change", "direct mention"},
		}},
		MissedLinks: []Link{{Document: "docs/b.md", Target: "src/b.go"}},
		UnmatchedSuggestions: []Suggestion{{
			Link:     Link{Document: "docs/a.md", Target: "src/extra.go"},
			Score:    0.4,
			Evidence: []string{"sibling file"},
		}},
		AlreadyLinked: []Suggestion{{
			Link:  Link{Document: "docs/c.md", Target: "src/c.go"},
			Score: 0.3,
		}},
		DuplicateSuggestions: []Suggestion{{
			Link:  Link{Document: "docs/a.md", Target: "src/a.go"},
			Score: 0.2,
		}},
		InvalidSuggestions: []InvalidSuggestion{{
			Index:      4,
			Suggestion: Suggestion{Link: Link{Target: "src/invalid.go"}},
			Reason:     "link document cannot be empty",
		}},
		RawSuggestionCount:    5,
		UniqueSuggestionCount: 3,
		Precision:             1.0 / 3.0,
		Recall:                0.5,
	}
}

func reverseLinks(values []Link) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseSuggestions(values []Suggestion) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
