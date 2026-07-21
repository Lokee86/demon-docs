package codemap

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunDatasetWorkersBoundsConcurrency(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	errors := runDatasetWorkers(datasetWorkerLimit*2, func(index int) error {
		current := active.Add(1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		active.Add(-1)
		return nil
	})
	for _, err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := maximum.Load(); got > datasetWorkerLimit {
		t.Fatalf("maximum concurrency %d exceeds limit %d", got, datasetWorkerLimit)
	}
}

func TestRunDatasetWorkersReturnsErrorsByJobIndex(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	workerErrors := runDatasetWorkers(3, func(index int) error {
		switch index {
		case 0:
			time.Sleep(2 * time.Millisecond)
			return first
		case 1:
			return second
		default:
			return nil
		}
	})
	if !errors.Is(workerErrors[0], first) || !errors.Is(workerErrors[1], second) || workerErrors[2] != nil {
		t.Fatalf("unexpected ordered errors: %#v", workerErrors)
	}
}
