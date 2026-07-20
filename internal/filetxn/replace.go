package filetxn

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func Replace(path string, data []byte, mode fs.FileMode) (err error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".ddocs-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		if temporaryPath != "" {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(mode.Perm()); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("preserve permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := atomicReplace(temporaryPath, path); err != nil {
		return err
	}
	temporaryPath = ""
	return nil
}
