package reverseindex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootsUsesConfiguredRepositoryRelativeRoots(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	configured := filepath.Join(repositoryRoot, "services")
	mustWrite(t, filepath.Join(docsRoot, "guide.md"), "# Guide\n")
	if err := os.MkdirAll(filepath.Join(configured, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	roots, err := ResolveRoots(repositoryRoot, docsRoot, repositoryRoot, nil, []string{"services", "services/api"})
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 || roots[0] != configured {
		t.Fatalf("roots=%v", roots)
	}
}

func TestResolveRootsAcceptsRelativeAndAbsoluteCommandPaths(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	cwd := filepath.Join(repositoryRoot, "services")
	api := filepath.Join(cwd, "api")
	worker := filepath.Join(cwd, "worker")
	mustWrite(t, filepath.Join(docsRoot, "guide.md"), "# Guide\n")
	if err := os.MkdirAll(api, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(worker, 0o755); err != nil {
		t.Fatal(err)
	}
	roots, err := ResolveRoots(repositoryRoot, docsRoot, cwd, []string{"api", worker}, []string{"ignored-config-root"})
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 2 || roots[0] != api || roots[1] != worker {
		t.Fatalf("roots=%v", roots)
	}
}

func TestResolveRootsRequiresExplicitScope(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	mustWrite(t, filepath.Join(docsRoot, "guide.md"), "# Guide\n")
	if _, err := ResolveRoots(repositoryRoot, docsRoot, repositoryRoot, nil, nil); err == nil {
		t.Fatal("expected missing reverse-index roots error")
	}
	if _, err := ResolveRoots(repositoryRoot, docsRoot, repositoryRoot, []string{"."}, nil); err == nil {
		t.Fatal("expected repository root rejection")
	}
	if _, err := ResolveRoots(repositoryRoot, docsRoot, repositoryRoot, []string{"docs"}, nil); err == nil {
		t.Fatal("expected docs root rejection")
	}
}
