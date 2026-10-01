// Package probe exposes internal resource gauges of an ai.Client to
// barness-ai's own acceptance tests (spec Testing Decisions §3 资源探针). It
// is internal, so only code inside the ai module tree can create one; hosts
// leave ai.Config.Probe nil and pay nothing for it.
package probe

import "sync/atomic"

// Probe counts one Client's queued stream events. A nil *Probe is valid and
// records nothing.
type Probe struct {
	queued atomic.Int64
	peak   atomic.Int64
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
