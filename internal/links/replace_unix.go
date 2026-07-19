//go:build !windows

package links

import "os"

func atomicReplace(tempPath, destinationPath string) error {
	return os.Rename(tempPath, destinationPath)
}
