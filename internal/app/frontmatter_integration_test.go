package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontmatterCheckIsReadOnlyAndFixIsIdempotent(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	page := filepath.Join(docs, "page.md")
	outside := filepath.Join(repo, "outside.md")
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), frontmatterTestConfig(false, "yaml"))
	writeTestFile(t, page, "# Page\n\nBody\n")
	writeTestFile(t, outside, "# Outside\n")

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"check", "--docs"}, &stdout, &stderr); code != 1 {
			t.Fatalf("check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		unchanged, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		if string(unchanged) != "# Page\n\nBody\n" {
			t.Fatalf("check changed the document: %q", unchanged)
		}
		if _, err := os.Stat(filepath.Join(repo, ".ddocs", "refs", "ddocs", "state")); !os.IsNotExist(err) {
			t.Fatalf("check wrote immutable state: %v", err)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		fixed, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		text := string(fixed)
		for _, expected := range []string{
			"---\n",
			"author: Test Author\n",
			"created:",
			"document_id:",
			"document_type: general\n",
			"summary: Documentation.\n",
			"# Page\n\nBody\n",
		} {
			if !strings.Contains(text, expected) {
				t.Fatalf("fixed document missing %q:\n%s", expected, text)
			}
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"check", "--docs"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "check passed") {
			t.Fatalf("second check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})

	outsideText, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(outsideText) != "# Outside\n" {
		t.Fatalf("file outside docs root changed: %q", outsideText)
	}
}

func TestFrontmatterOnlyFixRefreshesExistingLinkIdentity(t *testing.T) {
	repo := t.TempDir()
	configText := strings.Replace(frontmatterTestConfig(false, "yaml"), "[links]\nenabled = false", "[links]\nenabled = true", 1)
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), configText)
	writeTestFile(t, filepath.Join(repo, "docs", "area", "source.md"), "# Source\n\n[Target](../old/target.md)\n")
	writeTestFile(t, filepath.Join(repo, "docs", "old", "target.md"), "# Intended Target\n")
	writeTestFile(t, filepath.Join(repo, "docs", "other", "target.md"), "# Other Target\n")

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--links"}, &stdout, &stderr); code != 0 {
			t.Fatalf("link baseline code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--frontmatter"}, &stdout, &stderr); code != 0 {
			t.Fatalf("frontmatter fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		if err := os.MkdirAll(filepath.Join(repo, "docs", "moved", "deep"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(repo, "docs", "new"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(repo, "docs", "area", "source.md"), filepath.Join(repo, "docs", "moved", "deep", "source.md")); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(repo, "docs", "old", "target.md"), filepath.Join(repo, "docs", "new", "target.md")); err != nil {
			t.Fatal(err)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--links"}, &stdout, &stderr); code != 0 {
			t.Fatalf("moved link fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if strings.Contains(stdout.String(), "unresolved") || strings.Contains(stdout.String(), "Ambiguous link") {
			t.Fatalf("document-ID move became ambiguous: %q", stdout.String())
		}
		moved := readTestFile(t, filepath.Join(repo, "docs", "moved", "deep", "source.md"))
		if !strings.Contains(moved, "[Target](../../new/target.md)") {
			t.Fatalf("link did not follow the intended document ID:\n%s", moved)
		}
	})
}

func TestFrontmatterFixPreservesExistingTOMLFormat(t *testing.T) {
	repo := t.TempDir()
	page := filepath.Join(repo, "docs", "page.md")
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), frontmatterTestConfig(false, "yaml"))
	writeTestFile(t, page, "+++\nauthor = \"Human\"\ncreated = \"2026-07-20\"\ndocument_id = \"11111111-2222-4333-8444-555555555555\"\ndocument_type = \"guide\"\n+++\n# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})

	fixed, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	text := string(fixed)
	if !strings.HasPrefix(text, "+++\n") || strings.HasPrefix(text, "---\n") {
		t.Fatalf("TOML frontmatter was converted: %q", text)
	}
	if !strings.Contains(text, `summary = "Documentation."`) || !strings.HasSuffix(text, "# Page\n") {
		t.Fatalf("TOML repair lost fields or body: %q", text)
	}
}

func TestFrontmatterWarnModePrintsWithoutFailing(t *testing.T) {
	repo := t.TempDir()
	configText := strings.Replace(frontmatterTestConfig(false, "yaml"), `unknown_fields = "remove"`, `unknown_fields = "warn"`, 1)
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), configText)
	writeTestFile(t, filepath.Join(repo, "docs", "page.md"), "---\nauthor: Human\ncreated: \"2026-07-20\"\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\nunknown: kept\n---\n# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"check", "--docs"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "warning:") || !strings.Contains(stdout.String(), "unknown") {
			t.Fatalf("check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
}

func TestInvalidFrontmatterSchemaFailsBeforeIndexMutation(t *testing.T) {
	repo := t.TempDir()
	configText := strings.Replace(frontmatterTestConfig(true, "yaml"), `default_format = "yaml"`, `default_format = "json"`, 1)
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), configText)
	writeTestFile(t, filepath.Join(repo, "docs", "page.md"), "# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--docs"}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "default_format") {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	if _, err := os.Stat(filepath.Join(repo, "docs", "INDEX.md")); !os.IsNotExist(err) {
		t.Fatalf("invalid schema allowed index mutation: %v", err)
	}
}

func TestGeneratedIndexesReceiveFrontmatterInSameFix(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), frontmatterTestConfig(true, "yaml"))
	writeTestFile(t, filepath.Join(docs, "area", "page.md"), "# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})

	rootIndex := readTestFile(t, filepath.Join(docs, "INDEX.md"))
	childIndex := readTestFile(t, filepath.Join(docs, "area", "INDEX.md"))
	for path, text := range map[string]string{
		"root":  rootIndex,
		"child": childIndex,
	} {
		normalized := strings.ReplaceAll(text, "\r\n", "\n")
		if !strings.HasPrefix(normalized, "---\n") || !strings.Contains(normalized, "document_id:") || !strings.Contains(normalized, "doc-ledger:files:start") {
			t.Fatalf("generated %s index did not receive frontmatter and managed content:\n%s", path, text)
		}
		if strings.Count(normalized, "\n---\n") != 1 {
			t.Fatalf("generated %s index contains duplicate YAML frontmatter:\n%s", path, text)
		}
	}
	if !strings.Contains(rootIndex, "# Docs") {
		t.Fatalf("generated root index lost its planned heading:\n%s", rootIndex)
	}
	if !strings.Contains(childIndex, "# Area") || !strings.Contains(childIndex, "Parent index: [Docs](../INDEX.md)") {
		t.Fatalf("generated child index lost its planned heading or parent title:\n%s", childIndex)
	}

	withWorkingDirectory(t, repo, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"check", "--docs"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "check passed") {
			t.Fatalf("check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
}

func frontmatterTestConfig(indexEnabled bool, defaultFormat string) string {
	return `docs_root = "docs"
index_file = "INDEX.md"

[index]
enabled = ` + boolText(indexEnabled) + `

[links]
enabled = false

[frontmatter]
enabled = true
default_format = "` + defaultFormat + `"
allowed_formats = ["yaml", "toml"]
default_author = "Test Author"
unknown_fields = "remove"

[frontmatter.fields.document_id]
type = "uuid"
required = true
immutable = true
generated = true

[frontmatter.fields.author]
type = "string"
required = true
default_from = "frontmatter.default_author"

[frontmatter.fields.document_type]
type = "string"
required = true
default = "general"

[frontmatter.fields.created]
type = "date"
required = true
immutable = true
generated = true

[frontmatter.fields.summary]
type = "string"
required = true
default = "Documentation."
`
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
