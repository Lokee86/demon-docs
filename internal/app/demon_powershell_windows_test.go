//go:build windows

package app

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
)

func TestPowerShellHookBootstrapInstallsFunctions(t *testing.T) {
	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Skip("Windows PowerShell is unavailable")
	}
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"demon", "__shell-hook", "powershell"}, &out, &errOut); code != 0 {
		t.Fatalf("hook code=%d stderr=%q", code, errOut.String())
	}
	bootstrap := strings.TrimSpace(out.String())
	if strings.Contains(bootstrap, "\n") {
		t.Fatalf("native-command output split into multiple PowerShell objects: %q", bootstrap)
	}
	command := bootstrap + `; if (-not (Get-Command Invoke-DdocsDemonHook -CommandType Function -ErrorAction SilentlyContinue)) { exit 17 }; if (-not (Get-Command Leave-DdocsDemon -CommandType Function -ErrorAction SilentlyContinue)) { exit 18 }`
	result, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-Command", command).CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell could not install generated hook: %v\n%s", err, result)
	}
}
