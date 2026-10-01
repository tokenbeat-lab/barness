// Package clock is the replaceable time and jitter source of barness-ai's
// retry backoff (spec I8: retries are repeatable and never need a real sleep
// in tests). It is internal, so only code inside the ai module tree can create
// one; hosts leave ai.Config.Clock nil and get the system clock.
//
// It is deliberately the smallest control point: what the retry rules read
// (the current time for a retry-after date, a jitter draw) and the backoff
// wait itself. Message timestamps and protocol timeouts do not use it.
package clock

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

// Clock is either the system clock (the nil *Clock) or a fake one.
type Clock struct {
	mu     sync.Mutex
	now    time.Time
	jitter float64
	sleeps []time.Duration
	// held makes fake sleeps wait for their context; each started sleep is
	// reported on it.
	held chan time.Duration
}

// NewFake returns a clock that starts at now, always draws jitter, and whose
// sleeps return at once, advancing its time by the requested duration.
func NewFake(now time.Time, jitter float64) *Clock {
	return &Clock{now: now, jitter: jitter}
}

// Now is the current time.
func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Jitter draws a value in [0, 1), as JavaScript's Math.random.
func (c *Clock) Jitter() float64 {
	if c == nil {
		return rand.Float64()
	}
	return c.jitter
}

// Sleep waits d, or until ctx ends, and returns ctx's error in that case. A
// non-positive d returns at once unless ctx has already ended.
func (c *Clock) Sleep(ctx context.Context, d time.Duration) error {
	if c == nil {
		if d <= 0 {
			return ctx.Err()
		}
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-t.C:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.mu.Lock()
	c.sleeps = append(c.sleeps, d)
	held := c.held
	if held == nil && d > 0 {
		c.now = c.now.Add(d)
	}
	c.mu.Unlock()
	if held != nil {
		held <- d
		<-ctx.Done()
	}
	return ctx.Err()
}

// Sleeps are the durations every Sleep so far was asked to wait, in order.
func (c *Clock) Sleeps() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]time.Duration(nil), c.sleeps...)
}

// HoldSleeps makes every later Sleep wait until its context ends instead of
// returning at once. Each such Sleep sends its duration on the returned
// channel when it starts, so a test can interrupt a backoff wait it knows is
// in progress.
func (c *Clock) HoldSleeps() <-chan time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Buffered beyond the few sleeps one test interrupts, so a started sleep
	// never blocks on a test that has not read it yet.
	c.held = make(chan time.Duration, 16)
	return c.held
}
