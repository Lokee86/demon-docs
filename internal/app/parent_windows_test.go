//go:build windows

package app

import (
	"os"
	"os/exec"
	"testing"
)

func TestParentAliveCurrentProcess(t *testing.T) {
	if !parentAlive(os.Getpid()) {
		t.Fatal("current process reported dead")
	}
}

func TestParentAliveExitedProcess(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child process: %v", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for child process: %v", err)
	}

	if parentAlive(pid) {
		t.Fatal("exited process reported alive")
	}
}
