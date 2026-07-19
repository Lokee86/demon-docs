package regression_test

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

type regressionFixture struct {
	name  string
	setup func(*testing.T, string)
}

type processResult struct {
	code           int
	stdout, stderr string
}

func TestGoCLIRegressionMatrix(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	repo := filepath.Dir(filepath.Dir(file))
	bin := filepath.Join(t.TempDir(), "ddocs")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/ddocs")
	build.Dir = repo
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Go CLI: %v\n%s", err, output)
	}

	for _, fixture := range regressionFixtures() {
		t.Run(fixture.name, func(t *testing.T) {
			base := t.TempDir()
			project := filepath.Join(base, "project")
			fixture.setup(t, project)
			env := replaceEnv(os.Environ(), "XDG_CONFIG_HOME", filepath.Join(base, "empty-xdg"))

			firstFix := runProcess(t, project, env, bin, "fix")
			requireSuccess(t, "first fix", firstFix)
			afterFirstFix := snapshot(t, project)

			check := runProcess(t, project, env, bin, "check")
			requireReconciledCheck(t, check)

			secondFix := runProcess(t, project, env, bin, "fix")
			requireSuccess(t, "second fix", secondFix)
			afterSecondFix := snapshot(t, project)
			compareSnapshots(t, "first fix", afterFirstFix, "second fix", afterSecondFix)
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
	return processResult{code: code, stdout: out.String(), stderr: stderr.String()}
}

func requireSuccess(t *testing.T, command string, result processResult) {
	t.Helper()
	if result.code != 0 {
		t.Fatalf("%s failed with code %d\nstdout=%q\nstderr=%q", command, result.code, result.stdout, result.stderr)
	}
}

func requireReconciledCheck(t *testing.T, result processResult) {
	t.Helper()
	if result.code == 0 {
		return
	}
	if result.code != 1 || result.stderr != "" {
		t.Fatalf("check failed with code %d\nstdout=%q\nstderr=%q", result.code, result.stdout, result.stderr)
	}
	lines := strings.Split(strings.TrimSpace(result.stdout), "\n")
	if len(lines) < 2 || lines[0] != "ddocs check failed" {
		t.Fatalf("unexpected check failure output: %q", result.stdout)
	}
	for _, line := range lines[1:] {
		if !strings.HasPrefix(line, "message: Orphan document: ") {
			t.Fatalf("check found reconciliation drift: %q", result.stdout)
		}
	}
}

func compareSnapshots(t *testing.T, leftName string, left map[string][]byte, rightName string, right map[string][]byte) {
	t.Helper()
	keys := make([]string, 0, len(left)+len(right))
	seen := map[string]bool{}
	for path := range left {
		seen[path] = true
	}
	for path := range right {
		seen[path] = true
	}
	for path := range seen {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	for _, path := range keys {
		leftData, leftOK := left[path]
		rightData, rightOK := right[path]
		if leftOK != rightOK {
			t.Fatalf("file presence mismatch at %s: %s=%t %s=%t", path, leftName, leftOK, rightName, rightOK)
		}
		if !bytes.Equal(leftData, rightData) {
			t.Fatalf("byte mismatch at %s\n%s", path, firstByteDifference(leftName, leftData, rightName, rightData))
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

func firstByteDifference(leftName string, left []byte, rightName string, right []byte) string {
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
	return fmt.Sprintf("first difference=%d %s_len=%d %s_len=%d\n%s=%q\n%s=%q", at, leftName, len(left), rightName, len(right), leftName, left, rightName, right)
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
