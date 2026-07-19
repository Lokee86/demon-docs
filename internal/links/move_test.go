package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoveRewritesRecognizedLinksWithoutState(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n\n## Intro\n")
	if err := os.MkdirAll(filepath.Join(root, "docs", "manual"), 0o755); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(root, "docs", "index.md")
	writeMoveFixture(t, indexPath, strings.Join([]string{
		"[Guide](guide.md#intro)",
		"![Guide image](guide.md?raw=1#intro)",
		"[guide-ref]: guide.md?view=1#intro",
		"[[guide|Wiki]]",
		"<a href=\"guide.md#intro\">HTML</a>",
		"",
	}, "\n"))

	plan, err := PlanMove(root, filepath.Join(root, "docs", "guide.md"), filepath.Join(root, "docs", "manual", "guide-renamed.md"))
	if err != nil {
		t.Fatal(err)
	}
	if plan.RewrittenLinks != 5 || len(plan.Updates) != 1 {
		t.Fatalf("rewritten=%d updates=%d", plan.RewrittenLinks, len(plan.Updates))
	}
	if err := ApplyMove(plan); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "guide.md")); !os.IsNotExist(err) {
		t.Fatalf("old source still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "manual", "guide-renamed.md")); err != nil {
		t.Fatal(err)
	}
	updated := readMoveFixture(t, indexPath)
	for _, want := range []string{
		"[Guide](manual/guide-renamed.md#intro)",
		"![Guide image](manual/guide-renamed.md?raw=1#intro)",
		"[guide-ref]: manual/guide-renamed.md?view=1#intro",
		"[[guide-renamed|Wiki]]",
		"<a href=\"manual/guide-renamed.md#intro\">HTML</a>",
	} {
		if !strings.Contains(updated, want) {
			t.Errorf("missing %q:\n%s", want, updated)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".ddocs")); !os.IsNotExist(err) {
		t.Fatalf("stateless move created .ddocs: %v", err)
	}
}

func TestMoveRewritesRelativeLinksInsideMovedMarkdown(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "README.md"), "# Root\n")
	writeMoveFixture(t, filepath.Join(root, "docs", "topic.md"), "[Root](../README.md)\n[Section](#section)\n")
	if err := os.MkdirAll(filepath.Join(root, "docs", "archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := PlanMove(root, filepath.Join(root, "docs", "topic.md"), filepath.Join(root, "docs", "archive", "topic.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMove(plan); err != nil {
		t.Fatal(err)
	}
	moved := readMoveFixture(t, filepath.Join(root, "docs", "archive", "topic.md"))
	if moved != "[Root](../../README.md)\n[Section](#section)\n" {
		t.Fatalf("moved document link=%q", moved)
	}
}

func TestMoveDirectoryRewritesIncomingAndMovedSourceLinks(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "assets", "image.png"), "image")
	writeMoveFixture(t, filepath.Join(root, "docs", "guide", "page.md"), "![Image](../../assets/image.png)\n")
	writeMoveFixture(t, filepath.Join(root, "overview.md"), "[Page](docs/guide/page.md)\n")
	if err := os.MkdirAll(filepath.Join(root, "docs", "archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := PlanMove(root, filepath.Join(root, "docs", "guide"), filepath.Join(root, "docs", "archive"))
	if err != nil {
		t.Fatal(err)
	}
	if !plan.SourceIsDirectory {
		t.Fatal("directory move was not identified")
	}
	if err := ApplyMove(plan); err != nil {
		t.Fatal(err)
	}
	page := readMoveFixture(t, filepath.Join(root, "docs", "archive", "guide", "page.md"))
	if page != "![Image](../../../assets/image.png)\n" {
		t.Fatalf("moved page=%q", page)
	}
	overview := readMoveFixture(t, filepath.Join(root, "overview.md"))
	if overview != "[Page](docs/archive/guide/page.md)\n" {
		t.Fatalf("overview=%q", overview)
	}
}

func TestMoveMakesBareWikiPathExplicitWhenMoveWouldCreateAmbiguity(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "guide.md"), "# Primary\n")
	writeMoveFixture(t, filepath.Join(root, "other", "guide.md"), "# Other\n")
	writeMoveFixture(t, filepath.Join(root, "index.md"), "[[guide]]\n")
	if err := os.MkdirAll(filepath.Join(root, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := PlanMove(root, filepath.Join(root, "guide.md"), filepath.Join(root, "archive", "guide.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMove(plan); err != nil {
		t.Fatal(err)
	}
	if got := readMoveFixture(t, filepath.Join(root, "index.md")); got != "[[archive/guide]]\n" {
		t.Fatalf("index=%q", got)
	}
}

func TestMoveRejectsSymlinkedDestinationOutsideRepository(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "doc.md"), "# Doc\n")
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := PlanMove(root, filepath.Join(root, "doc.md"), filepath.Join(link, "doc.md"))
	if err == nil || !strings.Contains(err.Error(), "destination resolves outside repository root") {
		t.Fatalf("error=%v", err)
	}
}

func TestMoveSupportsCaseOnlyRename(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "Guide.md"), "# Guide\n")
	writeMoveFixture(t, filepath.Join(root, "index.md"), "[Guide](Guide.md)\n")

	plan, err := PlanMove(root, filepath.Join(root, "Guide.md"), filepath.Join(root, "guide.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMove(plan); err != nil {
		t.Fatal(err)
	}
	if got := readMoveFixture(t, filepath.Join(root, "index.md")); got != "[Guide](guide.md)\n" {
		t.Fatalf("index=%q", got)
	}
	if _, err := os.Stat(filepath.Join(root, "guide.md")); err != nil {
		t.Fatal(err)
	}
}

func TestMoveRejectsAffectedAmbiguousWikiTarget(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "a", "guide.md"), "# A\n")
	writeMoveFixture(t, filepath.Join(root, "b", "guide.md"), "# B\n")
	writeMoveFixture(t, filepath.Join(root, "index.md"), "[[guide]]\n")
	if err := os.MkdirAll(filepath.Join(root, "archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := PlanMove(root, filepath.Join(root, "a"), filepath.Join(root, "archive"))
	if err == nil || !strings.Contains(err.Error(), "ambiguous wiki target") {
		t.Fatalf("error=%v", err)
	}
}

func TestMovePreflightRefusesChangedMarkdown(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "guide.md"), "# Guide\n")
	index := filepath.Join(root, "index.md")
	writeMoveFixture(t, index, "[Guide](guide.md)\n")

	plan, err := PlanMove(root, filepath.Join(root, "guide.md"), filepath.Join(root, "renamed.md"))
	if err != nil {
		t.Fatal(err)
	}
	writeMoveFixture(t, index, "changed\n")
	if err := ApplyMove(plan); err == nil || !strings.Contains(err.Error(), "changed before apply") {
		t.Fatalf("error=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "guide.md")); err != nil {
		t.Fatalf("source moved despite failed preflight: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "renamed.md")); !os.IsNotExist(err) {
		t.Fatalf("destination exists after failed preflight: %v", err)
	}
}

func TestMoveRejectsIgnoredSource(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, ".docignore"), "ignored/**\n")
	writeMoveFixture(t, filepath.Join(root, "ignored", "doc.md"), "# Doc\n")

	_, err := PlanMove(root, filepath.Join(root, "ignored", "doc.md"), filepath.Join(root, "doc.md"))
	if err == nil || !strings.Contains(err.Error(), "source is excluded") {
		t.Fatalf("error=%v", err)
	}
}

func TestMoveRejectsDestinationOutsideRepository(t *testing.T) {
	root := t.TempDir()
	writeMoveFixture(t, filepath.Join(root, "doc.md"), "# Doc\n")
	_, err := PlanMove(root, filepath.Join(root, "doc.md"), filepath.Join(filepath.Dir(root), "outside.md"))
	if err == nil || !strings.Contains(err.Error(), "destination must be inside repository root") {
		t.Fatalf("error=%v", err)
	}
}

func writeMoveFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readMoveFixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
