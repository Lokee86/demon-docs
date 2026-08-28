package codemapsemantic

import "testing"

func TestSaveAndLoadBaselines(t *testing.T) {
	root := t.TempDir()
	baselines := []Baseline{
		{Document: "docs/a.md", DocumentSHA256: "abc", SnapshotID: "sha256:one"},
		{Document: "docs/b.md", DocumentSHA256: "def", SnapshotID: "sha256:two"},
	}
	if err := SaveAll(root, baselines); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadMany(root, []string{"docs/a.md", "docs/b.md", "docs/missing.md"})
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 || loaded["docs/a.md"].SnapshotID != "sha256:one" || loaded["docs/b.md"].DocumentSHA256 != "def" {
		t.Fatalf("unexpected baselines: %#v", loaded)
	}
}
