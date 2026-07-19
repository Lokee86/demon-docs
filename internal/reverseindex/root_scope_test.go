package reverseindex

import (
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
)

func TestBuildSkipsRepositoryRootAndControlDirectoryTargets(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	mustWrite(t, filepath.Join(docsRoot, "scope.md"), "# Scope\n\n## Code map\n\n- `root.go`\n- `.ddocs/config.go`\n- `src/feature.go`\n")
	mustWrite(t, filepath.Join(repositoryRoot, "root.go"), "package root\n")
	mustWrite(t, filepath.Join(repositoryRoot, ".ddocs", "config.go"), "package control\n")
	mustWrite(t, filepath.Join(repositoryRoot, "src", "feature.go"), "package src\n")

	plan, err := Build(repositoryRoot, docsRoot, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 {
		t.Fatalf("expected only src index, got %#v", plan.Updates)
	}
	if got := filepath.Dir(plan.Updates[0].Path); got != filepath.Join(repositoryRoot, "src") {
		t.Fatalf("index folder=%s", got)
	}
}
