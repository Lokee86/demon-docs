package parity_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPythonGoCLIContractParity(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Fatal("python is required for the CLI parity release gate")
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

	cases := []struct {
		name string
		args []string
	}{
		{name: "top help", args: []string{"--help"}},
		{name: "fix help", args: []string{"fix", "--help"}},
		{name: "check help", args: []string{"check", "--help"}},
		{name: "watch help", args: []string{"watch", "--help"}},
		{name: "config help", args: []string{"config", "--help"}},
		{name: "config paths help", args: []string{"config", "paths", "--help"}},
		{name: "config show help", args: []string{"config", "show", "--help"}},
		{name: "config init help", args: []string{"config", "init", "--help"}},
		{name: "version", args: []string{"--version"}},
		{name: "missing command"},
		{name: "unknown command", args: []string{"wat"}},
		{name: "unexpected positional", args: []string{"fix", "extra"}},
		{name: "unknown option", args: []string{"fix", "--bogus"}},
		{name: "missing option value", args: []string{"fix", "--root"}},
		{name: "invalid float", args: []string{"watch", "--debounce-seconds", "nope"}},
		{name: "invalid boolean argument", args: []string{"fix", "--parent-link-folder-indexes=nope"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cwd := t.TempDir()
			env := replaceEnv(os.Environ(), "XDG_CONFIG_HOME", filepath.Join(cwd, "empty-xdg"))
			pythonResult := runProcess(t, cwd, env, python, append([]string{filepath.Join(repo, "main.py")}, tc.args...)...)
			goResult := runProcess(t, cwd, env, bin, tc.args...)
			compareProcess(t, tc.name, pythonResult, goResult)
		})
	}
}
