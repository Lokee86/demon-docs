package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirectoryName = ".ddocs"
	ConfigName    = "config.toml"
)

type Location struct {
	Root       string
	ConfigPath string
}

func Discover(start string) (Location, bool) {
	root, err := startingDirectory(start)
	if err != nil {
		return Location{}, false
	}
	for {
		configPath := filepath.Join(root, DirectoryName, ConfigName)
		if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
			return Location{Root: root, ConfigPath: configPath}, true
		}
		parent := filepath.Dir(root)
		if parent == root {
			return Location{}, false
		}
		root = parent
	}
}

func FindMarker(start string) (string, bool) {
	root, err := startingDirectory(start)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(root, DirectoryName)); err == nil {
			return root, true
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", false
		}
		root = parent
	}
}

func ResolveDocsRoot(repoRoot, value string) (relative, absolute string, err error) {
	if strings.TrimSpace(value) == "" {
		return "", "", fmt.Errorf("docs root cannot be empty")
	}
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		return "", "", err
	}
	absolute = value
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(repoRoot, absolute)
	}
	absolute, err = filepath.Abs(absolute)
	if err != nil {
		return "", "", err
	}
	relative, err = filepath.Rel(repoRoot, absolute)
	if err != nil {
		return "", "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("docs root must be inside repository root: %s", value)
	}
	if err := validateRealContainment(repoRoot, absolute); err != nil {
		return "", "", err
	}
	return filepath.ToSlash(relative), absolute, nil
}

func Initialize(repoRoot, configText string) (string, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", err
	}
	marker := filepath.Join(repoRoot, DirectoryName)
	if err := os.Mkdir(marker, 0o755); err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("demon-docs repository already exists: %s", marker)
		}
		return "", err
	}
	configPath := filepath.Join(marker, ConfigName)
	if err := os.WriteFile(configPath, []byte(configText), 0o644); err != nil {
		_ = os.Remove(marker)
		return "", err
	}
	return configPath, nil
}

func RootForConfig(configPath string) (string, bool) {
	configPath = filepath.Clean(configPath)
	if filepath.Base(configPath) != ConfigName {
		return "", false
	}
	marker := filepath.Dir(configPath)
	if filepath.Base(marker) != DirectoryName {
		return "", false
	}
	return filepath.Dir(marker), true
}

func startingDirectory(start string) (string, error) {
	path, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}
	return filepath.Clean(path), nil
}
