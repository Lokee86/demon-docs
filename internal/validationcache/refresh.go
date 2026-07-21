package validationcache

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/filetxn"
)

// Surface identifies validation results invalidated by a generated rewrite.
// Callers declare only the surfaces their rewrite can affect; other clean
// results retain their identity under the published content hash.
type Surface uint8

const (
	SurfaceFrontmatter Surface = 1 << iota
	SurfaceFormat
)

// PublishedRewrite describes exact bytes before and after a successful write.
type PublishedRewrite struct {
	Path        string
	OldData     []byte
	NewData     []byte
	Invalidated Surface
}

// RefreshPublished updates durable cache identities after generated rewrites
// have successfully published their final bytes.
func RefreshPublished(repoRoot string, rewrites []PublishedRewrite) error {
	if len(rewrites) == 0 {
		return nil
	}
	store, err := Open(repoRoot)
	if err != nil {
		return fmt.Errorf("open validation cache: %w", err)
	}
	for _, rewrite := range rewrites {
		if _, err := store.RefreshPublished(repoRoot, rewrite.Path, rewrite.OldData, rewrite.NewData, rewrite.Invalidated); err != nil {
			return err
		}
	}
	if err := store.Save(); err != nil {
		return fmt.Errorf("save validation cache: %w", err)
	}
	return nil
}

// RefreshTransactions adapts prepared file transactions to the generated
// rewrite cache-refresh contract.
func RefreshTransactions(repoRoot string, rewrites []filetxn.Rewrite, invalidated Surface) error {
	published := make([]PublishedRewrite, len(rewrites))
	for index, rewrite := range rewrites {
		published[index] = PublishedRewrite{
			Path:        rewrite.Path(),
			OldData:     rewrite.OldData(),
			NewData:     rewrite.NewData(),
			Invalidated: invalidated,
		}
	}
	return RefreshPublished(repoRoot, published)
}

// RefreshPublished updates one in-memory cache entry when it still describes
// the exact bytes that were replaced. A stale or absent entry is left alone.
func (s *Store) RefreshPublished(repoRoot, path string, oldData, newData []byte, invalidated Surface) (bool, error) {
	if s == nil || ContentHash(oldData) == ContentHash(newData) {
		return false, nil
	}
	relative, err := repositoryRelativePath(repoRoot, path)
	if err != nil {
		return false, err
	}
	normalized := NormalizePath(relative)
	oldHash := ContentHash(oldData)
	newHash := ContentHash(newData)

	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[normalized]
	if !ok || entry.ContentSHA256 != oldHash {
		return false, nil
	}
	entry.ContentSHA256 = newHash
	if invalidated&SurfaceFrontmatter != 0 {
		entry.FrontmatterClean = false
	}
	if invalidated&SurfaceFormat != 0 {
		entry.FormatClean = false
	}
	if !entry.FrontmatterClean && !entry.FormatClean {
		delete(s.entries, normalized)
		delete(s.dirty, normalized)
		s.deleted[normalized] = true
		return true, nil
	}
	entry = cloneEntry(entry)
	s.entries[normalized] = entry
	s.dirty[normalized] = entry
	delete(s.deleted, normalized)
	return true, nil
}

func repositoryRelativePath(repoRoot, path string) (string, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve validation-cache repository root: %w", err)
	}
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(root, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve generated rewrite path %s: %w", path, err)
	}
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", fmt.Errorf("relativize generated rewrite path %s: %w", path, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("generated rewrite is outside validation-cache repository: %s", path)
	}
	return filepath.ToSlash(relative), nil
}
