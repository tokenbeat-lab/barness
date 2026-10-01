package ai

import (
	"context"
	"sync"

	"github.com/tokenbeat-lab/barness/ai/internal/probe"
)

// Stream is one call's event stream. It is returned before any resolution or
// I/O happens; a background producer does that work and publishes events.
//
// Next and Event are for a single consumer. Result may be awaited
// independently of, or in parallel with, event consumption and never requires
// Next. Close may be called concurrently and repeatedly.
//
// Events wait in a queue bounded by the Client's ResourcePolicy
// (MaxQueuedEvents, MaxQueuedEventBytes); the producer never waits for the
// consumer. A consumer that falls behind by more than the queue holds ends
// the call: it fails with CodeResourceLimit in PhaseEventQueue, the events
// already queued stay readable in order, and the terminal ErrorEvent, whose
// place is reserved, follows them. Awaiting only Result is valid, but every
// event is still queued, so a long output ends this way. Callers that only
// need the final message must use Complete, which queues nothing.
type Stream struct {
	cancel context.CancelFunc
	probe  *probe.Probe
	done   chan struct{} // closed once result and err are final
	wake   chan struct{} // capacity 1; signals the consumer

	maxEvents int
	maxBytes  int64

	mu       sync.Mutex
	queue    []Event
	bytes    int64 // eventBytes of the queued non-terminal events
	finished bool  // the producer published its terminal event
	drained  bool  // Next returned false
	result   Result
	err      error

	cur Event // consumer-owned
}

// Stream starts one generation turn with full protocol options.
func (c *Client) Stream(ctx context.Context, scope CallScope, target Target, req Request, opts Options) *Stream {
	return c.startStream(ctx, newCall(scope, target, req, opts, nil, Hooks{}))
}

// StreamSimple is Stream with protocol-neutral options mapped per model.
func (c *Client) StreamSimple(ctx context.Context, scope CallScope, target Target, req Request, opts SimpleOptions) *Stream {
	return c.startStream(ctx, newCall(scope, target, req, nil, &opts, Hooks{}))
}

func (c *Client) startStream(ctx context.Context, cl call) *Stream {
	ctx, cancel := context.WithCancel(ctx)
	s := &Stream{cancel: cancel, probe: c.probe, done: make(chan struct{}), wake: make(chan struct{}, 1),
		maxEvents: c.policy.MaxQueuedEvents, maxBytes: c.policy.MaxQueuedEventBytes}
	go func() {
		defer cancel()
		res, err := c.run(ctx, cl, s.push)
		s.mu.Lock()
		s.result, s.err, s.finished = res, err, true
		s.mu.Unlock()
		close(s.done)
		s.signal()
	}()
	return s
}

// push queues e without ever blocking. A non-terminal event that would take
// the queue past its bounds is not queued and the overflow is returned; the
// terminal is always queued.
func (s *Stream) push(e Event) *Error {
	size := eventBytes(e)
	s.mu.Lock()
	if !isTerminal(e) {
		switch {
		case len(s.queue) >= s.maxEvents:
			s.mu.Unlock()
			return limitFailure(PhaseEventQueue, "MaxQueuedEvents", int64(s.maxEvents))
		case s.bytes+size > s.maxBytes:
			s.mu.Unlock()
			return limitFailure(PhaseEventQueue, "MaxQueuedEventBytes", s.maxBytes)
		}
		s.bytes += size
	}
	s.queue = append(s.queue, e)
	s.mu.Unlock()
	s.probe.Enqueued()
	s.signal()
	return nil
}

// eventBytes is what MaxQueuedEventBytes counts of an event: the text it
// carries itself (deltas, block contents, tool call fields). The live
// PartialView is shared by all events and the terminal's place is reserved,
// so neither counts.
func eventBytes(e Event) int64 {
	switch e := e.(type) {
	case TextDeltaEvent:
		return int64(len(e.Delta))
	case TextEndEvent:
		return int64(len(e.Content))
	case ThinkingDeltaEvent:
		return int64(len(e.Delta))
	case ThinkingEndEvent:
		return int64(len(e.Content))
	case ToolCallDeltaEvent:
		return int64(len(e.Delta))
	case ToolCallEndEvent:
		c := e.ToolCall
		return int64(len(c.ID) + len(c.Name) + len(c.Arguments) + len(c.RawArguments))
	}
	return 0
}

func isTerminal(e Event) bool {
	switch e.(type) {
	case DoneEvent, ErrorEvent:
		return true
	}
	return false
}

func (s *Stream) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Next advances to the next event, blocking until one is available. It
// returns false once the terminal event has been consumed.
func (s *Stream) Next() bool {
	for {
		s.mu.Lock()
		if len(s.queue) > 0 {
			s.cur = s.queue[0]
			s.queue[0] = nil
			s.queue = s.queue[1:]
			if !isTerminal(s.cur) {
				s.bytes -= eventBytes(s.cur)
			}
			s.mu.Unlock()
			s.probe.Dequeued()
			return true
		}
		if s.finished {
			s.cur, s.drained = nil, true
			s.mu.Unlock()
			return false
		}
		s.mu.Unlock()
		<-s.wake
	}
}

// Event returns the event Next advanced to.
func (s *Stream) Event() Event { return s.cur }

// Err follows bufio.Scanner: nil until Next has returned false, then the same
// error Result returns (nil on success).
func (s *Stream) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.drained {
		return nil
	}
	return s.err
}

// Result waits for the call to finish and returns its final result. It does
// not consume events. The error is nil on success and the classified *Error
// when the message ended with StopReason error or aborted.
func (s *Stream) Result() (Result, error) {
	<-s.done
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result.clone(), s.err
}

// Close cancels the call if it is still running and waits for the producer to
// release its resources. After the call finished it changes nothing.
func (s *Stream) Close() error {
	s.cancel()
	<-s.done
	return nil
}
