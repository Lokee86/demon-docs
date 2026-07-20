package watch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

func TestRelevantEventFiltering(t *testing.T) {
	root := t.TempDir()
	c := config.Default()
	c.Watch.IgnoredDirs = nil
	c.Files.IncludePatterns = []string{"**/*.md", "**/*.png", "**/*.tmp"}
	c.Files.ExcludePatterns = []string{"**/*.tmp"}
	cases := []struct {
		rel  string
		want bool
	}{{"page.md", true}, {"diagram.png", true}, {"scratch.tmp", false}, {"notes.txt", false}, {".docignore", true}, {".git/page.md", false}, {".ddocs/config.toml", false}, {".obsidian/workspace.md", false}, {"logseq/config.md", false}, {"cache/file.swp", false}, {".#page.md", false}}
	for _, tc := range cases {
		path := filepath.Join(root, filepath.FromSlash(tc.rel))
		if got := Relevant(path, c, root); got != tc.want {
			t.Fatalf("%s=%t want %t", tc.rel, got, tc.want)
		}
	}
}

func TestRepositoryOwnedDocignoreAndOutsideEvents(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repositoryRoot, ".docignore"), []byte("/docs/private.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := config.Default()
	if !RelevantWithIgnoreRoot(filepath.Join(repositoryRoot, ".docignore"), c, docsRoot, repositoryRoot) {
		t.Fatal("repository control file was not relevant")
	}
	if RelevantWithIgnoreRoot(filepath.Join(docsRoot, "private.md"), c, docsRoot, repositoryRoot) {
		t.Fatal("repository-owned ignore rule was not applied")
	}
	if RelevantWithIgnoreRoot(filepath.Join(repositoryRoot, "outside.md"), c, docsRoot, repositoryRoot) {
		t.Fatal("event outside docs root was relevant")
	}
}

func TestFormatSchemaEventsBypassPrivateStateIgnore(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	schemaDir := filepath.Join(repositoryRoot, ".ddocs", "schemas")
	documentSchemaDir := filepath.Join(repositoryRoot, ".ddocs", "document-schemas")
	for _, directory := range []string{docsRoot, schemaDir, documentSchemaDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	shared := filepath.Join(schemaDir, "general.toml")
	local := filepath.Join(documentSchemaDir, "document.toml")
	privateObject := filepath.Join(repositoryRoot, ".ddocs", "objects", "object")
	if err := os.MkdirAll(filepath.Dir(privateObject), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{shared, local, privateObject} {
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	c := config.Default()
	c.Format.Enabled = true
	features := Features{Format: true, TrackLinks: true}
	if !relevantSelectedWithPolicy(shared, c, policy, docsRoot, repositoryRoot, features, false) {
		t.Fatal("shared schema event was ignored")
	}
	if !relevantSelectedWithPolicy(local, c, policy, docsRoot, repositoryRoot, features, false) {
		t.Fatal("document-specific schema event was ignored")
	}
	if relevantSelectedWithPolicy(privateObject, c, policy, docsRoot, repositoryRoot, features, false) {
		t.Fatal("unrelated private state event became relevant")
	}
}

func TestRelevantDirectoryAndConfiguredIgnores(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "guide")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	c := config.Default()
	if !Relevant(dir, c, root) {
		t.Fatal("directory event ignored")
	}
	c.Watch.IgnoredDirs = append(c.Watch.IgnoredDirs, "guide")
	if Relevant(dir, c, root) {
		t.Fatal("configured ignored directory was relevant")
	}
}
