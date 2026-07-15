package pathutil

import (
	"path/filepath"
	"testing"
)

func TestResolveRelativeUsesProvidedBase(t *testing.T) {
	base := t.TempDir()
	got, err := Resolve(filepath.Join("missing", "docs"), base)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(filepath.Join(base, "missing", "docs"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
func TestResolveAbsoluteIgnoresProvidedBase(t *testing.T) {
	absolute := filepath.Join(t.TempDir(), "docs")
	got, err := Resolve(absolute, filepath.Join(t.TempDir(), "other"))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(absolute)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
func TestRelativeStripsWindowsExtendedPrefix(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skip("Windows path behavior")
	}
	got, err := Relative(`\\?\C:\docs\guide\page.md`, `\\?\C:\docs`)
	if err != nil {
		t.Fatal(err)
	}
	if got != "guide/page.md" {
		t.Fatal(got)
	}
}
