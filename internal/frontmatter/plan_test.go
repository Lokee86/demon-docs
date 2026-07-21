package frontmatter

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestBuildApplyIsIdempotentPreservesCRLFAndKeepsIDAcrossMove(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(docs, "guide.md")
	body := "# Guide\r\n\r\nBody\r\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	field := cfg.Frontmatter.Fields["summary"]
	field.Default = "Guide summary"
	cfg.Frontmatter.Fields["summary"] = field
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)

	plan, err := Build(repo, docs, cfg, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Failed() || len(plan.Updates) != 1 {
		t.Fatalf("unexpected repair plan: %+v", plan)
	}
	if _, err := Apply(repo, docs, plan); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "\r\n") || !strings.HasSuffix(string(written), body) {
		t.Fatalf("newline or body changed: %q", written)
	}
	id := regexp.MustCompile(`(?m)^document_id:\s+([^\r\n]+)`).FindStringSubmatch(string(written))
	if len(id) != 2 {
		t.Fatalf("missing document ID: %q", written)
	}

	second, err := Build(repo, docs, cfg, true, now.AddDate(1, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if second.Failed() || len(second.Updates) != 0 {
		t.Fatalf("second pass not idempotent: %+v", second)
	}

	moved := filepath.Join(docs, "moved.md")
	if err := os.Rename(path, moved); err != nil {
		t.Fatal(err)
	}
	third, err := Build(repo, docs, cfg, true, now.AddDate(2, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if third.Failed() || len(third.Updates) != 0 {
		t.Fatalf("move changed stable frontmatter: %+v", third)
	}
	movedData, _ := os.ReadFile(moved)
	if !strings.Contains(string(movedData), "document_id: "+strings.TrimSpace(id[1])) {
		t.Fatalf("ID changed across move: %q", movedData)
	}
	tampered := regexp.MustCompile(`(?m)^created:.*$`).ReplaceAllString(string(movedData), `created: "2030-01-01"`)
	if err := os.WriteFile(moved, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	fourth, err := Build(repo, docs, cfg, true, now.AddDate(3, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if fourth.Failed() || len(fourth.Updates) != 1 {
		t.Fatalf("immutable history did not follow the document ID across the move: %+v", fourth)
	}
	if _, err := Apply(repo, docs, fourth); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(moved)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(restored), `created: "2026-07-20"`) {
		t.Fatalf("immutable date was not restored after the move: %q", restored)
	}
}

func TestBuildUsesFormatRulesAndGeneratedIndexDefaults(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string]string{
		filepath.Join(docs, "README.md"): "# Docs\n",
		filepath.Join(docs, "guide.md"): `---
author: Demon Docs
created: "2026-07-20"
document_id: 11111111-2222-4333-8444-555555555555
document_type: ""
summary: Existing summary
---
# Guide
`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.IndexFile = "README.md"
	cfg.Frontmatter = schema()
	cfg.Format.Enabled = true
	cfg.Format.DefaultSchema = "general"
	cfg.Format.PathRules = []config.FormatPathRule{{Pattern: "**/README.md", Schema: "index"}}
	plan, err := Build(repo, docs, cfg, true, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Failed() || len(plan.Updates) != 2 {
		t.Fatalf("unexpected repair plan: %+v", plan)
	}
	updates := map[string]string{}
	for _, update := range plan.Updates {
		updates[filepath.Base(update.Path)] = update.NewText
	}
	if !strings.Contains(updates["README.md"], "document_type: index") ||
		!strings.Contains(updates["README.md"], "author: Demon Docs") ||
		!strings.Contains(updates["README.md"], "summary: Generated documentation folder index.") {
		t.Fatalf("index defaults did not follow the path-selected schema: %q", updates["README.md"])
	}
	if !strings.Contains(updates["guide.md"], "document_type: general") ||
		!strings.Contains(updates["guide.md"], "author: Demon Docs") ||
		!strings.Contains(updates["guide.md"], "summary: Existing summary") {
		t.Fatalf("ordinary defaults changed unexpectedly: %q", updates["guide.md"])
	}
}

func TestBuildUsesGeneratedIndexDefaultsWithoutFormat(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(docs, "!INDEX.md")
	if err := os.WriteFile(path, []byte("# Docs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.IndexFile = "!INDEX.md"
	cfg.Frontmatter = schema()
	cfg.Format.Enabled = false

	plan, err := Build(repo, docs, cfg, true, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Failed() || len(plan.Updates) != 1 {
		t.Fatalf("unexpected repair plan: %+v", plan)
	}
	updated := plan.Updates[0].NewText
	if !strings.Contains(updated, "author: Demon Docs") ||
		!strings.Contains(updated, "summary: Generated documentation folder index.") {
		t.Fatalf("index defaults depended on document-format enforcement: %q", updated)
	}
}

func TestApplyRejectsLineEndingOnlyChangesAfterPlanning(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(docs, "guide.md")
	if err := os.WriteFile(path, []byte("# Guide\r\n\r\nBody\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Frontmatter = schema()
	field := cfg.Frontmatter.Fields["summary"]
	field.Default = "Guide summary"
	cfg.Frontmatter.Fields["summary"] = field
	plan, err := Build(repo, docs, cfg, true, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# Guide\n\nBody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(repo, docs, plan); err == nil {
		t.Fatal("expected exact-byte preflight failure")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Guide\n\nBody\n" {
		t.Fatalf("stale plan overwrote newer bytes: %q", data)
	}
}
