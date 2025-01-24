package sustainer

import (
	"context"
	"errors"
	"math"
	"time"
)

const (
	timeStep          = 100 * time.Millisecond
	timeForPerforming = 90 * time.Millisecond
)

var (
	ErrInvalidEventPerSecond  = errors.New("invalid parameter eventPerSecond")
	ErrEventPerSecondFactor10 = errors.New("eventPerSecond must be a factor of 10")
)

type Sustainer[T any] struct {
	eventPerSecond int
	eventsPerStep  int

	statistics Statistics
}

func New[T any](eventPerSecond int) (*Sustainer[T], error) {
	if eventPerSecond <= 0 {
		return nil, ErrInvalidEventPerSecond
	}

	if (eventPerSecond % 10) != 0 { //nolint:mnd
		return nil, ErrEventPerSecondFactor10
	}

	eventsPerStep := int(
		math.Round(
			(float64(timeStep) / float64(time.Second)) * float64(eventPerSecond),
		),
	)

	s := Sustainer[T]{
		eventPerSecond: eventPerSecond,
		eventsPerStep:  eventsPerStep,
	}

	return &s, nil
}

func (s *Sustainer[T]) Perform(ctx context.Context, input chan T) chan T {
	output := make(chan T, s.eventsPerStep)

	go func() {
		defer close(output)

		ticker := time.NewTicker(timeStep)
		defer ticker.Stop()

		t0 := time.Now()

		var lack, sent, throttled int

		for {
			select {
			case <-ctx.Done():
				return
			case t1 := <-ticker.C:
				pCtx, cancel := context.WithTimeout(ctx, timeForPerforming)
				cL, cS, cT := s.perform(pCtx, input, output)
				cancel() //nolint:wsl

				lack += cL
				sent += cS
				throttled += cT

				if t1.Second() != t0.Second() {
					s.statistics.Lack = lack
					s.statistics.Sent = sent
					s.statistics.Throttled = throttled

					t0 = t1
					lack, sent, throttled = 0, 0, 0
				}
			}
		}
	}()

	return output
}

//nolint:nonamedreturns
func (s *Sustainer[T]) perform(ctx context.Context, input chan T, output chan T) (lack, sent, throttled int) {
	defer func() {
		lack = s.eventsPerStep - sent - throttled
	}()

	for range s.eventsPerStep {
		select {
		case <-ctx.Done():
			return

		case item, isOpen := <-input:
			if !isOpen {
				return
			}

			select {
			case <-ctx.Done():
				return
			case output <- item:
				sent++
			default:
				throttled++
			}
		}
	}

	return
}

func (s *Sustainer[T]) Statistics() Statistics {
	stats := s.statistics
	return stats
}
