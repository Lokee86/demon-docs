//go:build windows

package frontmatter

import "golang.org/x/sys/windows"

func atomicReplace(tempPath, destinationPath string) error {
	temp, err := windows.UTF16PtrFromString(tempPath)
	if err != nil {
		return err
	}
	destination, err := windows.UTF16PtrFromString(destinationPath)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(temp, destination, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
