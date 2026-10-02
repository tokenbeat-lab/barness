package ai

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// Observer receives a Client's records of its logical calls and their
// attempts, for operators to trace status, retries, cancellation and usage
// per tenant (spec I9). It is not on the execution path: records are queued
// and delivered later, one at a time and in the order they were made, by a
// goroutine the Client starts while records wait. A slow, failing or
// panicking Observer never blocks or changes a call: once
// ResourcePolicy.MaxQueuedObservations records wait, further ones are
// dropped and counted (Client.ObserverStats), and an error or panic is
// counted as a failed delivery. Request callbacks (Hooks) are the
// complementary mechanism that can fail a call.
//
// Records carry identifiers and classifications only: never a key, an
// Authorization header, request or response content, tool arguments,
// native reasoning state, or an error's message text, which may quote the
// provider.
//
// TenantID, RequestID, ActorID and JobID are unbounded; a host should not
// use them as metric labels by default. A call refused before its snapshot
// was consistent still carries the scope's TenantID but stays unresolved
// (CallMetadata.Resolved false); usage, success rates and resource
// attribution should count resolved records only (ADR-0009).
//
// Unlike error text (see Error), records are safe for shared logs.
type Observer interface {
	Observe(Observation) error
}

// ObservationKind is what an Observation records.
type ObservationKind string

const (
	// ObservationCallStarted: a logical call was received, before anything
	// was resolved. Its Call holds the scope and the requested binding only.
	ObservationCallStarted ObservationKind = "call_started"
	// ObservationAttemptStarted: an attempt obtained its admission permit
	// and is about to be sent. A call refused before that has none.
	ObservationAttemptStarted ObservationKind = "attempt_started"
	// ObservationAttemptFinished: an attempt failed to obtain the initial
	// response, or the stream it opened was closed.
	ObservationAttemptFinished ObservationKind = "attempt_finished"
	// ObservationCallFinished: the call reached its terminal.
	ObservationCallFinished ObservationKind = "call_finished"
)

// Observation is one record an Observer receives. It is the Observer's own
// copy.
type Observation struct {
	Kind ObservationKind `json:"kind"`
	// Time is the system clock's, as message timestamps are.
	Time time.Time `json:"time"`
	// Call attributes the record. At CallStarted it holds the trusted scope
	// and the requested BindingID, with Resolved false; afterwards it is
	// the call's resolved attribution, which stays unresolved, naming no
	// account, provider, API or model, for a call refused before its
	// snapshot was consistent. Attempts are only set at CallFinished.
	Call CallMetadata `json:"call"`
	// Attempt is the attempt an attempt record is about; nil otherwise. At
	// AttemptStarted only its AttemptID is set; at AttemptFinished it is the
	// attempt as recorded in CallMetadata.Attempts.
	Attempt *Attempt `json:"attempt,omitempty"`
	// Duration is, at AttemptFinished, the time from AttemptStarted until
	// the attempt failed or its stream was closed and, at CallFinished, the
	// whole call's. It is encoded in nanoseconds.
	Duration time.Duration `json:"duration,omitempty"`
	// StopReason, Error and Usage are set at CallFinished only. Error is
	// nil when the call succeeded; Usage is the message's token usage, zero
	// when the provider reported none (telling unknown from zero is issue
	// 15's).
	StopReason StopReason     `json:"stopReason,omitempty"`
	Error      *ObservedError `json:"error,omitempty"`
	Usage      Usage          `json:"usage,omitzero"`
}

// ObservedError is a failed call's classification as an Observer sees it:
// Error without its message.
type ObservedError struct {
	Code              Code          `json:"code"`
	Phase             Phase         `json:"phase"`
	HTTPStatus        int           `json:"httpStatus,omitempty"`
	ProviderRequestID string        `json:"providerRequestId,omitempty"`
	RetryAfter        time.Duration `json:"retryAfter,omitempty"`
}

// ObserverStats counts a Client's records by fate since construction.
// Records still queued or being delivered are in none of the counts.
type ObserverStats struct {
	// Delivered: Observe returned nil.
	Delivered uint64 `json:"delivered"`
	// Failed: Observe returned an error or panicked.
	Failed uint64 `json:"failed"`
	// Dropped: the queue was full, so the record was never delivered.
	Dropped uint64 `json:"dropped"`
}

// ObserverStats reports what became of the Client's records; all zero when
// it has no Observer.
func (c *Client) ObserverStats() ObserverStats {
	return c.observations.stats()
}

// observations is a Client's bounded, asynchronous record queue. A nil
// *observations, a Client without an Observer, records nothing.
//
// No goroutine outlives the records: one is started when a record arrives
// and none is draining, and it exits once the queue is empty, so a Client
// needs no Close. An Observer that never returns keeps that one goroutine.
type observations struct {
	observer Observer
	max      int

	mu       sync.Mutex
	queue    []Observation
	draining bool

	delivered, failed, dropped atomic.Uint64
}

func newObservations(o Observer, max int) *observations {
	if o == nil {
		return nil
	}
	return &observations{observer: o, max: max}
}

// publish queues o without ever blocking, or drops it when max records
// already wait. The record being delivered does not count as waiting.
func (q *observations) publish(o Observation) {
	if q == nil {
		return
	}
	q.mu.Lock()
	if len(q.queue) >= q.max {
		q.mu.Unlock()
		q.dropped.Add(1)
		return
	}
	q.queue = append(q.queue, o)
	start := !q.draining
	q.draining = true
	q.mu.Unlock()
	if start {
		go q.drain()
	}
}

func (q *observations) drain() {
	for {
		q.mu.Lock()
		if len(q.queue) == 0 {
			q.draining = false
			q.queue = nil
			q.mu.Unlock()
			return
		}
		o := q.queue[0]
		q.queue[0] = Observation{}
		q.queue = q.queue[1:]
		q.mu.Unlock()
		if q.deliver(o) {
			q.delivered.Add(1)
		} else {
			q.failed.Add(1)
		}
	}
}

// deliver hands o to the Observer. A panic is contained and counted as a
// failure: the Observer runs on the Client's goroutine, where an uncaught
// panic would end the host process rather than just the record.
func (q *observations) deliver(o Observation) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return q.observer.Observe(o) == nil
}

func (q *observations) stats() ObserverStats {
	if q == nil {
		return ObserverStats{}
	}
	return ObserverStats{Delivered: q.delivered.Load(), Failed: q.failed.Load(), Dropped: q.dropped.Load()}
}

// callStarted records a call as received, before resolution.
func (q *observations) callStarted(meta CallMetadata) {
	if q == nil {
		return
	}
	meta.Attempts = nil
	q.publish(Observation{Kind: ObservationCallStarted, Time: time.Now(), Call: meta})
}

// callFinished records a call's terminal: the attribution it resolved, how
// it ended and the usage it reported. failure is the classified error, nil
// on success.
func (q *observations) callFinished(res Result, failure error, started time.Time) {
	if q == nil {
		return
	}
	now := time.Now()
	o := Observation{Kind: ObservationCallFinished, Time: now, Call: res.Metadata, Duration: now.Sub(started),
		StopReason: res.Message.StopReason, Usage: res.Message.Usage}
	o.Call.Attempts = slices.Clone(o.Call.Attempts)
	var e *Error
	if errors.As(failure, &e) {
		o.Error = &ObservedError{Code: e.Code, Phase: e.Phase, HTTPStatus: e.HTTPStatus,
			ProviderRequestID: e.ProviderRequestID, RetryAfter: e.RetryAfter}
	}
	q.publish(o)
}

// attempts returns the recorder of one resolved call's attempts; meta is
// copied, so later changes to the call's metadata do not reach records.
func (q *observations) attempts(meta CallMetadata) attemptRecorder {
	if q == nil {
		return attemptRecorder{}
	}
	meta.Attempts = nil
	return attemptRecorder{q: q, call: meta}
}

// attemptRecorder records the attempts of one call. The zero value records
// nothing.
type attemptRecorder struct {
	q    *observations
	call CallMetadata
}

func (r attemptRecorder) started(attemptID string) time.Time {
	now := time.Now()
	if r.q != nil {
		r.q.publish(Observation{Kind: ObservationAttemptStarted, Time: now, Call: r.call, Attempt: &Attempt{AttemptID: attemptID}})
	}
	return now
}

func (r attemptRecorder) finished(a Attempt, started time.Time) {
	if r.q == nil {
		return
	}
	now := time.Now()
	r.q.publish(Observation{Kind: ObservationAttemptFinished, Time: now, Call: r.call, Attempt: &a, Duration: now.Sub(started)})
}
