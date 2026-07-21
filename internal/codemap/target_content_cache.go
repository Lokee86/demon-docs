package codemap

import (
	"path/filepath"
	"sync"
)

type targetContentResult struct {
	ready  chan struct{}
	sha256 string
	err    error
}

type targetContentCache struct {
	mu       sync.Mutex
	entries  map[string]*targetContentResult
	readFile func(string) ([]byte, error)
}

func newTargetContentCache(readFile func(string) ([]byte, error)) *targetContentCache {
	return &targetContentCache{
		entries:  make(map[string]*targetContentResult),
		readFile: readFile,
	}
}

func (cache *targetContentCache) hashFile(path string) (string, error) {
	key := filepath.Clean(path)
	cache.mu.Lock()
	if existing := cache.entries[key]; existing != nil {
		cache.mu.Unlock()
		<-existing.ready
		return existing.sha256, existing.err
	}
	result := &targetContentResult{ready: make(chan struct{})}
	cache.entries[key] = result
	cache.mu.Unlock()

	contents, err := cache.readFile(path)
	if err == nil {
		result.sha256 = digest(contents)
	}
	result.err = err
	close(result.ready)
	return result.sha256, result.err
}
