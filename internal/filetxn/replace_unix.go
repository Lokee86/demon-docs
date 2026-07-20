//go:build !windows

package filetxn

import "os"

func atomicReplace(tempPath, destinationPath string) error {
	return os.Rename(tempPath, destinationPath)
}
