package validationworkers

import "sync"

const workerLimit = 16

// Run applies independent validation work with bounded concurrency. Each index
// is assigned exactly once, so callers can store results by index and merge
// them deterministically after all workers finish.
func Run(count int, work func(index int) error) []error {
	if count == 0 {
		return nil
	}
	workerCount := count
	if workerCount > workerLimit {
		workerCount = workerLimit
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
