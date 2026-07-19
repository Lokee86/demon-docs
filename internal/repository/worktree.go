package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

// DetectLinkedWorktree finds a linked worktree without changing it. The
// primary config is returned when the linked worktree has not been bootstrapped
// yet, so read-only callers can discover it and a later mutating command can
// perform the bootstrap.
func DetectLinkedWorktree(start string) (Location, bool, error) {
	root, err := startingDirectory(start)
	if err != nil {
		return Location{}, false, err
	}
	for {
		if location, detected, err := detectLinkedWorktreeAt(root); detected || err != nil {
			return location, detected, err
		}
		parent := filepath.Dir(root)
		if parent == root {
			return Location{}, false, nil
		}
		root = parent
	}
}

// BootstrapLinkedWorktree is the sole Git-aware repository adapter. It only
// handles a linked worktree whose primary worktree is already initialized; all
// ordinary Demon Docs discovery remains independent of Git.
func BootstrapLinkedWorktree(start string) (Location, bool, error) {
	location, detected, err := DetectLinkedWorktree(start)
	if err != nil || !detected {
		return location, detected, err
	}
	localConfig := filepath.Join(location.Root, DirectoryName, ConfigName)
	if fileExists(localConfig) {
		location.ConfigPath = localConfig
		return location, true, nil
	}
	marker := filepath.Join(location.Root, DirectoryName)
	if info, statErr := os.Stat(marker); statErr == nil && !info.IsDir() {
		return Location{}, true, fmt.Errorf("linked worktree marker is not a directory: %s", marker)
	}
	if err := os.MkdirAll(marker, 0o755); err != nil {
		return Location{}, true, err
	}
	config, err := os.ReadFile(location.ConfigPath)
	if err != nil {
		return Location{}, true, err
	}
	if err := os.WriteFile(localConfig, config, 0o644); err != nil {
		return Location{}, true, err
	}
	if _, err := ddrepo.Init(marker); err != nil {
		_ = os.RemoveAll(marker)
		return Location{}, true, fmt.Errorf("initialize linked worktree object storage: %w", err)
	}
	location.ConfigPath = localConfig
	return location, true, nil
}

func detectLinkedWorktreeAt(root string) (Location, bool, error) {
	gitFile := filepath.Join(root, ".git")
	data, err := os.ReadFile(gitFile)
	if err != nil {
		return Location{}, false, nil
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(strings.ToLower(line), "gitdir:") {
		return Location{}, false, nil
	}
	gitDir := strings.TrimSpace(line[len("gitdir:"):])
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(root, gitDir)
	}
	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return Location{}, false, err
	}
	commonText, err := os.ReadFile(filepath.Join(gitDir, "commondir"))
	if err != nil {
		return Location{}, false, nil
	}
	commonGitDir := strings.TrimSpace(string(commonText))
	if !filepath.IsAbs(commonGitDir) {
		commonGitDir = filepath.Join(gitDir, commonGitDir)
	}
	commonGitDir, err = filepath.Abs(commonGitDir)
	if err != nil {
		return Location{}, false, err
	}
	primaryRoot := filepath.Dir(commonGitDir)
	primaryConfig := filepath.Join(primaryRoot, DirectoryName, ConfigName)
	if !fileExists(primaryConfig) {
		return Location{}, false, nil
	}
	marker := filepath.Join(root, DirectoryName)
	if info, statErr := os.Stat(marker); statErr == nil {
		if !info.IsDir() {
			return Location{}, true, fmt.Errorf("linked worktree marker is not a directory: %s", marker)
		}
		configPath := filepath.Join(marker, ConfigName)
		if fileExists(configPath) {
			return Location{Root: root, ConfigPath: configPath}, true, nil
		}
	}
	return Location{Root: root, ConfigPath: primaryConfig}, true, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
