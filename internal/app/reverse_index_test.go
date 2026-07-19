package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReverseIndexCheckFixAndWatchOnce(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	if err := os.MkdirAll(filepath.Join(repositoryRoot, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsRoot, "feature.md"), []byte("# Feature Guide\n\n## Code map\n\n- `src/feature.go`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repositoryRoot, "src", "feature.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"reverse-index", "check", "src"}, &out, &errOut); code != 1 || !strings.Contains(out.String(), "check failed") {
			t.Fatalf("check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if _, err := os.Stat(filepath.Join(repositoryRoot, "src", "README.md")); !os.IsNotExist(err) {
			t.Fatal("reverse-index check wrote an index")
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"reverse-index", "fix", "src"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "updated 1 file") {
			t.Fatalf("fix code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"reverse-index", "check", "src"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "check passed") {
			t.Fatalf("clean check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"reverse-index", "watch", "--once", "src"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "watch updated 0 file") {
			t.Fatalf("watch code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseIndexRequiresConfiguredOrPositionalRoots(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d err=%q", code, errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"reverse-index", "check"}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "no reverse-index roots configured") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseIndexUsesConfiguredRoots(t *testing.T) {
	repositoryRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repositoryRoot, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repositoryRoot, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repositoryRoot, "docs", "feature.md"), []byte("# Feature\n\n## Code map\n\n- `src/feature.go`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repositoryRoot, "src", "feature.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d err=%q", code, errOut.String())
		}
		configPath := filepath.Join(repositoryRoot, ".ddocs", "config.toml")
		configText := "docs_root = \"docs\"\nindex_file = \"README.md\"\n\n[reverse_index]\nroots = [\"src\"]\n"
		if err := os.WriteFile(configPath, []byte(configText), 0o644); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"reverse-index", "check"}, &out, &errOut); code != 1 || !strings.Contains(out.String(), "check failed") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseIndexAppearsInTopLevelHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"--help"}, &out, &errOut); code != 0 {
		t.Fatalf("code=%d err=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "reverse-index") {
		t.Fatalf("help omitted reverse-index:\n%s", out.String())
	}
}
