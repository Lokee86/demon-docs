package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionAndUnknownCommandExitCodes(t *testing.T) {
	var out, err bytes.Buffer
	if code := Run(context.Background(), []string{"--version"}, &out, &err); code != 0 || out.String() != "ddocs 0.3.5\n" {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), err.String())
	}
	out.Reset()
	err.Reset()
	if code := Run(context.Background(), []string{"nope"}, &out, &err); code != 2 || !strings.Contains(err.String(), "invalid choice") {
		t.Fatalf("code=%d err=%q", code, err.String())
	}
}

func TestHelpUsesStdoutAndSuccess(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"init", "--help"}, {"status", "--help"}, {"fix", "--help"}, {"config", "paths", "--help"}, {"config", "show", "--help"}, {"config", "init", "--help"}} {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), args, &out, &errOut); code != 0 || out.Len() == 0 || errOut.Len() != 0 {
			t.Fatalf("args=%v code=%d out=%q err=%q", args, code, out.String(), errOut.String())
		}
	}
}

func TestMissingCommandAndUnexpectedArgumentsFail(t *testing.T) {
	for _, args := range [][]string{nil, {"status", "extra"}, {"fix", "extra"}, {"config", "paths", "extra"}, {"config", "show", "extra"}, {"config", "init", "--local", "extra"}, {"config", "init", "--local", "--global"}} {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), args, &out, &errOut); code != 2 || errOut.Len() == 0 {
			t.Fatalf("args=%v code=%d out=%q err=%q", args, code, out.String(), errOut.String())
		}
	}
}

func TestFixCheckAndOverrides(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "page.md"), []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	args := []string{"fix", "-i", "--root", root, "--no-local-config", "--no-global-config", "--index-file", "!README.md", "--parent-link-indexed-files"}
	if code := Run(context.Background(), args, &out, &errOut); code != 0 {
		t.Fatalf("code=%d err=%s", code, errOut.String())
	}
	if _, err := os.Stat(filepath.Join(root, "!README.md")); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile(filepath.Join(root, "page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "./!README.md") {
		t.Fatal(string(page))
	}
	out.Reset()
	errOut.Reset()
	if code := Run(context.Background(), []string{"check", "-i", "--root", root, "--no-local-config", "--no-global-config", "--index-file", "!README.md", "--parent-link-indexed-files"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "check passed") {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
}
func TestCheckReportsDriftWithoutWriting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "page.md")
	if err := os.WriteFile(path, []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"check", "-i", "--root", root}, &out, &errOut); code != 1 {
		t.Fatalf("code=%d out=%s err=%s", code, out.String(), errOut.String())
	}
	if _, err := os.Stat(filepath.Join(root, "INDEX.md")); !os.IsNotExist(err) {
		t.Fatal("check wrote index")
	}
}
func TestInitCreatesRepositoryAndCommandsDiscoverItFromChild(t *testing.T) {
	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "docs")
	child := filepath.Join(docsRoot, "guide")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsRoot, "page.md"), []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsRoot, "ignored.md"), []byte("# Ignored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	outsidePath := filepath.Join(repoRoot, "outside.md")
	outsideText := "# Outside\n"
	if err := os.WriteFile(outsidePath, []byte(outsideText), 0o644); err != nil {
		t.Fatal(err)
	}

	withWorkingDirectory(t, repoRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs/"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		configText, err := os.ReadFile(filepath.Join(repoRoot, ".ddocs", "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(configText), `docs_root = "docs"`) {
			t.Fatalf("config=%q", string(configText))
		}
		if err := os.WriteFile(filepath.Join(repoRoot, ".docignore"), []byte("/docs/ignored.md\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	})

	withWorkingDirectory(t, child, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "-a"}, &out, &errOut); code != 1 || !strings.Contains(out.String(), "unresolved") || !strings.Contains(out.String(), "frontmatter issue") {
			t.Fatalf("fix code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		rootIndex := filepath.Join(docsRoot, "INDEX.md")
		if _, err := os.Stat(rootIndex); err != nil {
			t.Fatal(err)
		}
		indexText, err := os.ReadFile(rootIndex)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(indexText), "ignored.md") {
			t.Fatalf("repository-level .docignore was not applied: %s", indexText)
		}
		outsideAfter, err := os.ReadFile(outsidePath)
		if err != nil {
			t.Fatal(err)
		}
		if string(outsideAfter) != outsideText {
			t.Fatalf("outside file changed: %q", outsideAfter)
		}
		if _, err := os.Stat(filepath.Join(repoRoot, "INDEX.md")); !os.IsNotExist(err) {
			t.Fatal("repository root was reconciled instead of docs root")
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"status"}, &out, &errOut); code != 0 {
			t.Fatalf("status code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		for _, value := range []string{"repository root = " + repoRoot, "docs root = " + docsRoot, "config = " + filepath.Join(repoRoot, ".ddocs", "config.toml"), "docignore = " + filepath.Join(repoRoot, ".docignore"), "docs root exists = true", "docignore exists = true"} {
			if !strings.Contains(out.String(), value) {
				t.Errorf("status missing %q:\n%s", value, out.String())
			}
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"init", "--root", "."}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "already initialized") {
			t.Fatalf("reinit code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestStatusFailsOutsideRepository(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"status"}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "no Demon Docs repository found") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestInitRequiresExistingDocsRoot(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(string) {
		for _, args := range [][]string{{"init"}, {"init", "--root", "missing"}} {
			var out, errOut bytes.Buffer
			if code := Run(context.Background(), args, &out, &errOut); code != 2 {
				t.Fatalf("args=%v code=%d out=%q err=%q", args, code, out.String(), errOut.String())
			}
		}
	})
}

func TestFixRepairsLinksBeforeFrontmatterAndRefreshesFinalFingerprints(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), `docs_root = "docs"
index_file = "INDEX.md"

[index]
enabled = true

[links]
enabled = true

[frontmatter]
enabled = true
default_format = "yaml"
allowed_formats = ["yaml"]
unknown_fields = "remove"

[frontmatter.fields.created]
type = "date"
required = true
immutable = true
generated = true

[format]
enabled = false
`)
	writeTestFile(t, filepath.Join(docs, "source.md"), "[target](old/target.md)\n")
	writeTestFile(t, filepath.Join(docs, "old", "target.md"), "# Original target\n")
	writeTestFile(t, filepath.Join(docs, "decoy", "target.md"), "# Decoy target\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--links"}, &out, &errOut); code != 0 {
			t.Fatalf("baseline code=%d out=%q err=%q", code, out.String(), errOut.String())
		}

		moved := filepath.Join(docs, "moved", "target.md")
		if err := os.MkdirAll(filepath.Dir(moved), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(docs, "old", "target.md"), moved); err != nil {
			t.Fatal(err)
		}

		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--all"}, &out, &errOut); code != 0 {
			t.Fatalf("combined fix code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		source, err := os.ReadFile(filepath.Join(docs, "source.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(source), "(moved/target.md)") {
			t.Fatalf("link was not repaired before frontmatter changed the moved target fingerprint:\n%s", source)
		}
		movedText, err := os.ReadFile(moved)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(movedText), "---\n") || !strings.Contains(string(movedText), "created:") {
			t.Fatalf("frontmatter was not applied after link repair:\n%s", movedText)
		}

		final := filepath.Join(docs, "final", "target.md")
		if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(moved, final); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--links"}, &out, &errOut); code != 0 {
			t.Fatalf("second move code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		source, err = os.ReadFile(filepath.Join(docs, "source.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(source), "(final/target.md)") {
			t.Fatalf("post-rewrite fingerprint state was not refreshed:\n%s", source)
		}
	})
}

func TestConfigInitAndShow(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"config", "init", "--local"}, &out, &errOut); code != 0 {
		t.Fatalf("%d %s", code, errOut.String())
	}
	out.Reset()
	if code := Run(context.Background(), []string{"config", "show"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "index_file = 'INDEX.md'") {
		t.Fatalf("code=%d out=%s err=%s", code, out.String(), errOut.String())
	}
}
