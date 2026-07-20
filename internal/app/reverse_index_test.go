package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReverseFlagCheckFixAndWatchOnce(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	mustMakeDir(t, filepath.Join(repositoryRoot, "src"))
	mustMakeDir(t, docsRoot)
	mustWriteAppFile(t, filepath.Join(docsRoot, "feature.md"), "# Feature Guide\n\n## Code map\n\n- `src/feature.go`\n")
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "src", "feature.go"), "package src\n")

	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r", "--reverse-root", "src"}, &out, &errOut); code != 1 || !strings.Contains(out.String(), "check failed") {
			t.Fatalf("check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if _, err := os.Stat(filepath.Join(repositoryRoot, "src", "INDEX.md")); !os.IsNotExist(err) {
			t.Fatal("reverse check wrote an index")
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--reverse", "--reverse-root", "src"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "updated 1 file") {
			t.Fatalf("fix code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r", "--reverse-root", filepath.Join(repositoryRoot, "src")}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "check passed") {
			t.Fatalf("clean check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"watch", "-r", "--once", "--reverse-root", "src"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "watch --reverse updated 0 file") {
			t.Fatalf("watch code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseFlagRequiresConfiguredOrCommandRoots(t *testing.T) {
	repositoryRoot := t.TempDir()
	mustMakeDir(t, filepath.Join(repositoryRoot, "docs"))
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "docs", "feature.md"), "# Feature\n\n## Code map\n\n- `src/feature.go`\n")
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d err=%q", code, errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r"}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "no reverse-index roots configured") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseFlagUsesConfiguredRootsAndCodemapHeading(t *testing.T) {
	repositoryRoot := t.TempDir()
	mustMakeDir(t, filepath.Join(repositoryRoot, "docs"))
	mustMakeDir(t, filepath.Join(repositoryRoot, "src"))
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "docs", "feature.md"), "# Feature\n\n## Implementation map\n\n- `src/feature.go`\n")
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "src", "feature.go"), "package src\n")
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d err=%q", code, errOut.String())
		}
		configPath := filepath.Join(repositoryRoot, ".ddocs", "config.toml")
		configText := "docs_root = \"docs\"\nindex_file = \"README.md\"\n\n[reverse_index]\nroots = [\"src\"]\n\n[codemap]\nheadings = [\"Implementation map\"]\n"
		mustWriteAppFile(t, configPath, configText)
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r"}, &out, &errOut); code != 1 || !strings.Contains(out.String(), "check failed") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseFlagErrorsWhenNoConfiguredCodemapSectionExists(t *testing.T) {
	repositoryRoot := t.TempDir()
	mustMakeDir(t, filepath.Join(repositoryRoot, "docs"))
	mustMakeDir(t, filepath.Join(repositoryRoot, "src"))
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "docs", "feature.md"), "# Feature\n\nNo implementation map yet.\n")
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "src", "feature.go"), "package src\n")
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d err=%q", code, errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r", "--reverse-root", "src"}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "no codemap section found") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseFlagErrorsWhenCodemapSectionHasNoTargets(t *testing.T) {
	repositoryRoot := t.TempDir()
	mustMakeDir(t, filepath.Join(repositoryRoot, "docs"))
	mustMakeDir(t, filepath.Join(repositoryRoot, "src"))
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "docs", "feature.md"), "# Feature\n\n## Code map\n\nNothing mapped yet.\n")
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d err=%q", code, errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r", "--reverse-root", "src"}, &out, &errOut); code != 2 || !strings.Contains(errOut.String(), "contains no code targets") {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}

func TestReverseCheckFailsForUnresolvedScopedTarget(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	mustMakeDir(t, filepath.Join(repositoryRoot, "src"))
	mustMakeDir(t, docsRoot)
	mustWriteAppFile(t, filepath.Join(docsRoot, "feature.md"), "# Feature\n\n## Code map\n\n- `src/missing.go`\n")

	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-r", "--reverse-root", "src"}, &out, &errOut); code != 1 {
			t.Fatalf("check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		for _, expected := range []string{"ddocs check failed", "diagnostic:", "src/missing.go"} {
			if !strings.Contains(out.String(), expected) {
				t.Fatalf("check output omitted %q: %q", expected, out.String())
			}
		}
	})
}

func TestReverseCheckReportsDeterministicOrphansWithoutFixFailure(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	servicesRoot := filepath.Join(repositoryRoot, "services")
	mustMakeDir(t, docsRoot)
	mustWriteAppFile(t, filepath.Join(docsRoot, "feature.md"), "# Feature\n\n## Code map\n\n- `services/api/referenced.go`\n")
	mustWriteAppFile(t, filepath.Join(servicesRoot, "api", "referenced.go"), "package api\n")
	mustWriteAppFile(t, filepath.Join(servicesRoot, "api", "z-unreferenced.go"), "package api\n")
	mustWriteAppFile(t, filepath.Join(servicesRoot, "api", "a-unreferenced.go"), "package api\n")
	mustWriteAppFile(t, filepath.Join(servicesRoot, "api", "README.md"), "# API\n")
	mustWriteAppFile(t, filepath.Join(servicesRoot, "api", ".docignore"), "generated/\n")
	mustWriteAppFile(t, filepath.Join(servicesRoot, "api", "generated", "client.go"), "package generated\n")
	mustWriteAppFile(t, filepath.Join(repositoryRoot, "client", "outside.go"), "package client\n")
	mustWriteAppFile(t, filepath.Join(repositoryRoot, ".ddocs", "config.toml"), "docs_root = \"docs\"\nindex_file = \"README.md\"\n\n[reverse_index]\nroots = [\"services\"]\n\n[codemap]\nheadings = [\"Code map\"]\n")

	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--reverse"}, &out, &errOut); code != 0 {
			t.Fatalf("fix code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if strings.Contains(out.String(), "Reverse-index orphan:") {
			t.Fatalf("fix reported check-only orphan status: %q", out.String())
		}

		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--reverse"}, &out, &errOut); code != 1 {
			t.Fatalf("check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		want := "ddocs check failed\nmessage: Reverse-index orphan: services/api/a-unreferenced.go\nmessage: Reverse-index orphan: services/api/z-unreferenced.go\n"
		if out.String() != want {
			t.Fatalf("check output=%q want %q", out.String(), want)
		}
	})
}

func TestReverseFlagAppearsInStandardCommandHelp(t *testing.T) {
	for _, command := range []string{"check", "fix", "watch"} {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{command, "--help"}, &out, &errOut); code != 0 {
			t.Fatalf("%s code=%d err=%q", command, code, errOut.String())
		}
		for _, text := range []string{"-r, --reverse", "--reverse-root PATH", "--codemap-heading TEXT", "check reports eligible in-scope code files with no resolved authored file target"} {
			if !strings.Contains(out.String(), text) {
				t.Fatalf("%s help omitted %q:\n%s", command, text, out.String())
			}
		}
	}
}

func mustMakeDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteAppFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
