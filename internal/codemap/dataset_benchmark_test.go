package codemap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkBuildDataset(b *testing.B) {
	repository := b.TempDir()
	const targetCount = 64
	const documentCount = 500
	const entriesPerDocument = 8

	for index := 0; index < targetCount; index++ {
		benchmarkWriteFile(b, repository, fmt.Sprintf("src/target_%02d.go", index), "package src\n\n"+strings.Repeat("var value = 1\n", 128))
	}
	for document := 0; document < documentCount; document++ {
		var source strings.Builder
		source.WriteString("# Document\n\n## Code map\n\n")
		for entry := 0; entry < entriesPerDocument; entry++ {
			target := (document + entry) % targetCount
			fmt.Fprintf(&source, "- `src/target_%02d.go`\n", target)
		}
		benchmarkWriteFile(b, repository, fmt.Sprintf("docs/document_%04d.md", document), source.String())
	}
	docsRoot := filepath.Join(repository, "docs")
	format := DefaultFormat()

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		dataset, err := BuildDataset(repository, docsRoot, format)
		if err != nil {
			b.Fatal(err)
		}
		if len(dataset.Documents) != documentCount || len(dataset.Entries) != documentCount*entriesPerDocument {
			b.Fatalf("unexpected dataset size: %d documents, %d entries", len(dataset.Documents), len(dataset.Entries))
		}
	}
}

func benchmarkWriteFile(b *testing.B, root, relative, contents string) {
	b.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		b.Fatal(err)
	}
}
