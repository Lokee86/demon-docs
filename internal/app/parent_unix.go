//go:build !windows

package app

import "syscall"

func parentAlive(pid int) bool {
	if pid <= 0 {
		return true
	}
	return syscall.Kill(pid, 0) == nil
}
