package links

import (
	"path/filepath"
	"testing"
)

func TestObsidianShortestMarkdownPathResolvesUniqueRepositoryFile(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "source", "guide.md"), "[Target](target.md)\n")
	writeTestFile(t, filepath.Join(root, "docs", "reference", "target.md"), "# Target\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 0 || len(plan.Links.Links) != 1 || plan.Links.Links[0].Status != "valid" {
		t.Fatalf("Obsidian shortest Markdown path was not resolved: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
	if plan.Links.Links[0].ResolvedPath != "docs/reference/target.md" {
		t.Fatalf("resolved path = %q", plan.Links.Links[0].ResolvedPath)
	}
}

func TestObsidianVaultRootMarkdownAndWikiPathsResolve(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "source", "guide.md"), "[Overview](docs/concepts/overview.md)\n[[docs/guides/overview|Guide overview]]\n")
	writeTestFile(t, filepath.Join(root, "docs", "concepts", "overview.md"), "# Concepts\n")
	writeTestFile(t, filepath.Join(root, "docs", "guides", "overview.md"), "# Guides\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 0 || len(plan.Links.Links) != 2 {
		t.Fatalf("Obsidian vault-root paths were not resolved: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
	if plan.Links.Links[0].ResolvedPath != "docs/concepts/overview.md" || plan.Links.Links[1].ResolvedPath != "docs/guides/overview.md" {
		t.Fatalf("unexpected resolved paths: %#v", plan.Links.Links)
	}
}

func TestObsidianBareWikiImageResolvesUniqueRepositoryFile(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "guide.md"), "![[system-overview.jpg]]\n")
	writeTestFile(t, filepath.Join(root, "docs", "assets", "system-overview.jpg"), "image")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 0 || len(plan.Links.Links) != 1 || plan.Links.Links[0].Status != "valid" {
		t.Fatalf("Obsidian bare wiki image was not resolved: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
	if plan.Links.Links[0].ResolvedPath != "docs/assets/system-overview.jpg" {
		t.Fatalf("resolved path = %q", plan.Links.Links[0].ResolvedPath)
	}
}

func TestObsidianBareMarkdownPathIsAmbiguousWhenFilenameIsDuplicated(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[Overview](overview.md)\n")
	writeTestFile(t, filepath.Join(root, "docs", "concepts", "overview.md"), "# Concepts\n")
	writeTestFile(t, filepath.Join(root, "docs", "guides", "overview.md"), "# Guides\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 1 || len(plan.Links.Links) != 1 || plan.Links.Links[0].Status != "ambiguous" {
		t.Fatalf("duplicate shortest Markdown path was not ambiguous: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
}

func TestExplicitStaleMarkdownPathRemainsBrokenOnInitialScan(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[Configuration](concepts/archive/configuration.md)\n")
	writeTestFile(t, filepath.Join(root, "docs", "concepts", "configuration.md"), "# Configuration\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 1 || len(plan.Links.Links) != 1 || plan.Links.Links[0].Status != "broken" {
		t.Fatalf("explicit stale path should remain broken on baseline: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
}
