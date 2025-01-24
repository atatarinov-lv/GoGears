package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func testProcess(in struct{}) (struct{}, error) {
	return in, nil
}

type RunnerTestSuite struct {
	suite.Suite
}

func TestExampleTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RunnerTestSuite))
}

func (suite *RunnerTestSuite) TestStoppingAtInput() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	defer close(input)
	runner, output := New(2, input, testProcess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		suite.NoError(runner(ctx))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range output { //nolint:revive
		}
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

func (suite *RunnerTestSuite) TestStoppingAtOutput() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	defer close(input)
	runner, _ := New(2, input, testProcess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		suite.NoError(runner(ctx))
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

func (suite *RunnerTestSuite) TestPerformExists() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	defer close(input)
	runner, _ := New(2, input, testProcess)

	wg.Add(1)
	go func() {
		defer wg.Done()
		suite.NoError(runner.Perform(ctx))
	}()

	cancel()
	wg.Wait()
}
