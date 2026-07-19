package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemapbench"
)

func TestBenchmarkEngineRecoversEvidenceOutsideAuthoredMap(t *testing.T) {
	root := t.TempDir()
	writeBenchmarkFixture(t, root, "Implementation lives in `src/runtime.go`.\n\n## Code map\n\n- `src/runtime.go`\n")

	result, err := (benchmarkEngine{}).Run(context.Background(), codemapBenchmarkOptions{
		RepositoryRoot: root,
		HoldoutCount:   1,
		Format:         "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	report := decodeBenchmarkResult(t, result.Payload)
	if report.Recall != 1 || len(report.RecoveredLinks) != 1 {
		t.Fatalf("expected the prose evidence to recover the link: %#v", report)
	}
}

func TestBenchmarkEngineDoesNotReadAnswersFromAuthoredMap(t *testing.T) {
	root := t.TempDir()
	writeBenchmarkFixture(t, root, "Runtime behavior.\n\n## Code map\n\n- `src/runtime.go`\n")

	result, err := (benchmarkEngine{}).Run(context.Background(), codemapBenchmarkOptions{
		RepositoryRoot: root,
		HoldoutCount:   1,
		Format:         "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	report := decodeBenchmarkResult(t, result.Payload)
	if report.Recall != 0 || len(report.RecoveredLinks) != 0 || len(report.MissedLinks) != 1 {
		t.Fatalf("authored map leaked into evidence: %#v", report)
	}
}

func writeBenchmarkFixture(t *testing.T, root, document string) {
	t.Helper()
	for name, contents := range map[string]string{
		"docs/runtime.md": document,
		"src/runtime.go":  "package runtime\n",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func decodeBenchmarkResult(t *testing.T, payload []byte) codemapbench.Report {
	t.Helper()
	var envelope struct {
		SchemaVersion int `json:"schema_version"`
		codemapbench.Report
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.SchemaVersion != codemapbench.ReportSchemaVersion {
		t.Fatalf("schema version = %d", envelope.SchemaVersion)
	}
	return envelope.Report
}
