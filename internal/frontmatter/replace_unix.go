//go:build !windows

package frontmatter

import "os"

func atomicReplace(tempPath, destinationPath string) error {
	return os.Rename(tempPath, destinationPath)
}
