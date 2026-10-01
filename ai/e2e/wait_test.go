package e2e

import (
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// waitDeadline bounds every wait in these scenarios.
const waitDeadline = 5 * time.Second

// within runs fn and fails the check if it does not return within the deadline.
func within[T any](ev *evidence.Case, what string, fn func() T) T {
	ch := make(chan T, 1)
	go func() { ch <- fn() }()
	return waitValue(ev, what, ch)
}

func waitValue[T any](ev *evidence.Case, what string, ch <-chan T) T {
	select {
	case v := <-ch:
		ev.Check(what, true, "")
		return v
	case <-time.After(waitDeadline):
		ev.Check(what, false, "no result within %s", waitDeadline)
		ev.T().FailNow()
		var zero T
		return zero
	}
}

// callResult is what Result or Complete returned.
type callResult struct {
	res ai.Result
	err error
}

// resultWithin awaits s.Result within the deadline.
func resultWithin(ev *evidence.Case, what string, s *ai.Stream) callResult {
	return within(ev, what, func() callResult {
		res, err := s.Result()
		return callResult{res, err}
	})
}

func waitFor(ev *evidence.Case, what string, ch <-chan struct{}) {
	waitValue(ev, what, ch)
}

func doneWhen(wait func()) <-chan struct{} {
	ch := make(chan struct{})
	go func() { wait(); close(ch) }()
	return ch
}
