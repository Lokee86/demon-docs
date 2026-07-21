package documentpolicy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestUnchangedCleanDocumentFormatUsesCacheAndSchemaChangesInvalidate(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	schemas := filepath.Join(root, ".ddocs", "schemas")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(schemas, 0o755); err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(schemas, "general.toml")
	if err := os.WriteFile(schemaPath, []byte("version = 1\nname = 'general'\nunknown_sections = 'allow'\nduplicate_sections = 'allow'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(docs, "guide.md")
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: general\nsummary: Existing\n---\n# Guide\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter.Enabled = true
	cfg.Format = config.Format{Enabled: true, SchemaDir: ".ddocs/schemas", DocumentSchemaDir: ".ddocs/document-schemas", DefaultSchema: "general"}
	first, err := Build(root, docs, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if first.cacheHits != 0 || len(first.Diagnostics) != 0 {
		t.Fatalf("unexpected first format result: hits=%d diagnostics=%v", first.cacheHits, first.Diagnostics)
	}
	second, err := Build(root, docs, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if second.cacheHits != 1 || len(second.Diagnostics) != 0 {
		t.Fatalf("unchanged clean format was not cached: hits=%d diagnostics=%v", second.cacheHits, second.Diagnostics)
	}
	if err := os.WriteFile(schemaPath, []byte("version = 1\nname = 'general'\ndescription = 'changed'\nunknown_sections = 'allow'\nduplicate_sections = 'allow'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	third, err := Build(root, docs, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if third.cacheHits != 0 || len(third.Diagnostics) != 0 {
		t.Fatalf("schema change did not invalidate format cache: hits=%d diagnostics=%v", third.cacheHits, third.Diagnostics)
	}
}

func TestScopedBuildReusesUntouchedDocumentsWithoutReadingThem(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	schemas := filepath.Join(root, ".ddocs", "schemas")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(schemas, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(schemas, "general.toml"), []byte("version = 1\nname = 'general'\nunknown_sections = 'allow'\nduplicate_sections = 'allow'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: general\nsummary: Existing\n---\n# Guide\n"
	changed := filepath.Join(docs, "changed.md")
	untouched := filepath.Join(docs, "untouched.md")
	untouchedText := strings.Replace(text, "11111111-2222-4333-8444-555555555555", "22222222-3333-4444-8555-666666666666", 1)
	for path, source := range map[string]string{changed: text, untouched: untouchedText} {
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter.Enabled = true
	cfg.Format = config.Format{Enabled: true, SchemaDir: ".ddocs/schemas", DocumentSchemaDir: ".ddocs/document-schemas", DefaultSchema: "general"}
	if _, err := Build(root, docs, cfg, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(untouched, []byte("not valid markdown format\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changed, []byte("---\ninvalid: [\n---\n# Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildScoped(root, docs, cfg, false, []string{changed})
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
	schemas := filepath.Join(root, ".ddocs", "schemas")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(schemas, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(schemas, "general.toml"), []byte("version = 1\nname = 'general'\nunknown_sections = 'allow'\nduplicate_sections = 'allow'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	text := "---\nauthor: Human\ncreated: '2026-07-20'\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: general\nsummary: Existing\n---\n# Guide\n"
	changed := filepath.Join(docs, "changed.md")
	untouched := filepath.Join(docs, "untouched.md")
	if err := os.WriteFile(changed, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = "docs"
	cfg.Frontmatter.Enabled = true
	cfg.Format = config.Format{Enabled: true, SchemaDir: ".ddocs/schemas", DocumentSchemaDir: ".ddocs/document-schemas", DefaultSchema: "general"}
	if _, err := Build(root, docs, cfg, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(untouched, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := BuildScoped(root, docs, cfg, false, []string{changed})
	if !errors.Is(err, ErrScopedReuseUnavailable) {
		t.Fatalf("error=%v want scoped reuse sentinel", err)
	}
}
