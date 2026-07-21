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

const duplicateTestID = "11111111-2222-4333-8444-555555555555"

func TestBuildStaysInsideDocsRootAndDetectsDuplicateIDs(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := duplicateTestDocument(duplicateTestID)
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

func TestBuildRepairReassignsDuplicateGeneratedDocumentIDs(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := duplicateTestDocument(duplicateTestID)
	one := filepath.Join(docs, "one.md")
	two := filepath.Join(docs, "two.md")
	for _, path := range []string{one, two} {
		if err := os.WriteFile(path, []byte(frontmatter), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Frontmatter = schema()
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)

	plan, err := Build(repo, docs, cfg, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Failed() {
		t.Fatalf("duplicate generated ID should be repairable: %+v", plan.Diagnostics)
	}
	if len(plan.Updates) != 1 || plan.Updates[0].Path != two {
		t.Fatalf("expected only lexicographically later duplicate to change: %+v", plan.Updates)
	}
	if strings.Contains(plan.Updates[0].NewText, "document_id: "+duplicateTestID) {
		t.Fatalf("duplicate ID was not replaced: %s", plan.Updates[0].NewText)
	}
	if !regexp.MustCompile(`(?m)^document_id: [0-9a-f-]{36}$`).MatchString(plan.Updates[0].NewText) {
		t.Fatalf("replacement document ID is not a UUID: %s", plan.Updates[0].NewText)
	}
	if _, err := Apply(repo, docs, plan); err != nil {
		t.Fatal(err)
	}
	oneText, err := os.ReadFile(one)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(oneText), "document_id: "+duplicateTestID) {
		t.Fatalf("canonical owner lost its document ID: %s", oneText)
	}
	second, err := Build(repo, docs, cfg, true, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if second.Failed() || len(second.Updates) != 0 {
		t.Fatalf("duplicate repair did not converge: %+v", second)
	}
}

func TestBuildRepairPreservesRecordedDuplicateOwner(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := duplicateTestDocument(duplicateTestID)
	one := filepath.Join(docs, "one.md")
	two := filepath.Join(docs, "two.md")
	if err := os.WriteFile(two, []byte(frontmatter), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Frontmatter = schema()
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)

	initial, err := Build(repo, docs, cfg, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if initial.Failed() {
		t.Fatalf("initial document was invalid: %+v", initial.Diagnostics)
	}
	if _, err := Apply(repo, docs, initial); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(one, []byte(frontmatter), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := Build(repo, docs, cfg, true, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Failed() {
		t.Fatalf("recorded duplicate owner should be repairable: %+v", plan.Diagnostics)
	}
	if len(plan.Updates) != 1 || plan.Updates[0].Path != one {
		t.Fatalf("recorded owner should retain its ID while the copy changes: %+v", plan.Updates)
	}
	if strings.Contains(plan.Updates[0].NewText, "document_id: "+duplicateTestID) {
		t.Fatalf("copied document retained the recorded owner's ID: %s", plan.Updates[0].NewText)
	}
}

func TestBuildRepairLeavesDuplicateNonGeneratedDocumentIDsUnresolved(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := duplicateTestDocument(duplicateTestID)
	for _, name := range []string{"one.md", "two.md"} {
		if err := os.WriteFile(filepath.Join(docs, name), []byte(frontmatter), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Frontmatter = schema()
	definition := cfg.Frontmatter.Fields["document_id"]
	definition.Generated = false
	cfg.Frontmatter.Fields["document_id"] = definition

	plan, err := Build(repo, docs, cfg, true, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Failed() || len(plan.Updates) != 0 {
		t.Fatalf("non-generated duplicate IDs require authored resolution: %+v", plan)
	}
}

func duplicateTestDocument(id string) string {
	return "---\nauthor: Human\ncreated: \"2026-07-20\"\ndocument_id: " + id + "\ndocument_type: guide\nsummary: Existing\n---\nBody\n"
}
