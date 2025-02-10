package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func testProcess(_ context.Context, in struct{}) (struct{}, error) {
	return in, nil
}

type RunnerTestSuite struct {
	suite.Suite
}

func TestRunnerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RunnerTestSuite))
}

func (s *RunnerTestSuite) Test_stopByCtx() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	defer close(input)
	runner, _ := New(2, input, testProcess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.NoError(runner(ctx))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case input <- struct{}{}:
			}
		}
	}()

	time.Sleep(time.Second)
	cancel()
	wg.Wait()
}

func (s *RunnerTestSuite) Test_stopByClosedInput() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	close(input)
	runner, _ := New(2, input, testProcess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.NoError(runner(ctx))
	}()

	wg.Wait()
}

func (s *RunnerTestSuite) Test_output() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	defer close(input)
	runner, output := New(2, input, testProcess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.NoError(runner(ctx))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case input <- struct{}{}:
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range output {
			continue
		}
	}()

	time.Sleep(time.Second)
	cancel()
	wg.Wait()
}
