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

	plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat())
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
	clean, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(clean.Updates) != 0 {
		t.Fatalf("second build was not deterministic: %#v", clean.Updates)
	}
}

func TestBuildTraversesConfiguredRootsRecursively(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "services")
	mustWrite(t, filepath.Join(docsRoot, "feature.md"), "# Feature\n\n## Code map\n\n- `services/api/handler.go`\n- `client/view.gd`\n")
	mustWrite(t, filepath.Join(codeRoot, "api", "handler.go"), "package api\n")
	mustWrite(t, filepath.Join(codeRoot, "worker", "worker.go"), "package worker\n")
	mustWrite(t, filepath.Join(repositoryRoot, "client", "view.gd"), "extends Node\n")

	plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 2 {
		t.Fatalf("expected recursive indexes for two service folders, got %#v", plan.Updates)
	}
	for _, update := range plan.Updates {
		if strings.Contains(update.Path, filepath.Join(repositoryRoot, "client")) {
			t.Fatalf("wrote outside configured root: %s", update.Path)
		}
	}
}

func TestBuildHonorsNestedDocignoreFiles(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "services")
	mustWrite(t, filepath.Join(docsRoot, "feature.md"), "# Feature\n\n## Code map\n\n- `services/api/handler.go`\n- `services/api/generated/client.go`\n")
	mustWrite(t, filepath.Join(codeRoot, "api", "handler.go"), "package api\n")
	mustWrite(t, filepath.Join(codeRoot, "api", "generated", "client.go"), "package generated\n")
	mustWrite(t, filepath.Join(codeRoot, "api", ".docignore"), "generated/\n")

	plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || filepath.Dir(plan.Updates[0].Path) != filepath.Join(codeRoot, "api") {
		t.Fatalf("nested .docignore was not respected: %#v", plan.Updates)
	}
}

func TestBuildScopesMissingTargetDiagnostics(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "services")
	mustWrite(t, filepath.Join(docsRoot, "missing.md"), "# Missing\n\n## Code map\n\n- `services/missing.go`\n- `client/missing.gd`\n")
	if err := os.MkdirAll(codeRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat())
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(plan.Diagnostics, "\n")
	if !strings.Contains(joined, "services/missing.go") {
		t.Fatalf("scoped missing target was not reported: %v", plan.Diagnostics)
	}
	if strings.Contains(joined, "client/missing.gd") {
		t.Fatalf("out-of-scope target was reported: %v", plan.Diagnostics)
	}
}

func TestBuildRendersDirectoryMappingsAsFolderDocumentation(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "service")
	mustWrite(t, filepath.Join(docsRoot, "service.md"), "# Service Design\n\n## Code map\n\n- `service/`\n")
	mustWrite(t, filepath.Join(codeRoot, "main.go"), "package service\n")

	plan, err := Build(repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat())
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
