package runner

import (
	"context"
	"fmt"

	"golang.org/x/sync/semaphore"
)

// A Runner is a runner to start processing.
type Runner func(ctx context.Context) error

func (r Runner) Perform(ctx context.Context) error {
	return r(ctx)
}

// New creates a new runner and returns it and a channel for getting results from function _process_.
func New[In any, Out any](
	maxWorkers int,
	input <-chan In,
	process func(In) (Out, error),
) (
	Runner,
	chan Out,
) {
	output := make(chan Out)

	runner := func(ctx context.Context) error {
		defer close(output)

		sem := semaphore.NewWeighted(int64(maxWorkers))

		// wait for already started goroutines to finish
		defer func() { //nolint:contextcheck
			_ = sem.Acquire(context.Background(), int64(maxWorkers))
		}()

		for {
			select {
			case <-ctx.Done():
				return nil
			case inputObj, open := <-input:
				if !open {
					return nil
				}

				if err := sem.Acquire(context.Background(), 1); err != nil { //nolint:contextcheck
					return fmt.Errorf("failed to acquire semaphore: %w", err)
				}

				go func(s In) {
					defer sem.Release(1)

					if result, err := process(s); err == nil {
						select {
						case <-ctx.Done():
							return
						case output <- result:
						}
					}
				}(inputObj)
			}
		}
	}

	return runner, output
}
