package localassembly

import (
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// LocalPolicy is a finite resource policy for one person running an agent on
// their own machine (spec I9 "正式示例", ADR-0002). It is deployment policy,
// not a pi-compatible default and not a validated universal value: every
// number below states the load it is meant for, why it was chosen and when to
// change it. The admission values follow issue 18's "准入取值建议" for local
// use; the burst scenario E08-admission-burst-pressure (ai/e2e) measures what
// they cost, and its ratios are kept with the evidence bundle (pressure.json).
//
// Assumed load: one interactive session plus a few parallel sub-agents, long
// histories with screenshots, reasoning models whose turns can run for
// minutes, and a terminal or editor that reads events as they arrive. Memory
// is the developer's own, so limits are generous and exist to stop a
// misbehaving provider or a stuck consumer, not to share capacity.
//
// The byte, queue and concurrency values were checked by the design-load
// scenarios E08-policy-pressure-local-* (ai/e2e, opt-in with
// BARNESS_AI_PRESSURE=1, run by the release gate): all 8 permits busy, each
// call sending a 4 MiB history with three 2 MiB screenshots (a 12.6 MB body)
// and streaming a 128K-token turn or a 128 KiB tool call through Responses.
// Every turn completed; the largest frame used 13% of MaxFrameBytes, a
// turn's body 44% of MaxOutputBytes, the request 38% of MaxRequestBytes,
// and the reachable heap grew by 0.4–0.5 GiB (two runs; the test provider's own 100 MiB
// copy of the requests included), within the 1 GiB a developer machine is
// assumed to spare. A stream whose consumer stopped reading ended with
// resource_limit at MaxQueuedEvents, about 11 minutes of backlog at 100
// tokens per second. The headroom of each byte, queue and concurrency limit
// the load exercises is kept as policy-pressure.json; MaxErrorBodyBytes and
// the time limits are not pressure values and rest on the reasoning beside
// them.
func LocalPolicy() *ai.ResourcePolicy {
	return &ai.ResourcePolicy{
		// 32 MiB: a long history (a 1M-token context is roughly 4 MB of
		// text) plus a handful of screenshots. Providers refuse larger
		// bodies anyway; lower it to the smallest body limit among the
		// providers you bind (check each provider's current documentation).
		MaxRequestBytes: 32 << 20,
		// 5 MiB per image, counted as image bytes: full-resolution
		// screenshots fit; raise it only if your provider accepts larger
		// images and you send them.
		MaxImageBytes: 5 << 20,
		// 4 MiB per SSE frame. Responses repeats the whole response in its
		// terminal frame, so this must hold the largest response you
		// accept: 128K output tokens are about 0.5 MB of text, and
		// reasoning items and JSON escaping can multiply that. Raise it if
		// long turns end with resource_limit in the stream phase.
		MaxFrameBytes: 4 << 20,
		// 1 MiB of argument JSON per tool call: a tool writing a file of
		// tens of thousands of lines in one call fits. The bound is memory:
		// a call holds its text about four times (the growing buffer, the
		// final copy and its parsed form), 4 MiB, 32 MiB over all 8 permits.
		// CPU stays linear in the size: the pressure scenario streams one
		// call in 64-byte deltas at about 50 ms for 512 KiB and 0.2 s for
		// 2 MiB (tool-arguments-scaling, issue 32).
		MaxToolJSONBytes: 1 << 20,
		// 64 KiB: provider error bodies are short JSON; this only guards
		// against a proxy returning a large HTML page.
		MaxErrorBodyBytes: 64 << 10,
		// 64 MiB of streamed body per call: every delta frame carries
		// 100–300 bytes of envelope for a few tokens, so a 128K-token turn
		// streams tens of MB. Lower it if your models' outputs are short.
		MaxOutputBytes: 64 << 20,
		// Events wait here only while the consumer is busy. A terminal
		// renderer keeps up, so 65536 events / 16 MiB is minutes of
		// backlog; a consumer slower than that is stuck, and the call ends
		// with resource_limit instead of holding memory.
		MaxQueuedEvents:     1 << 16,
		MaxQueuedEventBytes: 16 << 20,
		// 8 attempts at once: one session plus parallel sub-agents. A
		// single tenant owns the process, so both limits are equal.
		MaxConcurrentPerTenant: 8,
		MaxConcurrentProcess:   8,
		// Local, single user ("本地单用户"): waiting a few seconds is better
		// than an error, and requests are few, so the queue may be as long
		// as the concurrency limit. That is deliberately more than the rule
		// of thumb (8 × 5 s ÷ 20 s typical call = 2): the burst scenario,
		// saturating all 8 permits with calls ending evenly over 20 s, admits
		// 2 of 8 waiters and refuses 6 after their full 5 s wait. A local
		// program rarely saturates, and when it does a sub-agent waiting 5 s
		// to be refused costs only time. Worst-case queued memory is
		// MaxAdmissionWaiters × MaxRequestBytes = 8 × 32 MiB. Lower the cap
		// to 2–3 if your program runs at its limit for long stretches.
		AdmissionWait:       5 * time.Second,
		MaxAdmissionWaiters: 8,
		// Observations are only queued when an Observer is set.
		MaxQueuedObservations: 1024,
		// 30 minutes per call: a high-effort reasoning turn can take more
		// than ten minutes. It also bounds a call whose provider trickles
		// bytes just often enough to defeat ReadIdleTimeout.
		CallTimeout: 30 * time.Minute,
		// 10 s to connect: a dead route or DNS failure shows quickly.
		ConnectTimeout: 10 * time.Second,
		// 2 minutes to the response headers: prefill of a very long prompt
		// can delay them; most providers answer within seconds.
		ResponseHeaderTimeout: 2 * time.Minute,
		// 5 minutes without upstream bytes: some reasoning models stream
		// nothing while they think. Lower it to fail stuck connections
		// sooner if your models send keep-alive pings.
		ReadIdleTimeout: 5 * time.Minute,
	}
}
