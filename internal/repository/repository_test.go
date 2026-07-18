package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeAndDiscoverFromChild(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "docs", "guide")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	path, err := Initialize(root, "docs_root = \"docs\"\n")
	if err != nil {
		t.Fatal(err)
	}
	location, ok := Discover(child)
	if !ok || location.Root != root || location.ConfigPath != path {
		t.Fatalf("location=%+v ok=%t", location, ok)
	}
	if found, ok := FindMarker(child); !ok || found != root {
		t.Fatalf("marker=%q ok=%t", found, ok)
	}
	if _, err := Initialize(root, "docs_root = \"docs\"\n"); err == nil {
		t.Fatal("reinitialization unexpectedly succeeded")
	}
}

func TestResolveDocsRoot(t *testing.T) {
	root := t.TempDir()
	relative, absolute, err := ResolveDocsRoot(root, "docs")
	if err != nil {
		t.Fatal(err)
	}
	if relative != "docs" || absolute != filepath.Join(root, "docs") {
		t.Fatalf("relative=%q absolute=%q", relative, absolute)
	}
	if _, _, err := ResolveDocsRoot(root, filepath.Join(root, "..", "outside")); err == nil {
		t.Fatal("outside docs root accepted")
	}
}

func TestRootForConfig(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, DirectoryName, ConfigName)
	if got, ok := RootForConfig(path); !ok || got != root {
		t.Fatalf("root=%q ok=%t", got, ok)
	}
	if _, ok := RootForConfig(filepath.Join(root, "config.toml")); ok {
		t.Fatal("ordinary config treated as repository config")
	}
}
