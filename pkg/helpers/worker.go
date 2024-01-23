package helpers

import (
	"context"
	"sync"
)

type WorkerResult struct {
	result any
	err    error
}

func RunWorkers[T any, K any](ctx context.Context, numWorkers int, input []T, worker func(ctx context.Context, input T) (K, error)) ([]K, error) {
	// if input is lower that number of workers set number of workers to input length
	if len(input) < numWorkers {
		numWorkers = len(input)
	}

	buff := func() int {
		maxSize := 100
		if size := len(input); size <= maxSize {
			return size
		}
		return maxSize
	}()

	// prepare channels and wait group
	jobs := make(chan T, buff)
	resChan := make(chan WorkerResult, buff)
	wg := sync.WaitGroup{}

	// start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(fCtx context.Context, fJobs <-chan T, fResChan chan<- WorkerResult) {
			defer func() {
				wg.Done()
			}()

			for job := range fJobs {
				result, err := worker(ctx, job)
				fResChan <- WorkerResult{result: result, err: err}
			}
		}(ctx, jobs, resChan)
	}

	go func() {
		// send all jobs to workers
		for _, job := range input {
			jobs <- job
		}
		// close channel no more jobs
		close(jobs)

		// close resChan when all workers are done
		wg.Wait()
		close(resChan)
	}()

	// process results
	result := make([]K, 0, len(input))
	for o := range resChan {
		if o.err != nil {
			return nil, o.err
		}
		if o.result != nil {
			result = append(result, o.result.(K))
		}
	}

	return result, nil
}
