package runner

import (
	"context"
	"fmt"

	"golang.org/x/sync/semaphore"
)

func NewSimple[In any](
	maxWorkers int,
	input <-chan In,
	process func(context.Context, In),
) Runner {
	runner := func(ctx context.Context) error {
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
					process(ctx, s)
				}(inputObj)
			}
		}
	}

	return runner
}
