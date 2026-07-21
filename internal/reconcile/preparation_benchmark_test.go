package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func BenchmarkTreePreparation(b *testing.B) {
	root := b.TempDir()
	cfg := config.Default()
	const folderCount = 128
	const filesPerFolder = 4

	for folderIndex := 0; folderIndex < folderCount; folderIndex++ {
		folder := filepath.Join(root, fmt.Sprintf("group-%03d", folderIndex))
		for fileIndex := 0; fileIndex < filesPerFolder; fileIndex++ {
			path := filepath.Join(folder, fmt.Sprintf("document-%02d.md", fileIndex))
			benchmarkWriteIndexFile(b, path, fmt.Sprintf("# Document %d %d\n\nBody text.\n", folderIndex, fileIndex))
		}
	}
	if _, _, err := ConvergeWithin(root, root, cfg); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		result, err := Tree(root, cfg)
		if err != nil {
			b.Fatal(err)
		}
		if len(result.Updates) != 0 {
			b.Fatalf("stable tree produced %d updates", len(result.Updates))
		}
	}
}

func benchmarkWriteIndexFile(b *testing.B, path, contents string) {
	b.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		b.Fatal(err)
	}
}
