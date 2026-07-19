package codemapcorpus

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func TestBuildProvidesBenchmarkCorpusInputs(t *testing.T) {
	root := t.TempDir()
	files := fixtureFiles()
	writeFiles(t, root, files)
	tracked := make([]string, 0, len(files))
	for file := range files {
		if file != ".docignore" && file != "vendor/ignored.go" && file != ".ddocs/private.go" {
			tracked = append(tracked, file)
		}
	}
	repository := initializeHistory(t, root, tracked)

	writeFiles(t, root, map[string]string{
		"docs/a.md":     files["docs/a.md"] + "\nChanged.\n",
		"server/a/a.go": files["server/a/a.go"] + "\n// changed\n",
	})
	worktree, err := repository.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"docs/a.md", "server/a/a.go"} {
		if _, err := worktree.Add(file); err != nil {
			t.Fatal(err)
		}
	}
	commit(t, worktree, "change doc and owner", time.Unix(2, 0))

	corpus, err := Build(root, fixtureDataset(), Options{MaxCommits: 20, MaxPathsPerCommit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if contains(corpus.RepositoryFiles, ".ddocs/private.go") || contains(corpus.RepositoryFiles, "vendor/ignored.go") {
		t.Fatalf("ignored files leaked into corpus: %v", corpus.RepositoryFiles)
	}
	if !contains(corpus.RepositoryFiles, "server/a/a.go") {
		t.Fatal("repository files omitted server/a/a.go")
	}
	if corpus.Documents["docs/a.md"] == "" {
		t.Fatal("document text was not loaded")
	}
	if got := corpus.KnownTargets("docs/a.md"); !reflect.DeepEqual(got, []string{"server/a/a.go"}) {
		t.Fatalf("known targets = %v", got)
	}
	if len(corpus.Commits) != 1 || !reflect.DeepEqual(corpus.Commits[0].Paths, []string{"docs/a.md", "server/a/a.go"}) {
		t.Fatalf("commits = %#v", corpus.Commits)
	}
	related := corpus.RelatedDocuments["docs/a.md"]
	if len(related) != 1 || related[0].Path != "docs/b.md" || !reflect.DeepEqual(related[0].Targets, []string{"server/b/b.go"}) {
		t.Fatalf("related documents = %#v", related)
	}

	input, err := corpus.Input("docs/a.md", []string{"visible/only.go"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input.ExistingTargets, []string{"visible/only.go"}) {
		t.Fatalf("input ignored caller holdout targets: %v", input.ExistingTargets)
	}
}

func fixtureDataset() codemap.Dataset {
	return codemap.Dataset{
		Documents: []codemap.DocumentRecord{{Path: "docs/a.md"}, {Path: "docs/b.md"}},
		Entries: []codemap.DatasetEntry{
			{Entry: codemap.Entry{DocumentPath: "docs/a.md"}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved, ResolvedPath: "server/a/a.go"}},
			{Entry: codemap.Entry{DocumentPath: "docs/b.md"}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved, ResolvedPath: "server/b/b.go"}},
		},
	}
}

func fixtureFiles() map[string]string {
	return map[string]string{
		".docignore":            "vendor/\n",
		".ddocs/private.go":     "package hidden\n",
		"vendor/ignored.go":     "package vendor\n",
		"docs/a.md":             "# A\n\nSee [B](b.md).\n",
		"docs/b.md":             "# B\n",
		"go.mod":                "module example.com/demo\n\ngo 1.26\n",
		"server/a/a.go":         "package a\nimport _ \"example.com/demo/server/b\"\n",
		"server/b/b.go":         "package b\n",
		"client/project.godot":  "[application]\n",
		"client/main.gd":        "const Tool = preload(\"res://shared/tool.gd\")\n",
		"client/shared/tool.gd": "extends Node\n",
		"web/a.ts":              "import { b } from './b'\n",
		"web/b.ts":              "export const b = 1\n",
		"api/a.rb":              "require_relative 'b'\n",
		"api/b.rb":              "B = true\n",
		"tools/pkg/a.py":        "from .b import Thing\n",
		"tools/pkg/b.py":        "class Thing: pass\n",
	}
}

func TestBuildRejectsMissingDatasetDocument(t *testing.T) {
	root := t.TempDir()
	_, err := Build(root, codemap.Dataset{Documents: []codemap.DocumentRecord{{Path: "docs/missing.md"}}}, Options{})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing document error, got %v", err)
	}
}
