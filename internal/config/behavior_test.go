package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadCompletePublicConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc-ledger.toml")
	text := `root = "manual"
index_file = "INDEX.md"
[markers]
prefix = "nav"
[parent_link]
label = "Up"
enabled = false
folder_indexes = true
[sections.files]
heading = "Pages"
[sections.stubs]
name = "Drafts"
[sections.folders]
title = "Areas"
[drafts]
folder = "_drafts"
description_prefix = "Draft: "
[files]
include_patterns = ["**/*.md", "**/*.pdf"]
exclude_patterns = ["private/**/*.md"]
[editable]
parent_index_extensions = [".md", ".mdx"]
[descriptions]
file_template = "File {title}."
folder_template = "Folder {title}."
[watch]
debounce_seconds = 0.25
ignored_dirs = [".git", "vendor"]
ignored_suffixes = [".tmp"]
[aliases]
files = ["Files", "Pages Old"]
folders = ["Folders"]
[template]
managed_sections = ["files"]
include_ownership = false
include_does_not_belong = false
include_related_docs = false
include_notes = false
`
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Root != "manual" || c.IndexFile != "INDEX.md" || c.Files.IndexFile != "INDEX.md" || c.Markers.Prefix != "nav" {
		t.Fatalf("top-level: %+v", c)
	}
	if c.ParentLink.Label != "Up" || !c.ParentLink.FolderIndexes || c.ParentLink.IndexedFiles {
		t.Fatalf("parent: %+v", c.ParentLink)
	}
	if c.Sections.FilesHeading != "Pages" || c.Sections.StubsHeading != "Drafts" || c.Sections.FoldersHeading != "Areas" {
		t.Fatalf("sections: %+v", c.Sections)
	}
	if c.Draft.Folder != "_drafts" || c.Draft.DescriptionPrefix != "Draft: " {
		t.Fatal(c.Draft)
	}
	if !reflect.DeepEqual(c.Files.IncludePatterns, []string{"**/*.md", "**/*.pdf"}) || !reflect.DeepEqual(c.Files.ExcludePatterns, []string{"private/**/*.md"}) || !reflect.DeepEqual(c.Files.EditableParentIndexExtensions, []string{".md", ".mdx"}) {
		t.Fatal(c.Files)
	}
	if c.Description.FileTemplate != "File {title}." || c.Description.FolderTemplate != "Folder {title}." || c.Watch.DebounceSeconds != 0.25 {
		t.Fatal("description/watch")
	}
	if c.Template.IncludeOwnership || c.Template.IncludeDoesNotBelong || c.Template.IncludeRelatedDocs || c.Template.IncludeNotes || !reflect.DeepEqual(c.Template.ManagedSections, []string{"files"}) {
		t.Fatal(c.Template)
	}
}

func TestParentEnabledAliasAndExplicitKeys(t *testing.T) {
	for _, tc := range []struct {
		text         string
		folder, file bool
	}{{"[parent_link]\nenabled = true\n", true, true}, {"[parent_link]\nenabled = false\nfolder_indexes = true\n", true, false}, {"[parent_link]\nenabled = true\nindexed_files = false\n", true, false}} {
		path := filepath.Join(t.TempDir(), "c.toml")
		if err := os.WriteFile(path, []byte(tc.text), 0o644); err != nil {
			t.Fatal(err)
		}
		c, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if c.ParentLink.FolderIndexes != tc.folder || c.ParentLink.IndexedFiles != tc.file {
			t.Fatalf("%q => %+v", tc.text, c.ParentLink)
		}
	}
}

func TestConfigSelectionPrecedence(t *testing.T) {
	dir := t.TempDir()
	dot := filepath.Join(dir, ".doc-ledger.toml")
	plain := filepath.Join(dir, "doc-ledger.toml")
	global := filepath.Join(dir, "xdg", "doc-ledger", "config.toml")
	if err := os.WriteFile(plain, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(global), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	env := func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return filepath.Join(dir, "xdg")
		}
		return ""
	}
	if got := Select(dir, "explicit.toml", false, false, env, dir); got != "explicit.toml" {
		t.Fatal(got)
	}
	if got := Select(dir, "", false, false, env, dir); got != plain {
		t.Fatal(got)
	}
	if err := os.WriteFile(dot, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Select(dir, "", false, false, env, dir); got != dot {
		t.Fatal(got)
	}
	if got := Select(dir, "", true, false, env, dir); got != global {
		t.Fatal(got)
	}
	if got := Select(dir, "", true, true, env, dir); got != "" {
		t.Fatal(got)
	}
}

func TestEditableExtensionsAreExact(t *testing.T) {
	c := Default()
	if !IsParentEditable("page.md", c) || IsParentEditable("page.MD", c) || IsParentEditable("image.png", c) {
		t.Fatal("extension matching changed")
	}
}
