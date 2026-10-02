// Package probe exposes internal resource gauges of an ai.Client to
// barness-ai's own acceptance tests (spec Testing Decisions §3 资源探针). It
// is internal, so only code inside the ai module tree can create one; hosts
// leave ai.Config.Probe nil and pay nothing for it.
package probe

import "sync/atomic"

// Probe counts one Client's queued stream events, running calls, held
// admission permits and attempts waiting for one. A nil *Probe is valid and
// records nothing.
type Probe struct {
	queued  atomic.Int64
	peak    atomic.Int64
	active  atomic.Int64
	permits atomic.Int64
	waiters atomic.Int64
}

// New returns a zeroed probe.
func New() *Probe { return &Probe{} }

// Enqueued records one event entering a stream's queue.
func (p *Probe) Enqueued() {
	if p == nil {
		return
	}
	n := p.queued.Add(1)
	for {
		peak := p.peak.Load()
		if n <= peak || p.peak.CompareAndSwap(peak, n) {
			return
		}
	}
}

// Dequeued records one event leaving a stream's queue.
func (p *Probe) Dequeued() {
	if p != nil {
		p.queued.Add(-1)
	}
}

// QueuedEvents is the number of events currently held in the Client's stream
// queues. Events of a Stream that is dropped unread stay counted.
func (p *Probe) QueuedEvents() int64 { return p.queued.Load() }

// PeakQueuedEvents is the highest QueuedEvents seen so far.
func (p *Probe) PeakQueuedEvents() int64 { return p.peak.Load() }

// CallStarted records a call's production process starting.
func (p *Probe) CallStarted() {
	if p != nil {
		p.active.Add(1)
	}
}

// CallEnded records a call's production process having released what it
// held and settled its result.
func (p *Probe) CallEnded() {
	if p != nil {
		p.active.Add(-1)
	}
}

// ActiveCalls is the number of calls whose production process is running.
func (p *Probe) ActiveCalls() int64 { return p.active.Load() }

// Admitted records a built-in admission permit being granted.
func (p *Probe) Admitted() {
	if p != nil {
		p.permits.Add(1)
	}
}

// Released records a built-in admission permit being returned.
func (p *Probe) Released() {
	if p != nil {
		p.permits.Add(-1)
	}
}

// Permits is the number of built-in admission permits currently held.
func (p *Probe) Permits() int64 { return p.permits.Load() }

// WaitStarted records an attempt starting to wait for a permit.
func (p *Probe) WaitStarted() {
	if p != nil {
		p.waiters.Add(1)
	}
}

// WaitEnded records an attempt no longer waiting for a permit, granted or
// not.
func (p *Probe) WaitEnded() {
	if p != nil {
		p.waiters.Add(-1)
	}
}

// AdmissionWaiters is the number of attempts waiting for a permit.
func (p *Probe) AdmissionWaiters() int64 { return p.waiters.Load() }
