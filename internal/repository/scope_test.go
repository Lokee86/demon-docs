package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveInitializedScope(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	child := filepath.Join(docs, "guide")
	if err := os.MkdirAll(filepath.Join(root, DirectoryName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, DirectoryName, ConfigName)
	if err := os.WriteFile(configPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	scope, err := ResolveScope(ScopeOptions{
		WorkingDirectory: child,
		ConfigPath:       configPath,
		ConfiguredRoot:   "docs",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !scope.Initialized || scope.RepositoryRoot != root || scope.DocsRoot != docs || scope.ConfigPath != configPath || scope.IgnorePath != filepath.Join(root, ".docignore") {
		t.Fatalf("unexpected scope: %+v", scope)
	}
}

func TestRepositoryRootOverrideIsRelativeToRepository(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "docs", "guide")
	configPath := filepath.Join(root, DirectoryName, ConfigName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	scope, err := ResolveScope(ScopeOptions{
		WorkingDirectory: child,
		ConfigPath:       configPath,
		ConfiguredRoot:   "other",
		RootOverride:     "docs",
		HasRootOverride:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if scope.DocsRoot != filepath.Join(root, "docs") {
		t.Fatalf("docs root=%q", scope.DocsRoot)
	}
}

func TestResolveScopeRejectsRepositoryEscape(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, DirectoryName, ConfigName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveScope(ScopeOptions{
		WorkingDirectory: root,
		ConfigPath:       configPath,
		ConfiguredRoot:   "../outside",
	}); err == nil {
		t.Fatal("repository escape was accepted")
	}
}

func TestLegacyScopeOwnsItsDocsRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "legacy.toml")
	if err := os.WriteFile(configPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	scope, err := ResolveScope(ScopeOptions{
		WorkingDirectory: root,
		ConfigPath:       configPath,
		ConfiguredRoot:   "notes",
	})
	if err != nil {
		t.Fatal(err)
	}
	if scope.Initialized || scope.RepositoryRoot != filepath.Join(root, "notes") || scope.IgnorePath != filepath.Join(root, "notes", ".docignore") {
		t.Fatalf("unexpected legacy scope: %+v", scope)
	}
}
