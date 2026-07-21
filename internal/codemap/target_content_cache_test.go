package codemap

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestTargetContentCacheReadsSharedTargetOnce(t *testing.T) {
	var reads atomic.Int32
	cache := newTargetContentCache(func(path string) ([]byte, error) {
		reads.Add(1)
		return []byte("shared target"), nil
	})

	const callers = 32
	results := make([]string, callers)
	var workers sync.WaitGroup
	workers.Add(callers)
	for index := 0; index < callers; index++ {
		go func(index int) {
			defer workers.Done()
			hash, err := cache.hashFile("target.go")
			if err != nil {
				t.Errorf("hash target: %v", err)
				return
			}
			results[index] = hash
		}(index)
	}
	workers.Wait()

	if got := reads.Load(); got != 1 {
		t.Fatalf("target reads = %d, want 1", got)
	}
	for index := 1; index < len(results); index++ {
		if results[index] != results[0] {
			t.Fatalf("hash %d = %q, want %q", index, results[index], results[0])
		}
	}
}
