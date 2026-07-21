package frontmatter

import (
	"errors"
	"fmt"
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

func TestScopedBuildReusesUntouchedDocumentsWithoutReadingThem(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\n# Guide\n"
	changed := filepath.Join(docs, "changed.md")
	untouched := filepath.Join(docs, "untouched.md")
	if err := os.WriteFile(changed, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	untouchedText := strings.Replace(text, "11111111-2222-4333-8444-555555555555", "22222222-3333-4444-8555-666666666666", 1)
	if err := os.WriteFile(untouched, []byte(untouchedText), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	if _, err := Build(root, docs, cfg, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(untouched, []byte("not valid frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changed, []byte("not valid frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildScoped(root, docs, cfg, false, time.Now(), []string{changed})
	if err != nil {
		t.Fatal(err)
	}
	if plan.cacheHits != 1 || !plan.Failed() {
		t.Fatalf("scoped result did not reuse untouched document: hits=%d diagnostics=%v", plan.cacheHits, plan.Diagnostics)
	}
}

func TestScopedBuildRequestsFallbackWhenUntouchedCacheIsIncomplete(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\n# Guide\n"
	changed := filepath.Join(docs, "changed.md")
	untouched := filepath.Join(docs, "untouched.md")
	if err := os.WriteFile(changed, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	if _, err := Build(root, docs, cfg, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(untouched, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := BuildScoped(root, docs, cfg, false, time.Now(), []string{changed})
	if !errors.Is(err, ErrScopedReuseUnavailable) {
		t.Fatalf("error=%v want scoped reuse sentinel", err)
	}
}

func TestScopedBuildRequestsFallbackWhenDuplicateTouchesUntouchedDocument(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	base := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: %s\ndocument_type: guide\nsummary: Existing\n---\n# Guide\n"
	changed := filepath.Join(docs, "a.md")
	untouched := filepath.Join(docs, "b.md")
	untouchedID := "22222222-3333-4444-8555-666666666666"
	if err := os.WriteFile(changed, []byte(fmt.Sprintf(base, "11111111-2222-4333-8444-555555555555")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(untouched, []byte(fmt.Sprintf(base, untouchedID)), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter = schema()
	if _, err := Build(root, docs, cfg, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changed, []byte(fmt.Sprintf(base, untouchedID)), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := BuildScoped(root, docs, cfg, true, time.Now(), []string{changed})
	if !errors.Is(err, ErrScopedReuseUnavailable) {
		t.Fatalf("error=%v want scoped reuse sentinel", err)
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
	repeatFix, err := Build(root, docs, cfg, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if repeatFix.cacheHits != 1 || len(repeatFix.immutable) != 0 {
		t.Fatalf("repeated clean fix republished immutable state: hits=%d immutable=%v", repeatFix.cacheHits, repeatFix.immutable)
	}
}
