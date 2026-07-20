package filetxn

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type pendingRewrite struct {
	rewrite Rewrite
	path    string
	mode    fs.FileMode
}

func Apply(rewrites []Rewrite) ([]Suppression, error) {
	pending := make([]pendingRewrite, 0, len(rewrites))
	seen := make(map[string]struct{}, len(rewrites))
	for index, rewrite := range rewrites {
		if !rewrite.prepared {
			return nil, fmt.Errorf("rewrite %d was not created by filetxn.New", index)
		}
		if rewrite.path == "" || rewrite.path == "." {
			return nil, fmt.Errorf("rewrite %d has an empty path", index)
		}
		path := filepath.Clean(rewrite.path)
		key := pathKey(path)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("rewrite batch contains duplicate path: %s", path)
		}
		seen[key] = struct{}{}
		if Digest(rewrite.oldData) != rewrite.expectedOldSHA256 {
			return nil, fmt.Errorf("rewrite %s has inconsistent old hash", path)
		}
		if Digest(rewrite.newData) != rewrite.expectedNewSHA256 {
			return nil, fmt.Errorf("rewrite %s has inconsistent new hash", path)
		}
		pending = append(pending, pendingRewrite{rewrite: rewrite, path: path})
	}

	preflightErrors := runWorkers(len(pending), func(index int) error {
		return preflight(&pending[index])
	})
	for _, err := range preflightErrors {
		if err != nil {
			return nil, err
		}
	}

	suppressions := make([]Suppression, len(pending))
	attempted := make([]int, 0, len(pending))
	for index := range pending {
		item := &pending[index]
		if err := preflight(item); err != nil {
			return nil, applyFailure(pending, attempted, err)
		}
		attempted = append(attempted, index)
		if err := Replace(item.path, item.rewrite.newData, item.mode); err != nil {
			return nil, applyFailure(pending, attempted, fmt.Errorf("apply rewrite %s: %w", item.path, err))
		}
		current, err := os.ReadFile(item.path)
		if err != nil {
			return nil, applyFailure(pending, attempted, fmt.Errorf("verify rewrite %s: %w", item.path, err))
		}
		if actual := Digest(current); actual != item.rewrite.expectedNewSHA256 {
			return nil, applyFailure(pending, attempted, fmt.Errorf("rewrite new hash mismatch %s: expected %s, got %s", item.path, item.rewrite.expectedNewSHA256, actual))
		}
		suppressions[index] = Suppression{
			Path:              item.path,
			ExpectedOldSHA256: item.rewrite.expectedOldSHA256,
			ExpectedNewSHA256: item.rewrite.expectedNewSHA256,
		}
	}
	return suppressions, nil
}

func preflight(item *pendingRewrite) error {
	info, err := os.Stat(item.path)
	if err != nil {
		return fmt.Errorf("stat rewrite source %s: %w", item.path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("rewrite source is not a regular file: %s", item.path)
	}
	current, err := os.ReadFile(item.path)
	if err != nil {
		return fmt.Errorf("read rewrite source %s: %w", item.path, err)
	}
	if actual := Digest(current); actual != item.rewrite.expectedOldSHA256 {
		return fmt.Errorf("rewrite source changed before apply %s: expected %s, got %s", item.path, item.rewrite.expectedOldSHA256, actual)
	}
	item.mode = info.Mode()
	return nil
}
