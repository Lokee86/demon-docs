package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaHeadingRenamePropagatesByStableSectionID(t *testing.T) {
	root := initializedDocumentPolicyRepo(t)
	schemaPath := filepath.Join(root, ".ddocs", "schemas", "decision.toml")
	writeTestFile(t, schemaPath, decisionSchema("Decision"))
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"new", "decision", "docs/decision.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("new code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 0 {
			t.Fatalf("initial fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	path := filepath.Join(root, "docs", "decision.md")
	text := strings.Replace(readTestFile(t, path), "TODO", "Authored decision prose.", 1)
	text += "\n## Appendix\n\nUnresolved section.\n"
	writeTestFile(t, path, text)
	writeTestFile(t, schemaPath, decisionSchema("Resolution"))
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 1 {
			t.Fatalf("blocked rename code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if got := readTestFile(t, path); !strings.Contains(got, "## Decision") || strings.Contains(got, "## Resolution") {
			t.Fatalf("blocked document was partially renamed:\n%s", got)
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"format", "ignore", "--heading", "Appendix", "docs/decision.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("ignore code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 0 {
			t.Fatalf("rename fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	updated := readTestFile(t, path)
	if strings.Contains(updated, "## Decision") || !strings.Contains(updated, "## Resolution") || !strings.Contains(updated, "Authored decision prose.") {
		t.Fatalf("schema rename was not propagated safely:\n%s", updated)
	}
}

func TestLargeSharedSchemaChangeInvalidatesDocumentSpecificSchema(t *testing.T) {
	root := initializedDocumentPolicyRepo(t)
	schemaPath := filepath.Join(root, ".ddocs", "schemas", "custom.toml")
	writeTestFile(t, schemaPath, fourSectionSchema("Alpha", "Beta", "Gamma", "Delta"))
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"new", "custom", "docs/custom.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("new code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 0 {
			t.Fatalf("initial fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	path := filepath.Join(root, "docs", "custom.md")
	text := strings.Replace(readTestFile(t, path), "## Alpha", "## Appendix\n\nKeep this.\n\n## Alpha", 1)
	writeTestFile(t, path, text)
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"format", "ignore", "--heading", "Appendix", "docs/custom.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("ignore code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	documentID := frontmatterValue(t, readTestFile(t, path), "document_id")
	localPath := filepath.Join(root, ".ddocs", "document-schemas", documentID+".toml")
	if _, err := os.Stat(localPath); err != nil {
		t.Fatalf("document-specific schema missing before invalidation: %v", err)
	}
	writeTestFile(t, schemaPath, fourSectionSchema("One", "Two", "Three", "Four"))
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 1 {
			t.Fatalf("invalidation code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "document-specific schema invalidated") {
			t.Fatalf("missing invalidation diagnostic: %s", stdout.String())
		}
	})
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		t.Fatalf("invalidated document-specific schema still exists: %v", err)
	}
}

func TestDocumentSpecificInvalidationUsesAcceptedSnapshotAcrossIncrementalChanges(t *testing.T) {
	root := initializedDocumentPolicyRepo(t)
	schemaPath := filepath.Join(root, ".ddocs", "schemas", "custom.toml")
	writeTestFile(t, schemaPath, fourSectionSchema("Alpha", "Beta", "Gamma", "Delta"))
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"new", "custom", "docs/cumulative.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("new code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr); code != 0 {
			t.Fatalf("initial fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	path := filepath.Join(root, "docs", "cumulative.md")
	text := strings.Replace(readTestFile(t, path), "## Alpha", "## Appendix\n\nKeep this.\n\n## Alpha", 1)
	writeTestFile(t, path, text)
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"format", "ignore", "--heading", "Appendix", "docs/cumulative.md"}, &stdout, &stderr); code != 0 {
			t.Fatalf("ignore code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	documentID := frontmatterValue(t, readTestFile(t, path), "document_id")
	localPath := filepath.Join(root, ".ddocs", "document-schemas", documentID+".toml")

	steps := []struct {
		headings    []string
		invalidated bool
	}{
		{headings: []string{"One", "Beta", "Gamma", "Delta"}},
		{headings: []string{"One", "Two", "Gamma", "Delta"}},
		{headings: []string{"One", "Two", "Three", "Delta"}, invalidated: true},
	}
	for index, step := range steps {
		writeTestFile(t, schemaPath, fourSectionSchema(step.headings...))
		withWorkingDirectory(t, root, func(string) {
			var stdout, stderr bytes.Buffer
			code := Run(context.Background(), []string{"fix", "--format"}, &stdout, &stderr)
			if !step.invalidated && code != 0 {
				t.Fatalf("incremental step %d code=%d stdout=%q stderr=%q", index+1, code, stdout.String(), stderr.String())
			}
			if step.invalidated && (code != 1 || !strings.Contains(stdout.String(), "document-specific schema invalidated")) {
				t.Fatalf("cumulative invalidation step code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
		if !step.invalidated {
			if _, err := os.Stat(localPath); err != nil {
				t.Fatalf("local schema disappeared after incremental step %d: %v", index+1, err)
			}
		}
	}
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		t.Fatalf("cumulatively invalidated schema still exists: %v", err)
	}
}

func initializedDocumentPolicyRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	return root
}

func fourSectionSchema(headings ...string) string {
	var builder strings.Builder
	builder.WriteString(`version = 1
name = "custom"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false
`)
	for index, heading := range headings {
		fmt.Fprintf(&builder, "\n[[sections]]\nid = \"section-%d\"\nheading = %q\n", index+1, heading)
	}
	return builder.String()
}

func decisionSchema(heading string) string {
	return `version = 1
name = "decision"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false

[[sections]]
id = "decision"
heading = "` + heading + `"

[[sections]]
id = "context"
heading = "Context"

[[sections]]
id = "consequences"
heading = "Consequences"

[[sections]]
id = "notes"
heading = "Notes"
`
}

func frontmatterValue(t *testing.T, source, key string) string {
	t.Helper()
	for _, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, key+":") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, key+":")), `"`)
		}
	}
	t.Fatalf("frontmatter key %s not found", key)
	return ""
}
