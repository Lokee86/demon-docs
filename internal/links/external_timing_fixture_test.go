package links

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type copiedFixtureDirectory struct {
	path string
	info fs.FileInfo
}

func copyExternalFixture(t *testing.T, source, destination string) {
	t.Helper()
	directories := make([]copiedFixtureDirectory, 0)
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() && path != source && (entry.Name() == ".git" || entry.Name() == ".obsidian") {
			return filepath.SkipDir
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := destination
		if relative != "." {
			target = filepath.Join(destination, relative)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if err := os.MkdirAll(target, info.Mode().Perm()); err != nil {
				return err
			}
			directories = append(directories, copiedFixtureDirectory{path: target, info: info})
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, info.Mode().Perm()); err != nil {
			return err
		}
		return os.Chtimes(target, info.ModTime(), info.ModTime())
	})
	if err != nil {
		t.Fatal(err)
	}
	for index := len(directories) - 1; index >= 0; index-- {
		directory := directories[index]
		if err := os.Chmod(directory.path, directory.info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(directory.path, directory.info.ModTime(), directory.info.ModTime()); err != nil {
			t.Fatal(err)
		}
	}
}
