package pathutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRelativeUsesForwardSlashes(t *testing.T) {
	base := filepath.Join("root", "docs")
	target := filepath.Join(base, "guide", "setup.md")
	got, err := Relative(target, base)
	if err != nil {
		t.Fatal(err)
	}
	if got != "guide/setup.md" || strings.Contains(got, "\\") {
		t.Fatal(got)
	}
}
