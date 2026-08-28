package reconcile

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/validationcache"
)

// TreeScopedWithIgnoreRoot plans index updates only for folders affected by the
// supplied file-path batch. The repository tree remains the source of truth for
// cross-folder transitions and deterministic descriptions.
func TreeScopedWithIgnoreRoot(root, ignoreRoot string, c config.Config, changedPaths []string) (model.ReconcileResult, error) {
	folders, err := affectedIndexFolders(root, c, changedPaths)
	if err != nil {
		return model.ReconcileResult{}, err
	}
	if len(folders) == 0 {
		return model.ReconcileResult{}, nil
	}
	return treeWithIgnoreRoot(root, ignoreRoot, c, folders)
}

// ConvergeScopedWithin applies only affected folder-index updates until those
// folders are stable. Writes retain the same serial, stale-safe publication path
// as full convergence.
func ConvergeScopedWithin(root, ignoreRoot string, c config.Config, changedPaths []string) (model.ReconcileResult, int, error) {
	folders, err := affectedIndexFolders(root, c, changedPaths)
	if err != nil {
		return model.ReconcileResult{}, 0, err
	}
	if len(folders) == 0 {
		return model.ReconcileResult{}, 0, nil
	}

	changed := 0
	messages := []string{}
	var cache *validationcache.Store
	for pass := 0; pass < maxIndexConvergencePasses; pass++ {
		result, err := treeWithIgnoreRoot(root, ignoreRoot, c, folders)
		if err != nil {
			return result, changed, err
		}
		messages = append(messages, result.Messages...)
		if len(result.Updates) == 0 {
			result.Messages = messages
			if cache != nil {
				if err := cache.Save(); err != nil {
					return result, changed, fmt.Errorf("save validation cache after scoped index rewrites: %w", err)
				}
			}
			return result, changed, nil
		}
		if cache == nil {
			cache, err = validationcache.Open(ignoreRoot)
			if err != nil {
				return result, changed, fmt.Errorf("open validation cache for scoped index rewrites: %w", err)
			}
		}
		count, err := applyWithinWithValidationCache(result, root, ignoreRoot, cache)
		if err != nil {
			return result, changed, err
		}
		changed += count
	}
	result, err := treeWithIgnoreRoot(root, ignoreRoot, c, folders)
	if err != nil {
		return result, changed, err
	}
	messages = append(messages, result.Messages...)
	result.Messages = messages
	if cache != nil {
		if err := cache.Save(); err != nil {
			return result, changed, fmt.Errorf("save validation cache after scoped index rewrites: %w", err)
		}
	}
	if len(result.Updates) > 0 {
		return result, changed, fmt.Errorf("scoped index reconciliation did not converge after %d passes", maxIndexConvergencePasses)
	}
	return result, changed, nil
}

func affectedIndexFolders(root string, c config.Config, changedPaths []string) (map[string]bool, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absoluteRoot = filepath.Clean(absoluteRoot)
	folders := map[string]bool{}
	for _, changedPath := range changedPaths {
		path, err := filepath.Abs(changedPath)
		if err != nil {
			return nil, err
		}
		path = filepath.Clean(path)
		if !repository.Contains(absoluteRoot, path) {
			continue
		}
		folder := filepath.Dir(path)
		if filepath.Base(folder) == c.Draft.Folder {
			folder = filepath.Dir(folder)
		}
		addAffectedFolder(folders, absoluteRoot, folder)
		addAffectedFolder(folders, absoluteRoot, filepath.Dir(folder))
	}
	return folders, nil
}

func addAffectedFolder(folders map[string]bool, root, folder string) {
	folder = filepath.Clean(folder)
	if !repository.Contains(root, folder) {
		return
	}
	folders[pathOrderKey(folder)] = true
}
