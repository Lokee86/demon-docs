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

func TestBuildStaysInsideDocsRootAndDetectsDuplicateIDs(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := "---\nauthor: Human\ncreated: \"2026-07-20\"\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\nBody\n"
	for _, path := range []string{filepath.Join(docs, "one.md"), filepath.Join(docs, "two.md"), filepath.Join(repo, "outside.md")} {
		if err := os.WriteFile(path, []byte(frontmatter), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Frontmatter = schema()
	plan, err := Build(repo, docs, cfg, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	duplicates := 0
	for _, diagnostic := range plan.Diagnostics {
		if diagnostic.Field == "document_id" && strings.Contains(diagnostic.Message, "duplicate") {
			duplicates++
		}
		if diagnostic.Path == "outside.md" {
			t.Fatalf("outside file was inspected: %+v", diagnostic)
		}
	}
	if duplicates != 2 || !plan.Failed() {
		t.Fatalf("duplicate IDs not detected: %+v", plan.Diagnostics)
	}
	for path, values := range plan.immutable {
		if _, ok := values["document_id"]; ok {
			t.Fatalf("duplicate document ID was recorded as immutable for %s: %#v", path, values)
		}
	}
}
