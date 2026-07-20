package documentpolicy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/filetxn"
)

func WriteBuiltinSchemas(repoRoot string, cfg config.Format, force bool) ([]string, error) {
	directory := resolveDir(repoRoot, cfg.SchemaDir)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	builtins := BuiltinSchemas()
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	var written []string
	for _, name := range names {
		path := filepath.Join(directory, name+".toml")
		changed, err := writeTransactionalFile(path, []byte(builtins[name]), force)
		if err != nil {
			return written, err
		}
		if changed {
			written = append(written, path)
		}
	}
	return written, nil
}

func writeTransactionalFile(path string, next []byte, overwrite bool) (bool, error) {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return false, fmt.Errorf("refusing to overwrite symbolic link: %s", path)
		}
		if !info.Mode().IsRegular() {
			return false, fmt.Errorf("refusing to overwrite non-regular file: %s", path)
		}
		if !overwrite {
			return false, nil
		}
		current, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		if string(current) == string(next) {
			return false, nil
		}
		_, err = filetxn.Apply([]filetxn.Rewrite{filetxn.New(path, current, next)})
		return err == nil, err
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return false, err
	}
	if _, err := file.Write(next); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return false, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return false, err
	}
	return true, nil
}
