package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPermanentDirectoriesCannotBeReincluded(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, FileName), []byte("!.git/\n!.demon-docs/\n!.obsidian/\n!logseq/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".git/config", ".demon-docs/index.db", ".obsidian/workspace.json", "logseq/config.edn", "nested/.git/config"} {
		ignored, err := policy.Ignored(filepath.Join(root, filepath.FromSlash(rel)), false)
		if err != nil {
			t.Fatal(err)
		}
		if !ignored {
			t.Fatalf("permanent path was not ignored: %s", rel)
		}
	}
}

func TestDocignoreGitignoreSemantics(t *testing.T) {
	root := t.TempDir()
	contents := "# generated files\n/generated/\n*.tmp\ndrafts/**\n!drafts/keep.md\n"
	if err := os.WriteFile(filepath.Join(root, FileName), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		rel   string
		isDir bool
		want  bool
	}{
		{"generated", true, true},
		{"nested/generated", true, false},
		{"notes.tmp", false, true},
		{"nested/notes.tmp", false, true},
		{"drafts/topic.md", false, true},
		{"drafts/keep.md", false, false},
		{"page.md", false, false},
	}
	for _, tc := range cases {
		got, err := policy.Ignored(filepath.Join(root, filepath.FromSlash(tc.rel)), tc.isDir)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("%s=%t want %t", tc.rel, got, tc.want)
		}
	}
}

func TestMissingDocignoreAllowsNormalPaths(t *testing.T) {
	root := t.TempDir()
	policy, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	ignored, err := policy.Ignored(filepath.Join(root, "guide", "page.md"), false)
	if err != nil {
		t.Fatal(err)
	}
	if ignored {
		t.Fatal("normal path was ignored without a .docignore")
	}
}
