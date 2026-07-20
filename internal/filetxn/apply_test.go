package filetxn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyAndRollbackBatch(t *testing.T) {
	root := t.TempDir()
	paths := []string{filepath.Join(root, "one.md"), filepath.Join(root, "two.md")}
	before := [][]byte{[]byte("one before\n"), []byte("two before\n")}
	after := [][]byte{[]byte("one after\n"), []byte("two after\n")}
	rewrites := make([]Rewrite, len(paths))
	for index, path := range paths {
		if err := os.WriteFile(path, before[index], 0o640); err != nil {
			t.Fatal(err)
		}
		rewrites[index] = New(path, before[index], after[index])
	}

	suppressions, err := Apply(rewrites)
	if err != nil {
		t.Fatal(err)
	}
	if len(suppressions) != len(rewrites) {
		t.Fatalf("suppressions=%d, want %d", len(suppressions), len(rewrites))
	}
	for index, path := range paths {
		assertFileContents(t, path, after[index])
	}
	if err := Rollback(rewrites); err != nil {
		t.Fatal(err)
	}
	for index, path := range paths {
		assertFileContents(t, path, before[index])
	}
}

func TestPreflightFailurePreventsEveryWrite(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.md")
	second := filepath.Join(root, "second.md")
	if err := os.WriteFile(first, []byte("first before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("second before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rewrites := []Rewrite{
		New(first, []byte("first before\n"), []byte("first after\n")),
		New(second, []byte("second before\n"), []byte("second after\n")),
	}
	if err := os.WriteFile(second, []byte("external edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Apply(rewrites); err == nil {
		t.Fatal("expected stale-source preflight failure")
	}
	assertFileContents(t, first, []byte("first before\n"))
	assertFileContents(t, second, []byte("external edit\n"))
}

func TestRollbackRefusesNewerContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.md")
	before := []byte("before\n")
	after := []byte("after\n")
	rewrite := New(path, before, after)
	if err := os.WriteFile(path, before, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply([]Rewrite{rewrite}); err != nil {
		t.Fatal(err)
	}
	newer := []byte("newer\n")
	if err := os.WriteFile(path, newer, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Rollback([]Rewrite{rewrite}); err == nil {
		t.Fatal("expected rollback to reject newer content")
	}
	assertFileContents(t, path, newer)
}

func assertFileContents(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s=%q, want %q", path, got, want)
	}
}
