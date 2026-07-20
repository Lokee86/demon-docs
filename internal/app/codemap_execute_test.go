package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func TestCodemapExecutionHelpAndRequiredRoots(t *testing.T) {
	for _, command := range []string{"fix", "check", "inspect"} {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"codemap", command, "--help"}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("%s help code=%d stderr=%q", command, code, stderr.String())
		}
		for _, want := range []string{"usage: ddocs codemap " + command, "one Markdown file", "daemon never execute"} {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("%s help missing %q:\n%s", command, want, stdout.String())
			}
		}
	}
	for _, command := range []string{"check", "inspect"} {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"codemap", command}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "required: --root") {
			t.Fatalf("%s code=%d stderr=%q", command, code, stderr.String())
		}
	}

	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"codemaps", "fix", "--help"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "usage: ddocs codemap fix") {
		t.Fatalf("plural alias code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestCodemapFixDryRunCheckAndApplySingleFile(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeTestFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\nThe implementation is in `src/runtime.go`.\n\n## Code Map\n")
	writeTestFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n")
	if _, err := repository.Initialize(root, config.RepositoryStarterText("docs")); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, root, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"codemap", "check", "--root", "docs/runtime.md"}, &stdout, &stderr); code != 1 || !strings.Contains(stdout.String(), "check failed") {
			t.Fatalf("initial check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"codemap", "fix", "--root", "docs/runtime.md", "--dry-run"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "would update 1 file") {
			t.Fatalf("dry-run code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		unchanged, err := os.ReadFile(filepath.Join(docs, "runtime.md"))
		if err != nil || strings.Contains(string(unchanged), "codemap:start") {
			t.Fatalf("dry-run wrote file: %q err=%v", unchanged, err)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"codemap", "fix", "--root", "docs/runtime.md"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "updated 1 file") {
			t.Fatalf("fix code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		updated, err := os.ReadFile(filepath.Join(docs, "runtime.md"))
		if err != nil || !strings.Contains(string(updated), "- `src/runtime.go`") || !strings.Contains(string(updated), "codemap:start") {
			t.Fatalf("file was not updated: %q err=%v", updated, err)
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"codemap", "check", "--root", "docs/runtime.md"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "check passed") {
			t.Fatalf("final check code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
}
