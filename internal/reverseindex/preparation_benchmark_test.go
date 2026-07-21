package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
)

func BenchmarkReverseIndexPreparation(b *testing.B) {
	repositoryRoot := b.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "services")
	const folderCount = 128
	const filesPerFolder = 4

	for folderIndex := 0; folderIndex < folderCount; folderIndex++ {
		folderName := fmt.Sprintf("pkg%03d", folderIndex)
		folder := filepath.Join(codeRoot, folderName)
		for fileIndex := 0; fileIndex < filesPerFolder; fileIndex++ {
			path := filepath.Join(folder, fmt.Sprintf("runtime_%02d.go", fileIndex))
			benchmarkWriteReverseFile(b, path, "package "+folderName+"\n")
		}
		document := filepath.Join(docsRoot, fmt.Sprintf("package-%03d.md", folderIndex))
		target := filepath.ToSlash(filepath.Join("services", folderName, "runtime_00.go"))
		benchmarkWriteReverseFile(b, document, fmt.Sprintf("# Package %d\n\n## Code map\n\n- `%s`\n", folderIndex, target))
	}

	cfg := config.Default()
	format := codemap.DefaultFormat()
	plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, cfg, format)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := Apply(repositoryRoot, plan); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, cfg, format)
		if err != nil {
			b.Fatal(err)
		}
		if len(plan.Updates) != 0 || plan.IndexCount != folderCount {
			b.Fatalf("unexpected stable plan: %d updates, %d indexes, diagnostics=%s", len(plan.Updates), plan.IndexCount, strings.Join(plan.Diagnostics, "; "))
		}
	}
}

func benchmarkWriteReverseFile(b *testing.B, path, contents string) {
	b.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		b.Fatal(err)
	}
}
