package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/demon"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func initializedDemonRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Initialize(root, config.RepositoryStarterText("docs")); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestDemonStatusIsReadOnly(t *testing.T) {
	root := initializedDemonRepo(t)
	withWorkingDirectory(t, root, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"demon", "--status"}, &out, &errOut); code != 0 {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if _, err := os.Stat(filepath.Join(root, ".ddocs", "runtime")); !os.IsNotExist(err) {
			t.Fatalf("status created runtime state: %v", err)
		}
		if !strings.Contains(out.String(), "demon: stopped") || !strings.Contains(out.String(), "active shells: 0") {
			t.Fatalf("unexpected status: %s", out.String())
		}
	})
}

func TestDisableThenEnableClearsShutdownRequest(t *testing.T) {
	root := initializedDemonRepo(t)
	withWorkingDirectory(t, root, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"demon", "run", "--false"}, &out, &errOut); code != 0 {
			t.Fatalf("disable code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		r := demon.New(root)
		if !r.ShutdownRequested() {
			t.Fatal("disable did not request shutdown")
		}
		if err := config.SetDemonRun(filepath.Join(root, ".ddocs", "config.toml"), true); err != nil {
			t.Fatal(err)
		}
		r.ClearShutdown()
		if r.ShutdownRequested() {
			t.Fatal("re-enable left shutdown request behind")
		}
	})
}

func TestShellHookUsesTokenLeaveAndValidPowerShellInstallation(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"demon", "__shell-hook", "bash"}, &out, &errOut); code != 0 {
		t.Fatal(errOut.String())
	}
	bash := out.String()
	if !strings.Contains(bash, "__ddocs_demon_token") || !strings.Contains(bash, "__leave") || !strings.Contains(bash, "claimed=") || strings.Contains(bash, "__shutdown") {
		t.Fatalf("unsafe Bash hook: %s", bash)
	}
	out.Reset()
	if code := Run(context.Background(), []string{"demon", "__shell-hook", "powershell"}, &out, &errOut); code != 0 {
		t.Fatal(errOut.String())
	}
	powershell := out.String()
	if !strings.Contains(powershell, "Invoke-Expression (& ddocs demon __shell-hook powershell)") || !strings.Contains(powershell, "__DdocsDemonToken") || !strings.Contains(powershell, "claimed=") || strings.Contains(powershell, "<(ddocs") {
		t.Fatalf("invalid PowerShell hook: %s", powershell)
	}
}
