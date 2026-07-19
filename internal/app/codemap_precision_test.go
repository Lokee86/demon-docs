package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodemapPrecisionEvaluatesJSONBenchmark(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		benchmark := `{"schema_version":1,"suggestions":[{"document":"docs/a.md","target":"src/a.go","score":2,"evidence":["exact_path_mention:src/a.go"],"rank":1,"area":"test","subsystem":"test","score_bucket":"2-<8","rank_bucket":"1-5","primary_evidence_kind":"exact_path_mention","evidence_kinds":["exact_path_mention"],"label":"valid_missing_link","rationale":"reviewed","audit":{"document_ref":"docs/a.md:1","document_excerpt":"doc","target_ref":"src/a.go:1","target_excerpt":"code"}}]}`
		report := `{"schema_version":1,"unmatched_suggestions":[{"document":"docs/a.md","target":"src/a.go","score":2,"evidence":["exact_path_mention:src/a.go"]}]}`
		benchmarkPath, reportPath := filepath.Join(root, "benchmark.json"), filepath.Join(root, "report.json")
		if err := os.WriteFile(benchmarkPath, []byte(benchmark), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		code := Run(context.Background(), []string{"codemap", "precision", "--benchmark", benchmarkPath, "--suggestions", reportPath}, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "Overall precision: 100.00%") {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
}

func TestCodemapPrecisionSamplesUnlabeledTemplate(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		reportPath := filepath.Join(root, "report.json")
		report := `{"schema_version":1,"unmatched_suggestions":[{"document":"docs/a.md","target":"src/a.go","score":2,"evidence":["exact_path_mention:src/a.go"]}]}`
		if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		code := Run(context.Background(), []string{"codemap", "precision", "sample", "--suggestions", reportPath, "--count", "1"}, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		var sample struct {
			Sampling struct {
				SourceReport string `json:"source_report"`
			} `json:"sampling"`
			Suggestions []struct {
				Label     string `json:"label"`
				Rationale string `json:"rationale"`
			} `json:"suggestions"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &sample); err != nil {
			t.Fatal(err)
		}
		if sample.Sampling.SourceReport != "report.json" {
			t.Fatalf("source report = %q, want report.json", sample.Sampling.SourceReport)
		}
		if len(sample.Suggestions) != 1 || sample.Suggestions[0].Label != "" || sample.Suggestions[0].Rationale != "" {
			t.Fatalf("unexpected sample: %#v", sample)
		}
	})
}
