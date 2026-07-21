//go:build windows

package app

import "golang.org/x/sys/windows"

func parentAlive(pid int) bool {
	if pid <= 0 {
		return true
	}

	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return err != windows.ERROR_INVALID_PARAMETER
	}
	defer windows.CloseHandle(handle)

	status, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return true
	}

	switch status {
	case uint32(windows.WAIT_TIMEOUT):
		return true
	case windows.WAIT_OBJECT_0:
		return false
	default:
		return true
	}
}
