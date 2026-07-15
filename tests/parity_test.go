package parity_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type parityFixture struct {
	name  string
	setup func(*testing.T, string)
}

type processResult struct {
	code           int
	stdout, stderr string
}

func TestPythonGoParityMatrix(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Fatal("python is required for the parity release gate")
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

	for _, fixture := range parityFixtures() {
		t.Run(fixture.name, func(t *testing.T) {
			base := t.TempDir()
			pyProject := filepath.Join(base, "python")
			goProject := filepath.Join(base, "go")
			fixture.setup(t, pyProject)
			fixture.setup(t, goProject)
			env := replaceEnv(os.Environ(), "XDG_CONFIG_HOME", filepath.Join(base, "empty-xdg"))

			py := runProcess(t, pyProject, env, python, filepath.Join(repo, "main.py"), "fix")
			goResult := runProcess(t, goProject, env, bin, "fix")
			compareProcess(t, "fix", py, goResult)
			compareTreesExact(t, pyProject, goProject)

			py = runProcess(t, pyProject, env, python, filepath.Join(repo, "main.py"), "check")
			goResult = runProcess(t, goProject, env, bin, "check")
			compareProcess(t, "check", py, goResult)
			compareTreesExact(t, pyProject, goProject)
		})
	}
}

func runProcess(t *testing.T, dir string, env []string, name string, args ...string) processResult {
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
			t.Fatalf("run %s: %v", name, err)
		}
	}
	return processResult{
		code:   code,
		stdout: normalizeProcessOutput(out.String(), dir),
		stderr: normalizeProcessOutput(stderr.String(), dir),
	}
}

func normalizeProcessOutput(value, projectRoot string) string {
	value = strings.ReplaceAll(value, projectRoot, "<PROJECT>")
	value = strings.ReplaceAll(value, filepath.ToSlash(projectRoot), "<PROJECT>")
	return strings.ReplaceAll(value, "\r\n", "\n")
}

func compareProcess(t *testing.T, command string, python, goResult processResult) {
	t.Helper()
	if python != goResult {
		t.Fatalf("%s process mismatch\npython=%+v\ngo=%+v", command, python, goResult)
	}
}

func compareTreesExact(t *testing.T, pythonRoot, goRoot string) {
	t.Helper()
	pythonFiles := snapshot(t, pythonRoot)
	goFiles := snapshot(t, goRoot)
	keys := make([]string, 0, len(pythonFiles)+len(goFiles))
	seen := map[string]bool{}
	for path := range pythonFiles {
		seen[path] = true
	}
	for path := range goFiles {
		seen[path] = true
	}
	for path := range seen {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	for _, path := range keys {
		pythonData, pythonOK := pythonFiles[path]
		goData, goOK := goFiles[path]
		if pythonOK != goOK {
			t.Fatalf("file presence mismatch at %s: python=%t go=%t", path, pythonOK, goOK)
		}
		if !bytes.Equal(pythonData, goData) {
			t.Fatalf("byte mismatch at %s\n%s", path, firstByteDifference(pythonData, goData))
		}
	}
}

func snapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	result := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func firstByteDifference(left, right []byte) string {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	at := limit
	for index := 0; index < limit; index++ {
		if left[index] != right[index] {
			at = index
			break
		}
	}
	return fmt.Sprintf("first difference=%d python_len=%d go_len=%d\npython=%q\ngo=%q", at, len(left), len(right), left, right)
}

func replaceEnv(env []string, key, value string) []string {
	prefix := key + "="
	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			result = append(result, entry)
		}
	}
	return append(result, prefix+value)
}
