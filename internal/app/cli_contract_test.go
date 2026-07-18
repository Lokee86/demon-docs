package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIFileAndRenderingOverrides(t *testing.T) {
	tests := []struct {
		name, file string
		args       []string
		wantIndex  string
		contains   []string
		excludes   []string
	}{
		{"index_file", "page.md", []string{"--index-file", "!README.md"}, "!README.md", []string{"[page.md](page.md)"}, nil},
		{"draft_folder", "ideas/draft.md", []string{"--draft-folder", "ideas"}, "README.md", []string{"[draft.md](ideas/draft.md)"}, nil},
		{"include", "reference.pdf", []string{"--include", "**/*.pdf"}, "README.md", []string{"[reference.pdf](reference.pdf)"}, nil},
		{"exclude", "private.md", []string{"--exclude", "**/private.md"}, "README.md", nil, []string{"private.md"}},
		{"marker", "page.md", []string{"--marker-prefix", "nav"}, "README.md", []string{"<!-- nav:files:start -->"}, []string{"<!-- doc-ledger:files:start -->"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeTestFile(t, filepath.Join(root, filepath.FromSlash(test.file)), "# Fixture\n")
			args := append([]string{"fix", "--root", root}, test.args...)
			var stdout, stderr bytes.Buffer
			if code := Run(context.Background(), args, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%q", code, stderr.String())
			}
			data, err := os.ReadFile(filepath.Join(root, test.wantIndex))
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			for _, value := range test.contains {
				if !strings.Contains(text, value) {
					t.Errorf("index missing %q:\n%s", value, text)
				}
			}
			for _, value := range test.excludes {
				if strings.Contains(text, value) {
					t.Errorf("index unexpectedly contains %q:\n%s", value, text)
				}
			}
		})
	}
}

func TestCLIParentLabelAndBooleanOverrides(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	writeTestFile(t, filepath.Join(root, "page.md"), "# Page\n")
	writeTestFile(t, filepath.Join(root, "guide", "topic.md"), "# Topic\n")
	var stdout, stderr bytes.Buffer
	args := []string{"fix", "--root", root, "--parent-label", "Up", "--parent-link-indexed-files", "--no-parent-link-folder-indexes"}
	if code := Run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	page := readTestFile(t, filepath.Join(root, "page.md"))
	guide := readTestFile(t, filepath.Join(root, "guide", "README.md"))
	if !strings.Contains(page, "Up: [Docs](./README.md)") {
		t.Fatal(page)
	}
	if strings.Contains(guide, "Up:") || strings.Contains(guide, "Parent index:") {
		t.Fatal(guide)
	}
}

func TestConfigCommandsContract(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(cwd string) {
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(cwd, "xdg"))
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), []string{"config", "paths"}, &stdout, &stderr); code != 0 {
			t.Fatalf("paths code=%d stderr=%q", code, stderr.String())
		}
		for _, value := range []string{"cwd = " + cwd, filepath.Join(cwd, ".demon-docs.toml"), filepath.Join(cwd, "demon-docs.toml"), filepath.Join(cwd, ".doc-ledger.toml"), "selected local config = <none>", filepath.Join(cwd, "xdg", "demon-docs", "config.toml"), filepath.Join(cwd, "xdg", "doc-ledger", "config.toml"), "selected config = <none>"} {
			if !strings.Contains(stdout.String(), value) {
				t.Errorf("paths missing %q:\n%s", value, stdout.String())
			}
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"config", "show", "--no-local-config", "--no-global-config"}, &stdout, &stderr); code != 0 {
			t.Fatalf("show code=%d stderr=%q", code, stderr.String())
		}
		for _, value := range []string{"selected_config_path = <built-in defaults>", "root = 'docs'", "index_file = 'README.md'", "folder_indexes = true", "indexed_files = false"} {
			if !strings.Contains(stdout.String(), value) {
				t.Errorf("show missing %q:\n%s", value, stdout.String())
			}
		}

		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"config", "init", "--local"}, &stdout, &stderr); code != 0 {
			t.Fatalf("init code=%d stderr=%q", code, stderr.String())
		}
		path := filepath.Join(cwd, ".demon-docs.toml")
		if strings.TrimSpace(stdout.String()) != path {
			t.Fatalf("init output=%q want=%q", stdout.String(), path)
		}
		original := readTestFile(t, path)
		if code := Run(context.Background(), []string{"config", "init", "--local"}, &bytes.Buffer{}, &stderr); code != 2 || !strings.Contains(stderr.String(), "already exists") {
			t.Fatalf("overwrite code=%d stderr=%q", code, stderr.String())
		}
		writeTestFile(t, path, "changed")
		stderr.Reset()
		if code := Run(context.Background(), []string{"config", "init", "--local", "--force"}, &bytes.Buffer{}, &stderr); code != 0 || readTestFile(t, path) != original {
			t.Fatalf("force code=%d stderr=%q", code, stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run(context.Background(), []string{"config", "init", "--global"}, &stdout, &stderr); code != 0 {
			t.Fatalf("global init code=%d stderr=%q", code, stderr.String())
		}
		globalPath := filepath.Join(cwd, "xdg", "demon-docs", "config.toml")
		if strings.TrimSpace(stdout.String()) != globalPath {
			t.Fatalf("global init output=%q want=%q", stdout.String(), globalPath)
		}
	})
}

func withWorkingDirectory(t *testing.T, directory string, run func(string)) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	}()
	run(directory)
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
