package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func testProcessSimple(_ context.Context, _ struct{}) {}

type SimpleRunnerTestSuite struct {
	suite.Suite
}

func TestSimpleRunnerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SimpleRunnerTestSuite))
}

func (s *SimpleRunnerTestSuite) Test_stopByCtx() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	defer close(input)
	runner := NewSimple(2, input, testProcessSimple)

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

func (s *SimpleRunnerTestSuite) Test_stopByClosedInput() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan struct{})
	close(input)
	runner := NewSimple(2, input, testProcessSimple)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.NoError(runner(ctx))
	}()

	wg.Wait()
}
