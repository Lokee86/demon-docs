package codemapcorpus

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func BenchmarkBuildCorpus(b *testing.B) {
	root := b.TempDir()
	const goFileCount = 384
	const gdscriptFileCount = 128
	const documentCount = 96

	benchmarkWriteCorpusFile(b, root, "go.mod", "module example.com/corpusbench\n\ngo 1.26\n")
	for index := 0; index < goFileCount; index++ {
		next := (index + 1) % goFileCount
		file := fmt.Sprintf("services/pkg%03d/runtime.go", index)
		contents := fmt.Sprintf(
			"package pkg%03d\n\nimport _ \"example.com/corpusbench/services/pkg%03d\"\n\ntype RuntimeNode struct{}\n",
			index,
			next,
		)
		benchmarkWriteCorpusFile(b, root, file, contents)
	}

	benchmarkWriteCorpusFile(b, root, "client/project.godot", "[application]\n")
	for index := 0; index < gdscriptFileCount; index++ {
		next := (index + 1) % gdscriptFileCount
		file := fmt.Sprintf("client/scripts/runtime_%03d.gd", index)
		contents := fmt.Sprintf(
			"class_name RuntimeNode\nconst Next = preload(\"res://scripts/runtime_%03d.gd\")\nfunc start_runtime():\n\tpass\n",
			next,
		)
		benchmarkWriteCorpusFile(b, root, file, contents)
	}

	dataset := codemap.Dataset{}
	for index := 0; index < documentCount; index++ {
		next := (index + 1) % documentCount
		documentPath := fmt.Sprintf("docs/document_%03d.md", index)
		targetPath := fmt.Sprintf("services/pkg%03d/runtime.go", index%goFileCount)
		benchmarkWriteCorpusFile(
			b,
			root,
			documentPath,
			fmt.Sprintf("# Document %d\n\nSee [next](document_%03d.md).\n", index, next),
		)
		dataset.Documents = append(dataset.Documents, codemap.DocumentRecord{Path: documentPath})
		dataset.Entries = append(dataset.Entries, codemap.DatasetEntry{
			Entry: codemap.Entry{DocumentPath: documentPath},
			Resolution: codemap.TargetRecord{
				Status:       codemap.ResolutionResolved,
				ResolvedPath: targetPath,
			},
		})
	}

	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		corpus, err := Build(root, dataset, Options{MaxCommits: 1, MaxPathsPerCommit: 20})
		if err != nil {
			b.Fatal(err)
		}
		if len(corpus.Documents) != documentCount || len(corpus.DependencyEdges) != goFileCount+gdscriptFileCount {
			b.Fatalf(
				"unexpected corpus size: %d documents, %d dependencies",
				len(corpus.Documents),
				len(corpus.DependencyEdges),
			)
		}
	}
}

func benchmarkWriteCorpusFile(b *testing.B, root, relative, contents string) {
	b.Helper()
	filePath := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte(contents), 0o644); err != nil {
		b.Fatal(err)
	}
}
