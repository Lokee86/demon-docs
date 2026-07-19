package links

import "sync"

const linkUpdateWorkerLimit = 16

// runLinkWorkers applies work with a fixed upper bound on concurrent file
// operations. Each index is assigned to exactly one worker, so callers can
// store results by index and merge them deterministically after this returns.
func runLinkWorkers(count int, work func(index int) error) []error {
	if count == 0 {
		return nil
	}
	workerCount := count
	if workerCount > linkUpdateWorkerLimit {
		workerCount = linkUpdateWorkerLimit
	}

	errors := make([]error, count)
	jobs := make(chan int)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer workers.Done()
			for index := range jobs {
				errors[index] = work(index)
			}
		}()
	}
	for index := 0; index < count; index++ {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	return errors
}
