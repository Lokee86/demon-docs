package validationworkers

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRunProcessesEachIndexOnceWithBoundedConcurrency(t *testing.T) {
	const count = 64
	seen := make([]atomic.Int32, count)
	var active atomic.Int32
	var maximum atomic.Int32

	errors := Run(count, func(index int) error {
		current := active.Add(1)
		for {
			previous := maximum.Load()
			if current <= previous || maximum.CompareAndSwap(previous, current) {
				break
			}
		}
		seen[index].Add(1)
		time.Sleep(time.Millisecond)
		active.Add(-1)
		return nil
	})

	for index, err := range errors {
		if err != nil {
			t.Fatalf("work %d returned error: %v", index, err)
		}
		if got := seen[index].Load(); got != 1 {
			t.Fatalf("work %d ran %d times, want 1", index, got)
		}
	}
	if got := maximum.Load(); got > workerLimit {
		t.Fatalf("maximum concurrency %d exceeds limit %d", got, workerLimit)
	}
	if got := maximum.Load(); got < 2 {
		t.Fatalf("maximum concurrency %d did not demonstrate parallel work", got)
	}
}
