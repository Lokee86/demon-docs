package codemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildDatasetScansAndResolvesCodeMaps(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/main.go", "package main\n")
	writeFixture(t, repository, "src/runtime/keep.txt", "runtime\n")
	writeFixture(t, repository, "src/payload_a.go", "package src\n")
	writeFixture(t, repository, "src/payload_b.go", "package src\n")
	writeFixture(t, repository, "docs/guide.md", `# Guide

## Code map

### Runtime

* `+"`src/main.go`"+` — owns startup
* `+"`src/runtime/`"+`
* `+"`src/missing.go`"+`
* `+"`src/main.go#Run`"+` — entry symbol
* `+"`src/payload_*.go`"+` — generated payloads
* TODO: add another path

## Tests

* `+"`ignored/outside-map.go`"+`
`)
	writeFixture(t, repository, "docs/ignored.md", "## Code map\n\n* `src/ignored.go`\n")
	writeFixture(t, repository, ".docignore", "docs/ignored.md\n")

	dataset, err := BuildDataset(repository, filepath.Join(repository, "docs"), DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(dataset.Documents) != 1 || dataset.Documents[0].Path != "docs/guide.md" {
		t.Fatalf("unexpected documents: %#v", dataset.Documents)
	}
	if dataset.Documents[0].EntryCount != 5 || dataset.Documents[0].DiagnosticCount != 1 {
		t.Fatalf("unexpected document counts: %#v", dataset.Documents[0])
	}
	if len(dataset.Entries) != 5 {
		t.Fatalf("got %d entries: %#v", len(dataset.Entries), dataset.Entries)
	}
	assertResolution(t, dataset, "src/main.go", ResolutionResolved, true)
	assertResolution(t, dataset, "src/runtime/", ResolutionResolved, true)
	assertResolution(t, dataset, "src/missing.go", ResolutionMissing, false)
	assertResolution(t, dataset, "src/main.go#Run", ResolutionSymbolUnverified, true)
	assertResolution(t, dataset, "src/payload_*.go", ResolutionPatternResolved, true)
	for _, entry := range dataset.Entries {
		if entry.Entry.Target == "src/payload_*.go" && (len(entry.Resolution.Matches) != 2 || entry.Resolution.Matches[0].Path != "src/payload_a.go" || entry.Resolution.Matches[1].Path != "src/payload_b.go") {
			t.Fatalf("unexpected pattern matches: %#v", entry.Resolution.Matches)
		}
	}
	if len(dataset.Diagnostics) != 1 || dataset.Diagnostics[0].Code != "unparsed_entry" {
		t.Fatalf("unexpected diagnostics: %#v", dataset.Diagnostics)
	}
	if dataset.Documents[0].SHA256 == "" || dataset.Entries[0].Resolution.SHA256 == "" {
		t.Fatal("expected deterministic content hashes")
	}
}

func TestBuildDatasetSupportsDocumentRelativeTargets(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "docs/code/example.go", "package code\n")
	writeFixture(t, repository, "docs/guide.md", "## Implementation map\n\n- `code/example.go`\n")
	format := Format{SectionHeadings: []string{"implementation map"}, TargetBase: TargetBaseDocument}

	dataset, err := BuildDataset(repository, filepath.Join(repository, "docs"), format)
	if err != nil {
		t.Fatal(err)
	}
	if len(dataset.Entries) != 1 {
		t.Fatalf("unexpected entries: %#v", dataset.Entries)
	}
	entry := dataset.Entries[0]
	if entry.Resolution.ResolvedPath != "docs/code/example.go" || entry.Resolution.Status != ResolutionResolved {
		t.Fatalf("unexpected resolution: %#v", entry.Resolution)
	}
}

func TestDatasetJSONIsStableAndExportable(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/b.go", "package src\n")
	writeFixture(t, repository, "docs/b.md", "## Code map\n\n- `src/b.go`\n")
	writeFixture(t, repository, "docs/a.md", "# No map\n")

	first, err := BuildDataset(repository, filepath.Join(repository, "docs"), DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildDataset(repository, filepath.Join(repository, "docs"), DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := MarshalDataset(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := MarshalDataset(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("dataset output changed for identical inputs")
	}
	if firstJSON[len(firstJSON)-1] != '\n' {
		t.Fatal("dataset JSON should end with a newline")
	}
	output := filepath.Join(repository, "out", "codemaps.json")
	if err := ExportDataset(output, first); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != string(firstJSON) {
		t.Fatal("exported dataset differs from marshaled dataset")
	}
}

func assertResolution(t *testing.T, dataset Dataset, target string, status ResolutionStatus, exists bool) {
	t.Helper()
	for _, entry := range dataset.Entries {
		if entry.Entry.Target == target {
			if entry.Resolution.Status != status || entry.Resolution.Exists != exists {
				t.Fatalf("target %s resolution = %#v", target, entry.Resolution)
			}
			return
		}
	}
	t.Fatalf("target %s not found", target)
}

func writeFixture(t *testing.T, root, relative, contents string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
