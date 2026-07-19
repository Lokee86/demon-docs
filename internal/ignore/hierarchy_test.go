package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHierarchyAppliesNestedDocignoreWithinItsDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "code", "feature")
	if err := os.MkdirAll(filepath.Join(nested, "generated"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "code", "other", "generated"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, FileName), []byte("generated/\n*.tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	hierarchy, err := LoadHierarchy(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := hierarchy.LoadAncestors(nested); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		path string
		dir  bool
		want bool
	}{
		{filepath.Join(nested, "generated"), true, true},
		{filepath.Join(nested, "notes.tmp"), false, true},
		{filepath.Join(root, "code", "other", "generated"), true, false},
	}
	for _, tc := range cases {
		got, err := hierarchy.Ignored(tc.path, tc.dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("ignored(%s)=%t want %t", tc.path, got, tc.want)
		}
	}
}

func TestHierarchyAllowsDeeperNegation(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "code")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, FileName), []byte("*.generated.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, FileName), []byte("!keep.generated.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hierarchy, err := LoadHierarchy(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := hierarchy.LoadAncestors(child); err != nil {
		t.Fatal(err)
	}
	ignored, err := hierarchy.Ignored(filepath.Join(child, "keep.generated.go"), false)
	if err != nil {
		t.Fatal(err)
	}
	if ignored {
		t.Fatal("deeper .docignore negation did not override the repository pattern")
	}
}
