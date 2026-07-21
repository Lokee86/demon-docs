package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type harness struct {
	root      string
	workspace string
	repo      string
	home      string
	binDir    string
	ddocs     string
	demon     string
	env       []string
}

func newHarness(root, workspace string) *harness {
	home := filepath.Join(workspace, "home")
	binDir := filepath.Join(workspace, "bin")
	repo := filepath.Join(workspace, "repository")
	_ = os.MkdirAll(filepath.Join(home, "AppData", "Roaming"), 0o755)
	_ = os.MkdirAll(filepath.Join(home, "AppData", "Local"), 0o755)
	_ = os.MkdirAll(binDir, 0o755)
	_ = os.MkdirAll(repo, 0o755)

	env := append([]string{}, os.Environ()...)
	env = setEnv(env, "HOME", home)
	env = setEnv(env, "USERPROFILE", home)
	env = setEnv(env, "APPDATA", filepath.Join(home, "AppData", "Roaming"))
	env = setEnv(env, "LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	env = setEnv(env, "XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	env = setEnv(env, "PATH", binDir+string(os.PathListSeparator)+envValue(env, "PATH"))

	return &harness{root: root, workspace: workspace, repo: repo, home: home, binDir: binDir, env: env}
}

func (h *harness) buildBinaries() error {
	h.step("build fresh binaries")
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	h.ddocs = filepath.Join(h.binDir, "ddocs"+suffix)
	h.demon = filepath.Join(h.binDir, "demon"+suffix)
	if _, err := h.command(h.root, "go", "build", "-buildvcs=false", "-o", h.ddocs, "./cmd/ddocs"); err != nil {
		return err
	}
	_, err := h.command(h.root, "go", "build", "-buildvcs=false", "-o", h.demon, "./cmd/demon")
	return err
}

func (h *harness) useBinaries(ddocs, demon string) error {
	var err error
	if h.ddocs, err = filepath.Abs(ddocs); err != nil {
		return err
	}
	if h.demon, err = filepath.Abs(demon); err != nil {
		return err
	}
	for _, path := range []string{h.ddocs, h.demon} {
		if info, statErr := os.Stat(path); statErr != nil || info.IsDir() {
			return fmt.Errorf("binary is unavailable: %s", path)
		}
	}
	h.binDir = filepath.Dir(h.ddocs)
	h.env = setEnv(h.env, "PATH", h.binDir+string(os.PathListSeparator)+envValue(h.env, "PATH"))
	return nil
}

func (h *harness) command(dir, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = h.env
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if ctx.Err() != nil {
		return text, fmt.Errorf("%s timed out", commandLine(name, args))
	}
	if err != nil {
		return text, fmt.Errorf("%s failed: %w\n%s", commandLine(name, args), err, text)
	}
	return text, nil
}

func (h *harness) commandFails(dir, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = h.env
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return fmt.Errorf("%s timed out", commandLine(name, args))
	}
	if err == nil {
		return fmt.Errorf("%s unexpectedly succeeded\n%s", commandLine(name, args), output)
	}
	return nil
}

func setEnv(env []string, key, value string) []string {
	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		name, _, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(name, key) {
			continue
		}
		result = append(result, entry)
	}
	return append(result, key+"="+value)
}

func envValue(env []string, key string) string {
	for index := len(env) - 1; index >= 0; index-- {
		name, value, ok := strings.Cut(env[index], "=")
		if ok && strings.EqualFold(name, key) {
			return value
		}
	}
	return ""
}

func (h *harness) step(label string) { fmt.Println("[smoke]", label) }

func commandLine(name string, args []string) string {
	return strings.Join(append([]string{name}, args...), " ")
}
