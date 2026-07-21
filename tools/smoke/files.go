package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func writeFile(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

func appendFile(path, text string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	return err
}

func replaceFile(path, old, replacement string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)
	if !strings.Contains(text, old) {
		return fmt.Errorf("expected text not found in %s: %q", path, old)
	}
	return os.WriteFile(path, []byte(strings.Replace(text, old, replacement, 1)), 0o644)
}

func requireContains(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), expected) {
		return fmt.Errorf("%s does not contain %q", path, expected)
	}
	return nil
}

func requireNotContains(path, unexpected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), unexpected) {
		return fmt.Errorf("%s still contains %q", path, unexpected)
	}
	return nil
}

func requireMissing(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("unexpected path exists: %s", path)
}

func snapshot(root string) (map[string][32]byte, error) {
	result := map[string][32]byte{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = sha256.Sum256(data)
		return nil
	})
	return result, err
}

func equalSnapshots(left, right map[string][32]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for path, digest := range left {
		if right[path] != digest {
			return false
		}
	}
	return true
}

func waitFor(description string, timeout time.Duration, check func() bool) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s", description)
}

func fileContains(path, text string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), text)
}
