package links

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareMarkdownSourcesKeepsResultsInSourceOrder(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.md")
	middlePath := filepath.Join(root, "middle.md")
	lastPath := filepath.Join(root, "last.md")
	writeTestFile(t, firstPath, "[first](one.md)\n")
	writeTestFile(t, middlePath, "[middle](two.md)\n")
	writeTestFile(t, lastPath, "[[three]]\n")

	sources := []markdownSource{
		{path: firstPath},
		{path: middlePath},
		{path: lastPath},
	}
	prepared, err := prepareMarkdownSources(sources, []int{0, 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared[0].parsed.Links) != 1 || prepared[0].parsed.Links[0].RawPath != "one.md" {
		t.Fatalf("first source result = %#v", prepared[0])
	}
	if prepared[1].document.Text != "" || len(prepared[1].parsed.Links) != 0 {
		t.Fatalf("unselected source was prepared: %#v", prepared[1])
	}
	if len(prepared[2].parsed.Links) != 1 || prepared[2].parsed.Links[0].RawPath != "three" {
		t.Fatalf("last source result = %#v", prepared[2])
	}
}

func TestPrepareMarkdownSourcesReturnsErrorsInJobOrder(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first-missing.md")
	secondPath := filepath.Join(root, "second-missing.md")
	sources := []markdownSource{{path: firstPath}, {path: secondPath}}

	_, err := prepareMarkdownSources(sources, []int{0, 1})
	if err == nil {
		t.Fatal("expected missing-source error")
	}
	if !strings.Contains(err.Error(), firstPath) {
		t.Fatalf("error order was not deterministic: %v", err)
	}
}
