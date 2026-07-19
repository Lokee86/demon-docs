package links

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkInitialIndexing(b *testing.B) {
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		b.StopTimer()
		root := filepath.Join(b.TempDir(), fmt.Sprintf("repo-%d", iteration))
		createBenchmarkRepository(b, root, 250)
		b.StartTimer()
		plan, err := Reconcile(root)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := ApplyAndSave(&plan); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSingleFileIncrementalUpdate(b *testing.B) {
	root := b.TempDir()
	createBenchmarkRepository(b, root, 500)
	baseline, err := Reconcile(root)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		b.Fatal(err)
	}
	changedPath := filepath.Join(root, "docs", "page-0250.md")
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		target := "asset-a.bin"
		if iteration%2 == 1 {
			target = "asset-b.bin"
		}
		content := fmt.Sprintf("# Changed\n\n[target](../%s)\n", target)
		if err := os.WriteFile(changedPath, []byte(content), 0o644); err != nil {
			b.Fatal(err)
		}
		plan, err := Reconcile(root)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := ApplyAndSave(&plan); err != nil {
			b.Fatal(err)
		}
	}
}

func createBenchmarkRepository(tb testing.TB, root string, count int) {
	tb.Helper()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		tb.Fatal(err)
	}
	for _, name := range []string{"asset-a.bin", "asset-b.bin"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
	for index := 0; index < count; index++ {
		path := filepath.Join(root, "docs", fmt.Sprintf("page-%04d.md", index))
		content := fmt.Sprintf("# Page %d\n\n[target](../asset-a.bin)\n", index)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
}
