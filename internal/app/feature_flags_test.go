package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexesAndLinksCanRunSeparately(t *testing.T) {
	t.Run("links only", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "page.md"), "[asset](asset.bin)\n")
		writeTestFile(t, filepath.Join(root, "asset.bin"), "asset")
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--root", root, "-l"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if _, err := os.Stat(filepath.Join(root, "README.md")); !os.IsNotExist(err) {
			t.Fatalf("links-only run created an index: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, ".ddocs", "links.json")); err != nil {
			t.Fatalf("links-only run did not write link state: %v", err)
		}
	})

	t.Run("indexes only", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "page.md"), "# Page\n")
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "--root", root, "--indexes"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if _, err := os.Stat(filepath.Join(root, "README.md")); err != nil {
			t.Fatalf("indexes-only run did not create an index: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, ".ddocs", "links.json")); !os.IsNotExist(err) {
			t.Fatalf("indexes-only run wrote link state: %v", err)
		}
	})
}

func TestLinksOnlyDoesNotRequireDocsRoot(t *testing.T) {
	repositoryRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repositoryRoot, ".ddocs", "config.toml"), "docs_root = \"missing-docs\"\n")
	writeTestFile(t, filepath.Join(repositoryRoot, "README.md"), "# Repository\n")
	withWorkingDirectory(t, repositoryRoot, func(string) {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"fix", "-l"}, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	if _, err := os.Stat(filepath.Join(repositoryRoot, ".ddocs", "links.json")); err != nil {
		t.Fatalf("links-only run did not initialize state: %v", err)
	}
}

func TestWatchOnceHonorsLinksOnly(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "page.md"), "[asset](asset.bin)\n")
	writeTestFile(t, filepath.Join(root, "asset.bin"), "asset")
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"watch", "--root", root, "-l", "--once"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "README.md")); !os.IsNotExist(err) {
		t.Fatalf("links-only watch created an index: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".ddocs", "files.json")); err != nil {
		t.Fatalf("links-only watch did not write file state: %v", err)
	}
}
