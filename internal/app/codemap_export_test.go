package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func TestCodemapExportWritesDatasetToStdout(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		writeTestFile(t, filepath.Join(root, ".ddocs", "config.toml"), "docs_root = \"docs\"\n")
		writeTestFile(t, filepath.Join(root, "src", "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(root, "docs", "guide.md"), "## Code map\n\n- `src/main.go` — startup\n")

		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"codemap", "export"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stderr=%q", code, stderr.String())
		}
		var dataset codemap.Dataset
		if err := json.Unmarshal(stdout.Bytes(), &dataset); err != nil {
			t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
		}
		if len(dataset.Entries) != 1 || dataset.Entries[0].Entry.Target != "src/main.go" {
			t.Fatalf("unexpected dataset: %#v", dataset)
		}
		if dataset.Entries[0].Resolution.Status != codemap.ResolutionResolved {
			t.Fatalf("unexpected resolution: %#v", dataset.Entries[0].Resolution)
		}
	})
}

func TestCodemapExportSupportsOutputAndCustomHeading(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		writeTestFile(t, filepath.Join(root, ".ddocs", "config.toml"), "docs_root = \"docs\"\n")
		writeTestFile(t, filepath.Join(root, "docs", "code", "thing.go"), "package code\n")
		writeTestFile(t, filepath.Join(root, "docs", "guide.md"), "## Implementation map\n\n- `code/thing.go`\n")
		output := filepath.Join(root, "results", "codemaps.json")

		var stdout, stderr bytes.Buffer
		args := []string{"codemap", "export", "--heading", "Implementation map", "--target-base", "document", "--output", output}
		if code := Run(context.Background(), args, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stderr=%q", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "exported 1 codemap link(s)") {
			t.Fatalf("unexpected stdout: %q", stdout.String())
		}
		contents, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var dataset codemap.Dataset
		if err := json.Unmarshal(contents, &dataset); err != nil {
			t.Fatal(err)
		}
		if dataset.Entries[0].Resolution.ResolvedPath != "docs/code/thing.go" {
			t.Fatalf("unexpected resolution: %#v", dataset.Entries[0].Resolution)
		}
	})
}
