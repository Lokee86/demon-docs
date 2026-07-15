package parity_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestPythonGoParity(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("python is unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	repo := filepath.Dir(filepath.Dir(file))
	bin := filepath.Join(t.TempDir(), "doc-ledger")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/doc-ledger")
	build.Dir = repo
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Go CLI: %v\n%s", err, output)
	}

	base := t.TempDir()
	pyProject := filepath.Join(base, "python")
	goProject := filepath.Join(base, "go")
	for _, project := range []string{pyProject, goProject} {
		makeFixture(t, project)
	}
	env := append(os.Environ(), "XDG_CONFIG_HOME="+filepath.Join(base, "empty-xdg"))
	py := run(t, pyProject, env, python, filepath.Join(repo, "main.py"), "fix")
	goResult := run(t, goProject, env, bin, "fix")
	if py.code != goResult.code || py.stdout != goResult.stdout || py.stderr != goResult.stderr {
		t.Fatalf("fix process mismatch\npython=%+v\ngo=%+v", py, goResult)
	}
	compareTrees(t, pyProject, goProject)
	py = run(t, pyProject, env, python, filepath.Join(repo, "main.py"), "check")
	goResult = run(t, goProject, env, bin, "check")
	if py.code != goResult.code || py.stdout != goResult.stdout || py.stderr != goResult.stderr {
		t.Fatalf("check process mismatch\npython=%+v\ngo=%+v", py, goResult)
	}
}

type processResult struct {
	code           int
	stdout, stderr string
}

func run(t *testing.T, dir string, env []string, name string, args ...string) processResult {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return processResult{code, strings.ReplaceAll(out.String(), "\r\n", "\n"), strings.ReplaceAll(stderr.String(), "\r\n", "\n")}
}
func makeFixture(t *testing.T, project string) {
	t.Helper()
	write(t, filepath.Join(project, ".doc-ledger.toml"), `root = "docs"
index_file = "!README.md"
[markers]
prefix = "navmark"
[parent_link]
folder_indexes = true
indexed_files = true
[drafts]
folder = "_drafts"
[files]
include_patterns = ["**/*.md", "**/*.png"]
`)
	write(t, filepath.Join(project, "docs", "README.md"), "# Documentation\n\nUser preface.\n\n## Top-Level Files\n- [page.md](page.md) - Custom page.\n\n## Top-Level Folders\n- [Guide](guide/!README.md) - Custom guide.\n\n## Notes\nKeep this.\n")
	write(t, filepath.Join(project, "docs", "page.md"), "# Page\n\nBody\n")
	write(t, filepath.Join(project, "docs", "_drafts", "idea.md"), "# Idea\n")
	write(t, filepath.Join(project, "docs", "guide", "topic.md"), "# Topic\n")
	writeBytes(t, filepath.Join(project, "docs", "diagram.png"), []byte{0x89, 'P', 'N', 'G', 0, 'x'})
}
func write(t *testing.T, path, text string) { writeBytes(t, path, []byte(text)) }
func writeBytes(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
func compareTrees(t *testing.T, a, b string) {
	t.Helper()
	left := snapshot(t, a)
	right := snapshot(t, b)
	keys := make([]string, 0, len(left)+len(right))
	seen := map[string]bool{}
	for k := range left {
		seen[k] = true
	}
	for k := range right {
		seen[k] = true
	}
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !bytes.Equal(canonical(left[k]), canonical(right[k])) {
			t.Fatalf("tree mismatch at %s\npython=%q\ngo=%q", k, left[k], right[k])
		}
	}
}
func snapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	result := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		data, err := os.ReadFile(path)
		if err == nil {
			result[filepath.ToSlash(rel)] = data
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func canonical(data []byte) []byte {
	return bytes.TrimRight(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")), "\n")
}
