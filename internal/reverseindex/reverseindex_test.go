package reverseindex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
)

func TestBuildCreatesCodeFolderIndexWithDocumentationBacklinks(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "server", "runtime")
	mustWrite(t, filepath.Join(docsRoot, "runtime.md"), "# Runtime Guide\n\n## Code map\n\n- `server/runtime/session.go`\n- `server/runtime/session.go`\n")
	mustWrite(t, filepath.Join(codeRoot, "session.go"), "package runtime\n")
	mustWrite(t, filepath.Join(codeRoot, "transport.go"), "package runtime\n")
	mustWrite(t, filepath.Join(codeRoot, "README.md"), "# Runtime Package\n\nManual context stays here.\n")

	plan, err := Build(repositoryRoot, docsRoot, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || plan.IndexCount != 1 {
		t.Fatalf("updates=%d indexes=%d diagnostics=%v", len(plan.Updates), plan.IndexCount, plan.Diagnostics)
	}
	if _, err := Apply(repositoryRoot, plan); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(codeRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	for _, expected := range []string{"Manual context stays here.", "<!-- doc-ledger:reverse-index:start -->", "- [session.go](session.go)", "- [transport.go](transport.go)", "[Runtime Guide](../../docs/runtime.md)"} {
		if !strings.Contains(text, expected) {
			t.Errorf("missing %q:\n%s", expected, text)
		}
	}
	if strings.Count(text, "[Runtime Guide]") != 1 {
		t.Fatalf("duplicate documentation backlink:\n%s", text)
	}
	clean, err := Build(repositoryRoot, docsRoot, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(clean.Updates) != 0 {
		t.Fatalf("second build was not deterministic: %#v", clean.Updates)
	}
}

func TestBuildExcludesDocsAndNestedWorktrees(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	mustWrite(t, filepath.Join(docsRoot, "feature.md"), "# Feature\n\n## Code map\n\n- `src/feature.go`\n")
	mustWrite(t, filepath.Join(repositoryRoot, "src", "feature.go"), "package src\n")
	mustWrite(t, filepath.Join(repositoryRoot, ".worktrees", "other", "src", "copy.go"), "package src\n")

	plan, err := Build(repositoryRoot, docsRoot, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || filepath.Dir(plan.Updates[0].Path) != filepath.Join(repositoryRoot, "src") {
		t.Fatalf("unexpected updates: %#v", plan.Updates)
	}
}

func TestBuildRendersDirectoryMappingsAsFolderDocumentation(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "service")
	mustWrite(t, filepath.Join(docsRoot, "service.md"), "# Service Design\n\n## Code map\n\n- `service/`\n")
	mustWrite(t, filepath.Join(codeRoot, "main.go"), "package service\n")

	plan, err := Build(repositoryRoot, docsRoot, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || !strings.Contains(plan.Updates[0].NewText, "Folder documentation:") || !strings.Contains(plan.Updates[0].NewText, "[Service Design](../docs/service.md)") {
		t.Fatalf("unexpected directory mapping:\n%s", plan.Updates[0].NewText)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
