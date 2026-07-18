package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultsAndAliases(t *testing.T) {
	c := Default()
	if c.Root != "docs" || c.IndexFile != "README.md" || c.Markers.Prefix != "doc-ledger" || !c.ParentLink.FolderIndexes || c.ParentLink.IndexedFiles {
		t.Fatalf("unexpected defaults: %+v", c)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	text := `index_file = "!README.md"
[parent_link]
enabled = false
folder_indexes = true
[sections.files]
title = "Pages"
[editable]
extensions = [".md", ".mdx"]
`
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.IndexFile != "!README.md" || c.Files.IndexFile != "!README.md" || !c.ParentLink.FolderIndexes || c.ParentLink.IndexedFiles || c.Sections.FilesHeading != "Pages" || !reflect.DeepEqual(c.Files.EditableParentIndexExtensions, []string{".md", ".mdx"}) {
		t.Fatalf("aliases not preserved: %+v", c)
	}
}
func TestSelectionIsCurrentDirectoryOnly(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, ".doc-ledger.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if LocalPath(child) != "" {
		t.Fatal("local lookup searched parent")
	}
	if Discover(child) == "" {
		t.Fatal("legacy discovery should still search parent")
	}
}
func TestStarterConfigLoads(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"legacy.toml": StarterText(),
		"repo.toml":   RepositoryStarterText("manual"),
	} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		loaded, err := Load(p)
		if err != nil {
			t.Fatal(err)
		}
		if name == "repo.toml" && loaded.Root != "manual" {
			t.Fatalf("docs_root not loaded: %+v", loaded)
		}
	}
}
