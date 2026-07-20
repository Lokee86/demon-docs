package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesSchemasAndNewCreatesDocument(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		for _, name := range []string{"general", "service", "planning", "index"} {
			if _, err := os.Stat(filepath.Join(root, ".ddocs", "schemas", name+".toml")); err != nil {
				t.Fatalf("starter schema %s missing: %v", name, err)
			}
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"new", "general", "docs/new-guide.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("new code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	text := readTestFile(t, filepath.Join(root, "docs", "new-guide.md"))
	for _, want := range []string{"document_type: general", "document_id:", "# New Guide", "## Purpose", "## Overview", "## Related docs", "## Notes"} {
		if !strings.Contains(text, want) {
			t.Fatalf("created document missing %q:\n%s", want, text)
		}
	}
}

func TestFreshRepositoryGeneratedIndexConverges(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"new", "general", "docs/page.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("new code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"check", "--docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	index := readTestFile(t, filepath.Join(root, "docs", "README.md"))
	for _, want := range []string{
		"document_type: index",
		"author: TODO",
		"summary: Generated documentation folder index.",
		"## Direct Files",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("generated index missing %q:\n%s", want, index)
		}
	}
	if strings.Contains(index, "## Purpose") {
		t.Fatalf("generated index was enforced as a general document:\n%s", index)
	}
}

func TestNewRefusesExistingFileWithoutForce(t *testing.T) {
	root := initializedDocumentPolicyRepo(t)
	target := filepath.Join(root, "docs", "existing.md")
	writeTestFile(t, target, "original\n")
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"new", "general", "docs/existing.md"}, &stdout, &stderr); code != 2 {
			t.Fatalf("without force code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if got := readTestFile(t, target); got != "original\n" {
			t.Fatalf("existing file changed without force: %q", got)
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"new", "--force", "general", "docs/existing.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("force code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	if got := readTestFile(t, target); got == "original\n" {
		t.Fatal("--force did not overwrite existing file")
	}
}

func TestFormatIgnoreCreatesDocumentSpecificSchemaAndAllowsReorder(t *testing.T) {
	root := initializedDocumentPolicyRepo(t)
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"new", "general", "docs/page.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("new code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	path := filepath.Join(root, "docs", "page.md")
	text := readTestFile(t, path)
	text = strings.Replace(text, "## Purpose", "## Appendix\n\nHuman text.\n\n## Purpose", 1)
	writeTestFile(t, path, text)
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 1 {
			t.Fatalf("unresolved code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		before := readTestFile(t, path)
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"format", "ignore", "--heading", "Appendix", "docs/page.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("ignore code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if got := readTestFile(t, path); got != before {
			t.Fatal("ignore operation changed document content")
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 0 {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	updated := readTestFile(t, path)
	if strings.Index(updated, "## Appendix") < strings.Index(updated, "## Notes") {
		t.Fatalf("accepted unknown section was not moved after shared sections:\n%s", updated)
	}
	if !strings.Contains(updated, "Human text.") {
		t.Fatal("accepted unknown section lost authored prose")
	}
	documentID := frontmatterValue(t, updated, "document_id")
	local := readTestFile(t, filepath.Join(root, ".ddocs", "document-schemas", documentID+".toml"))
	if !strings.Contains(local, `heading = "Appendix"`) {
		t.Fatalf("document-specific schema missing accepted section:\n%s", local)
	}
}

func TestFormatMergeConcatenatesMixedSections(t *testing.T) {
	root := initializedDocumentPolicyRepo(t)
	path := filepath.Join(root, "docs", "page.md")
	writeTestFile(t, path, `---
document_id: 019f7d55-fb2c-7b96-873a-e8c5be32931b
document_type: general
---
# Page

## Purpose

First prose.

## Purpose

- Item
`)
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"format", "merge", "--heading", "Purpose", "docs/page.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("merge code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	updated := readTestFile(t, path)
	if strings.Count(updated, "## Purpose") != 1 || !strings.Contains(updated, "First prose.") || !strings.Contains(updated, "- Item") {
		t.Fatalf("mixed merge did not concatenate both bodies:\n%s", updated)
	}
}
