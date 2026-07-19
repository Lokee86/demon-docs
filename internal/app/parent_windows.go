//go:build windows

package app

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func parentAlive(pid int) bool {
	if pid <= 0 {
		return true
	}
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH").Output()
	if err != nil {
		return true
	}
	return strings.Contains(string(out), fmt.Sprintf(" %d ", pid)) || strings.Contains(string(out), fmt.Sprintf("%d ", pid))
}
