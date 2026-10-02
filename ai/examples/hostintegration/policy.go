package hostintegration

import (
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// CloudInteractivePolicy is a finite resource policy for a cloud instance
// serving interactive turns of many tenants (spec I9 "正式示例", ADR-0002). It
// is deployment policy, not a pi-compatible default and not a validated
// universal value: every number states the load it is meant for, why it was
// chosen and when to change it.
//
// Assumed load: one instance with about 2 GiB for barness-ai's buffers,
// clients that stream turns to a browser, a typical (p50) streamed turn of
// about 20 s, and per-account or cross-instance quotas enforced by the host's
// injected ai.Admission. A host that gets CodeAdmissionDenied answers its
// client with 429 and Retry-After: queueing belongs to the client or a task
// queue, and the built-in wait only absorbs short jitter.
//
// The admission values follow issue 18's "准入取值建议" and were checked by the
// burst scenario E08-admission-burst-pressure (ai/e2e), run at 1/20 of real
// time: MaxConcurrentProcess 32 saturated by calls ending evenly over the
// 20 s typical duration, then a burst of 32 + 2 × MaxAdmissionWaiters calls
// at once. In the run recorded with this change, the cap of 4 (a burst of
// 40) had 3 of its 4 waiters admitted and 1 refused after its full wait; a
// cap of 16 (a burst of 64) still admitted only 3 and left 13 waiting the
// full AdmissionWait for nothing. Timing varies a little between runs (under
// the race detector a fifth caller sometimes joins the queue after the first
// release); the scenario asserts the shares, not exact counts. Admission comes after a call
// built its request body (ADR-0008), so a burst briefly allocates every
// arriving call's body, refused or not. The measured ratios and peak heap
// are kept with the evidence bundle (pressure.json); re-run the scenario
// with your own concurrency, wait and typical duration before changing
// these values.
func CloudInteractivePolicy() *ai.ResourcePolicy {
	return &ai.ResourcePolicy{
		// 8 MiB per request body: long agent histories with a few images.
		// Bodies are held while attempts run, so the worst case is
		// (MaxConcurrentProcess + MaxAdmissionWaiters) × 8 MiB ≈ 288 MiB.
		// Lower it before raising concurrency on a small instance.
		MaxRequestBytes: 8 << 20,
		// 4 MiB per image (image bytes, not base64): screenshots and photos
		// a browser client uploads after its own downscaling.
		MaxImageBytes: 4 << 20,
		// 2 MiB per SSE frame: Responses repeats the whole response in its
		// terminal frame, which must fit; interactive turns stay well under
		// 64K output tokens (about 256 KB of text). Raise it with your
		// models' output limits.
		MaxFrameBytes: 2 << 20,
		// 512 KiB of argument JSON per tool call: more than any interactive
		// tool call needs; a larger one is a runaway generation.
		MaxToolJSONBytes: 512 << 10,
		// 32 KiB: provider error bodies are short JSON.
		MaxErrorBodyBytes: 32 << 10,
		// 32 MiB of streamed body per call: tens of thousands of delta
		// frames of 100–300 bytes each. 32 calls at the limit are 1 GiB,
		// the bulk of the assumed budget.
		MaxOutputBytes: 32 << 20,
		// 16384 events / 4 MiB of unread text per stream: a browser on a
		// slow link falls minutes behind before the turn ends with
		// resource_limit; the host's own send deadline normally ends it
		// earlier. A downstream that stops reading must not hold memory.
		MaxQueuedEvents:     1 << 14,
		MaxQueuedEventBytes: 4 << 20,
		// 4 attempts per tenant, 32 per instance: no tenant can take more
		// than an eighth of the instance. Size MaxConcurrentProcess from
		// the memory budget above and the provider accounts' own limits.
		MaxConcurrentPerTenant: 4,
		MaxConcurrentProcess:   32,
		// Cloud interactive ("云端交互式"): a short wait absorbs jitter, a
		// refusal reaches the client fast. Waiters only help when a permit
		// frees within AdmissionWait, so the cap is
		// MaxConcurrentProcess × AdmissionWait ÷ p50 duration
		// = 32 × 2 s ÷ 20 s ≈ 3.2, rounded up to 4. Worst-case queued
		// memory is 4 × MaxRequestBytes = 32 MiB.
		AdmissionWait:       2 * time.Second,
		MaxAdmissionWaiters: 4,
		// Observer records queued for delivery: a few per call; 4096 rides
		// out a slow exporter for seconds before records are dropped and
		// counted (Client.ObserverStats).
		MaxQueuedObservations: 4096,
		// 10 minutes per call: interactive reasoning turns end well within
		// it; a call still running is no longer interactive.
		CallTimeout: 10 * time.Minute,
		// 5 s to connect and 60 s to the response headers: the provider is
		// near; a slow start means an overloaded upstream.
		ConnectTimeout:        5 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		// 2 minutes without upstream bytes ends a stuck stream; raise it for
		// reasoning models that stream nothing while they think.
		ReadIdleTimeout: 2 * time.Minute,
	}
}

// CloudBatchPolicy is CloudInteractivePolicy for an instance running
// background jobs from a task queue ("云端批处理/后台任务"): turns may be long,
// and nobody waits for them interactively.
func CloudBatchPolicy() *ai.ResourcePolicy {
	p := CloudInteractivePolicy()
	// The task queue schedules and retries: refuse at once at capacity
	// (CodeAdmissionDenied) and let the job be retried later.
	p.AdmissionWait = 0
	// Unused when AdmissionWait is 0, but every capacity must be positive.
	p.MaxAdmissionWaiters = 1
	// Jobs run under the queue's own concurrency; per tenant, let one
	// tenant's backlog use more of an instance dedicated to jobs.
	p.MaxConcurrentPerTenant = 16
	// Long generations are expected; nobody reads events interactively.
	p.CallTimeout = 30 * time.Minute
	p.ReadIdleTimeout = 5 * time.Minute
	return p
}
