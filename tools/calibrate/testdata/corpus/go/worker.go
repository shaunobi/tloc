package ledger

import (
	"context"
	"errors"
	"sync"
)

type Job func(context.Context) error

// RunJobs executes jobs with bounded concurrency and returns joined failures.
func RunJobs(ctx context.Context, workers int, jobs []Job) error {
	if workers < 1 {
		return errors.New("workers must be positive")
	}
	queue := make(chan Job)
	errorsByJob := make(chan error, len(jobs))
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for job := range queue {
				if err := job(ctx); err != nil {
					errorsByJob <- err
				}
			}
		}()
	}
	for _, job := range jobs {
		select {
		case queue <- job:
		case <-ctx.Done():
			close(queue)
			group.Wait()
			return ctx.Err()
		}
	}
	close(queue)
	group.Wait()
	close(errorsByJob)
	var failures []error
	for err := range errorsByJob {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}
