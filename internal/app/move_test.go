package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIMoveWorksWithoutInitializationAndSupportsDryRun(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(cwd string) {
		writeTestFile(t, filepath.Join(cwd, "docs", "guide.md"), "# Guide\n")
		writeTestFile(t, filepath.Join(cwd, "docs", "index.md"), "[Guide](guide.md)\n")
		if err := os.MkdirAll(filepath.Join(cwd, "docs", "manual"), 0o755); err != nil {
			t.Fatal(err)
		}

		var stdout, stderr bytes.Buffer
		args := []string{"mv", "--dry-run", "docs/guide.md", "docs/manual/guide.md"}
		if code := Run(context.Background(), args, &stdout, &stderr); code != 0 {
			t.Fatalf("dry-run code=%d stderr=%q", code, stderr.String())
		}
		for _, want := range []string{"move: docs/guide.md -> docs/manual/guide.md", "update 1 Markdown file(s)", "rewrite 1 link(s)"} {
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("dry-run missing %q:\n%s", want, stdout.String())
			}
		}
		if _, err := os.Stat(filepath.Join(cwd, "docs", "guide.md")); err != nil {
			t.Fatalf("dry-run moved source: %v", err)
		}
		if _, err := os.Stat(filepath.Join(cwd, ".ddocs")); !os.IsNotExist(err) {
			t.Fatalf("dry-run created .ddocs: %v", err)
		}

		stdout.Reset()
		stderr.Reset()
		args = []string{"mv", "docs/guide.md", "docs/manual/guide.md"}
		if code := Run(context.Background(), args, &stdout, &stderr); code != 0 {
			t.Fatalf("move code=%d stderr=%q", code, stderr.String())
		}
		for _, want := range []string{"moved: docs/guide.md -> docs/manual/guide.md", "updated 1 Markdown file(s)", "rewrote 1 link(s)"} {
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("move missing %q:\n%s", want, stdout.String())
			}
		}
		if got := readTestFile(t, filepath.Join(cwd, "docs", "index.md")); got != "[Guide](manual/guide.md)\n" {
			t.Fatalf("index=%q", got)
		}
		if _, err := os.Stat(filepath.Join(cwd, ".ddocs")); !os.IsNotExist(err) {
			t.Fatalf("move created .ddocs: %v", err)
		}
	})
}

func TestCLIMoveRequiresTwoPaths(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"mv", "one.md"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "SOURCE and DESTINATION are required") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
