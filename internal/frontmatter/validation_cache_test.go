package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestUnchangedCleanDocumentUsesValidationCacheAndPolicyChangesInvalidate(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(docs, "guide.md")
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\n# Guide\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	first, err := Build(root, docs, cfg, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if first.cacheHits != 0 || len(first.Diagnostics) != 0 {
		t.Fatalf("unexpected first validation result: hits=%d diagnostics=%v", first.cacheHits, first.Diagnostics)
	}
	second, err := Build(root, docs, cfg, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if second.cacheHits != 1 || len(second.Diagnostics) != 0 {
		t.Fatalf("unchanged clean document was not cached: hits=%d diagnostics=%v", second.cacheHits, second.Diagnostics)
	}
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(text, "\n", "\r\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	lineEndingChange, err := Build(root, docs, cfg, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if lineEndingChange.cacheHits != 0 || len(lineEndingChange.Diagnostics) != 0 {
		t.Fatalf("raw content change did not invalidate cache: hits=%d diagnostics=%v", lineEndingChange.cacheHits, lineEndingChange.Diagnostics)
	}
	cfg.Frontmatter.UnknownFields = "warn"
	third, err := Build(root, docs, cfg, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if third.cacheHits != 0 || len(third.Diagnostics) != 0 {
		t.Fatalf("frontmatter policy change did not invalidate cache: hits=%d diagnostics=%v", third.cacheHits, third.Diagnostics)
	}
}

func TestValidationCacheDoesNotHideNewDuplicateDocumentID(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\n# Guide\n"
	for _, name := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(docs, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	firstPath := filepath.Join(docs, "b.md")
	if err := os.Remove(firstPath); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root, docs, cfg, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(firstPath, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := Build(root, docs, cfg, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.cacheHits != 0 || len(plan.Diagnostics) != 2 || !hasUnresolved(plan.Diagnostics, "document_id") {
		t.Fatalf("duplicate ID was hidden by cache: hits=%d diagnostics=%v", plan.cacheHits, plan.Diagnostics)
	}
}

func TestFixPublishesImmutableValuesRetainedByCheckCache(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(docs, "guide.md")
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\n# Guide\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	if _, err := Build(root, docs, cfg, false, now); err != nil {
		t.Fatal(err)
	}
	plan, err := Build(root, docs, cfg, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if plan.cacheHits != 1 || len(plan.immutable) != 1 {
		t.Fatalf("fix did not recover immutable values from cache: hits=%d immutable=%v", plan.cacheHits, plan.immutable)
	}
	if _, err := Apply(root, docs, plan); err != nil {
		t.Fatal(err)
	}
	final, err := Build(root, docs, cfg, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if final.cacheHits != 1 || final.Failed() {
		t.Fatalf("immutable publication changed clean cache result: hits=%d diagnostics=%v", final.cacheHits, final.Diagnostics)
	}
}
