package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultsAndAliases(t *testing.T) {
	c := Default()
	if c.Root != "docs" || c.IndexFile != "README.md" || c.Markers.Prefix != "doc-ledger" || !c.ParentLink.FolderIndexes || c.ParentLink.IndexedFiles || !c.Demon.Run {
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

func TestCodemapHeadingsLoadFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[codemap]\nheadings = [\"Implementation map\", \"Source map\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Codemap.Headings, []string{"Implementation map", "Source map"}) {
		t.Fatalf("codemap headings not loaded: %+v", loaded.Codemap)
	}
	if !strings.Contains(StarterText(), "[codemap]\nheadings =") {
		t.Fatal("starter config omitted codemap headings")
	}
}

func TestReverseIndexRootsLoadWithFoldersCompatibilityAlias(t *testing.T) {
	dir := t.TempDir()
	for name, section := range map[string]string{
		"roots.toml":   "[reverse_index]\nroots = [\"client\", \"services/game-server\"]\n",
		"folders.toml": "[reverse_index]\nfolders = [\"client\"]\n",
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(section), 0o644); err != nil {
			t.Fatal(err)
		}
		loaded, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(loaded.ReverseIndex.Roots) == 0 || loaded.ReverseIndex.Roots[0] != "client" {
			t.Fatalf("reverse-index roots not loaded from %s: %+v", name, loaded.ReverseIndex)
		}
	}
	if !strings.Contains(StarterText(), "[reverse_index]\nroots = []") {
		t.Fatal("starter config omitted reverse-index roots")
	}
}

func TestDemonRunDefaultsAndAtomicEditPreserveText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	original := "# keep this comment\nunknown = \"value\"\n\n[demon]\n# preserve me\nrun = true # trailing\n\n[watch]\ndebounce_seconds = 0.5\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDemonRun(path, false); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, want := range []string{"# keep this comment", "unknown = \"value\"", "# preserve me", "run = false # trailing", "debounce_seconds = 0.5"} {
		if !strings.Contains(text, want) {
			t.Fatalf("updated config missing %q: %s", want, text)
		}
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Demon.Run {
		t.Fatal("disabled demon was loaded as enabled")
	}
	if err := SetDemonRun(path, true); err != nil {
		t.Fatal(err)
	}
	loaded, err = Load(path)
	if err != nil || !loaded.Demon.Run {
		t.Fatalf("re-enable failed: %+v %v", loaded, err)
	}
}

func TestDemonRunAddsMissingSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("root = \"docs\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDemonRun(path, false); err != nil {
		t.Fatal(err)
	}
	text, _ := os.ReadFile(path)
	if !strings.Contains(string(text), "[demon]\nrun = false\n") {
		t.Fatalf("missing demon section: %q", text)
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

func TestDiscoverWithinDoesNotCrossRepositoryBoundary(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "docs", "guide")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".demon-docs.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverWithin(child, filepath.Join(root, "docs")); got != "" {
		t.Fatalf("discovery crossed boundary: %s", got)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", ".demon-docs.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverWithin(child, filepath.Join(root, "docs")); got != filepath.Join(root, "docs", ".demon-docs.toml") {
		t.Fatalf("bounded discovery missed config: %s", got)
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
