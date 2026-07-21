package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCountsIndexLinksButNotDraftLinksForOrphans(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	files := map[string]string{
		"INDEX.md":       "# Docs\n\n[Orphan](orphan.md)\n",
		"source.md":      "# Source\n\n[Linked](linked.md)\n",
		"linked.md":      "# Linked\n\n[Source](source.md)\n",
		"orphan.md":      "# Orphan\n",
		"draft-only.md":  "# Draft only\n",
		"stubs/draft.md": "# Draft\n\n[Draft only](../draft-only.md)\n",
	}
	for relative, text := range files {
		path := filepath.Join(docsRoot, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	withWorkingDirectory(t, repositoryRoot, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}

		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "-l"}, &out, &errOut); code != 0 {
			t.Fatalf("fix code=%d out=%q err=%q", code, out.String(), errOut.String())
		}

		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "-l"}, &out, &errOut); code != 1 {
			t.Fatalf("check code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if !strings.Contains(out.String(), "message: Orphan document: docs/draft-only.md") {
			t.Fatalf("missing orphan diagnostic: %q", out.String())
		}
		for _, unexpected := range []string{"docs/orphan.md", "docs/source.md", "docs/linked.md", "docs/stubs/draft.md", "docs/INDEX.md"} {
			if strings.Contains(out.String(), "Orphan document: "+unexpected) {
				t.Fatalf("unexpected orphan %s: %q", unexpected, out.String())
			}
		}
	})
}
