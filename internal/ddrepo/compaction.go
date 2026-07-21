package ddrepo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	git "github.com/go-git/go-git/v5"
)

type CompactionThresholds struct {
	LooseFileCount int
	LooseBytes     int64
}

type Options struct {
	Compaction CompactionThresholds
}

func DefaultCompactionThresholds() CompactionThresholds {
	// Automatic compaction is disabled until private-repository readers and
	// writers are coordinated across processes. The repository demon and CLI
	// run in separate processes, so an in-process mutex cannot prevent one
	// process from removing a packfile while another process is reading it.
	return CompactionThresholds{}
}

var repositoryWriteMu sync.Mutex
var compactionSlot = make(chan struct{}, 1)

var repackRepositoryObjects = func(repository *git.Repository) error {
	return repository.RepackObjects(&git.RepackConfig{})
}

var pruneRepositoryObjects = func(repository *git.Repository) error {
	return repository.Prune(git.PruneOptions{Handler: repository.DeleteObject})
}

// WithRepositoryWriteLock serializes private-repository writers and maintenance
// in this process. The callback must include its complete object publication.
func WithRepositoryWriteLock(path string, fn func() error) error {
	if _, err := ddocsPath(path); err != nil {
		return err
	}
	if fn == nil {
		return errors.New("ddrepo: repository write callback is nil")
	}
	repositoryWriteMu.Lock()
	defer repositoryWriteMu.Unlock()
	return fn()
}

// CompactIfNeeded checks loose objects after a successful publication. It
// repacks every object reachable from every reference before pruning anything.
func CompactIfNeeded(repository *git.Repository, path string, thresholds CompactionThresholds) (bool, error) {
	if repository == nil {
		return false, errors.New("ddrepo: compaction repository is nil")
	}
	if thresholds.LooseFileCount <= 0 && thresholds.LooseBytes <= 0 {
		return false, nil
	}
	storagePath, err := ddocsPath(path)
	if err != nil {
		return false, err
	}
	count, bytes, err := looseObjectStats(storagePath)
	if err != nil {
		return false, err
	}
	if !thresholds.exceeded(count, bytes) {
		return false, nil
	}

	compactionSlot <- struct{}{}
	defer func() { <-compactionSlot }()
	defer reindexAfterCompaction(repository)

	if err := repackRepositoryObjects(repository); err != nil {
		return false, fmt.Errorf("repack ddocs objects: %w", err)
	}
	if err := pruneRepositoryObjects(repository); err != nil {
		return false, fmt.Errorf("prune ddocs objects: %w", err)
	}
	return true, nil
}

func reindexAfterCompaction(repository *git.Repository) {
	if reindexer, ok := repository.Storer.(interface{ Reindex() }); ok {
		reindexer.Reindex()
	}
}

func (t CompactionThresholds) exceeded(count int, bytes int64) bool {
	return t.LooseFileCount > 0 && count > t.LooseFileCount ||
		t.LooseBytes > 0 && bytes > t.LooseBytes
}

func looseObjectStats(storagePath string) (int, int64, error) {
	objectsPath := filepath.Join(storagePath, "objects")
	entries, err := os.ReadDir(objectsPath)
	if err != nil {
		return 0, 0, err
	}
	var count int
	var bytes int64
	for _, entry := range entries {
		if !entry.IsDir() || len(entry.Name()) != 2 {
			continue
		}
		children, err := os.ReadDir(filepath.Join(objectsPath, entry.Name()))
		if err != nil {
			return 0, 0, err
		}
		for _, child := range children {
			if child.IsDir() {
				continue
			}
			info, err := child.Info()
			if err != nil {
				return 0, 0, err
			}
			count++
			bytes += info.Size()
		}
	}
	return count, bytes, nil
}

func (r *Repository) compactAfterWrite() {
	// Logical reference publication already succeeded. Maintenance is best
	// effort and must never change the result of that completed write.
	_, _ = CompactIfNeeded(r.git, r.path, r.compaction)
}
