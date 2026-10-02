package e2e

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// pressureScale runs the burst at 1/20 of real time: a 20 s typical call
// lasts 1 s here, a 2 s AdmissionWait 100 ms. The admission rule of thumb
// (MaxAdmissionWaiters ≈ MaxConcurrentProcess × AdmissionWait ÷ p50) is
// scale-invariant, so the ratios carry over.
const pressureScale = 20

// pressureTypical is the typical (p50) duration of a streamed turn the
// example policies assume, before scaling.
const pressureTypical = 20 * time.Second

// TestAdmissionBurstPressure is the burst scenario of issue 18's "准入取值建议"
// (spec I9 "正式示例", ADR-0002, ADR-0008): it checks the examples' admission
// values against data rather than asserting them.
//
// The process runs at its concurrency limit with calls that end evenly
// spread over a typical duration, as in steady state. A burst of more calls
// than the limit plus the waiter cap then arrives at once. For each policy
// it records how many burst calls were refused at once (waiter cap
// reached), admitted after waiting, or refused after waiting the full
// AdmissionWait, and the peak heap while they waited. The examples' values
// are justified by these ratios; the evidence keeps them as pressure.json.
//
// Unlike the other scenarios, this one runs on real time on purpose: what
// it measures is whether a permit frees within AdmissionWait, i.e. the
// built-in limiter's own timer against call durations. Nothing orders
// events by sleeping. Its assertions are shares with wide margins (the
// rule of thumb predicts 1 of 4 waiters refused vs 13 of 16), not counts.
func TestAdmissionBurstPressure(t *testing.T) {
	interactive := hostintegration.CloudInteractivePolicy()
	oversized := hostintegration.CloudInteractivePolicy()
	oversized.MaxAdmissionWaiters *= 4
	configs := []struct {
		name   string
		policy *ai.ResourcePolicy
		// check bounds the share of waiting calls refused after the full
		// wait: a cap sized by the rule keeps it low; an oversized cap makes
		// it high.
		check func(ev *evidence.Case, r pressureReport)
	}{
		{"cloud-interactive", interactive, func(ev *evidence.Case, r pressureReport) {
			ev.Check("most waiters are admitted", r.WaitedTimedOutRatio <= 0.5, "timed out after waiting: %.2f", r.WaitedTimedOutRatio)
		}},
		{"cloud-interactive-waiters-x4", oversized, func(ev *evidence.Case, r pressureReport) {
			ev.Check("most waiters wait in vain", r.WaitedTimedOutRatio >= 0.5, "timed out after waiting: %.2f", r.WaitedTimedOutRatio)
		}},
		{"cloud-batch", hostintegration.CloudBatchPolicy(), func(ev *evidence.Case, r pressureReport) {
			// AdmissionWait 0: the task queue retries, nobody waits here. A
			// call arriving just after an in-flight call ended takes the
			// freed permit at once (counted as admitted), so not every call
			// is refused, but none is refused after waiting.
			ev.Check("no call waits to be refused", r.RefusedAfterWait == 0, "%+v", r)
			ev.Check("nearly every burst call is refused at once", r.RefusedAtOnceRatio >= 0.9, "%+v", r)
		}},
		{"local", localassembly.LocalPolicy(), func(ev *evidence.Case, r pressureReport) {
			// LocalPolicy queues as many as it runs on purpose (see its
			// doc); a saturating burst shows what that costs.
			ev.Check("some waiters are admitted", r.AdmittedAfterWait > 0, "none")
		}},
	}
	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "E08-admission-burst-pressure-"+c.name)
			r := runBurst(t, ev, *c.policy)
			ev.Record("pressure", r)
			t.Logf("%s: %+v", c.name, r)
			ev.Check("the burst exceeded concurrency plus waiters", r.Burst > r.Waiters, "burst %d waiters %d", r.Burst, r.Waiters)
			ev.Check("every burst call is accounted for", r.RefusedAtOnce+r.AdmittedAfterWait+r.RefusedAfterWait == r.Burst, "%+v", r)
			ev.Check("the waiter cap held", r.RefusedAtOnce >= r.Burst-r.Waiters-r.ReleasesDuringWait, "%+v", r)
			c.check(ev, r)
		})
	}
}

// pressureReport is one burst's outcome, in real (unscaled) terms where it
// names durations.
type pressureReport struct {
	Concurrency int    `json:"maxConcurrentProcess"`
	Waiters     int    `json:"maxAdmissionWaiters"`
	Wait        string `json:"admissionWait"`
	Typical     string `json:"typicalCall"`
	// RuleOfThumb is MaxConcurrentProcess × AdmissionWait ÷ typical call.
	RuleOfThumb float64 `json:"ruleOfThumb"`
	Scale       int     `json:"timeScale"`
	Burst       int     `json:"burst"`
	// ReleasesDuringWait is how many in-flight calls the schedule ends
	// within AdmissionWait of the burst.
	ReleasesDuringWait  int     `json:"releasesDuringWait"`
	RefusedAtOnce       int     `json:"refusedAtOnce"`
	AdmittedAfterWait   int     `json:"admittedAfterWait"`
	RefusedAfterWait    int     `json:"refusedAfterWait"`
	RefusedAtOnceRatio  float64 `json:"refusedAtOnceRatio"`
	AdmittedRatio       float64 `json:"admittedAfterWaitRatio"`
	RefusedAfterRatio   float64 `json:"refusedAfterWaitRatio"`
	WaitedTimedOutRatio float64 `json:"waitedTimedOutRatio"`
	// RequestBytes is each burst call's request body; QueuedBytesBound is
	// MaxAdmissionWaiters × MaxRequestBytes, the policy's worst case.
	RequestBytes     int    `json:"requestBytes"`
	QueuedBytesBound int64  `json:"queuedBytesBound"`
	HeapBaseline     uint64 `json:"heapInuseBaseline"`
	HeapPeak         uint64 `json:"heapInusePeak"`
}

// runBurst saturates the process with calls ending evenly over the typical
// duration, fires the burst and classifies every burst call.
func runBurst(t *testing.T, ev *evidence.Case, policy ai.ResourcePolicy) pressureReport {
	t.Helper()
	wait := policy.AdmissionWait / pressureScale
	typical := pressureTypical / pressureScale
	c := policy.MaxConcurrentProcess
	aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
		*p = policy
		// One tenant saturates the process here; the per-tenant limit is
		// E08-admission-tenant-limit's subject.
		p.MaxConcurrentPerTenant = p.MaxConcurrentProcess
		p.AdmissionWait = wait
	}, nil)
	chunks := lfChunks(t, aw.text.Events)

	// In flight: c calls, each holding its permit after its first frame
	// until its scheduled end.
	ends := make([]*cue, c)
	for i := range ends {
		ends[i] = newCue()
		entered := newCue()
		r := provider.SSE(chunks)
		r.AfterChunk = func(n int) {
			if n == 0 {
				entered.fire()
				ends[i].await(ev, "in-flight call ends on schedule")
			}
		}
		aw.provider.Enqueue(r)
		s := aw.client.Stream(ctxFor(t), scopeFor(tenantA, fmt.Sprintf("req-inflight-%d", i)), textTarget(aw.text), textRequest(aw.text), nil)
		t.Cleanup(func() { _ = s.Close() })
		entered.await(ev, "in-flight call holds its permit")
	}
	ev.Check("the process runs at its limit", aw.probe.Permits() == int64(c), "permits %d", aw.probe.Permits())

	// An admitted burst call is a typical call too: it holds its permit
	// for the typical duration, so it frees nothing for the waiters behind
	// it during the burst.
	burst := c + 2*policy.MaxAdmissionWaiters
	for range burst {
		r := provider.SSE(chunks)
		r.AfterChunk = func(n int) {
			if n == 0 {
				time.Sleep(typical)
			}
		}
		aw.provider.Enqueue(r)
	}
	prompt := strings.Repeat("p", 256<<10)
	req := ai.Request{SystemPrompt: aw.text.SystemPrompt, Messages: []ai.Message{ai.UserText(prompt)}}

	var baseline runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&baseline)
	stopSampling := make(chan struct{})
	peak := make(chan uint64, 1)
	go func() {
		top := baseline.HeapInuse
		tick := time.NewTicker(5 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				top = max(top, m.HeapInuse)
			case <-stopSampling:
				peak <- top
				return
			}
		}
	}()

	// Steady state: the in-flight calls end evenly over the typical
	// duration, the first half an interval after the burst.
	start := time.Now()
	interval := typical / time.Duration(c)
	released := 0
	for i, end := range ends {
		at := interval/2 + time.Duration(i)*interval
		if at < wait {
			released++
		}
		timer := time.AfterFunc(at, end.fire)
		t.Cleanup(func() { timer.Stop() })
	}

	ctx := ctxFor(t)
	outcomes := make([]error, burst)
	var wg sync.WaitGroup
	for i := range burst {
		wg.Go(func() {
			_, outcomes[i] = aw.client.Complete(ctx, scopeFor(tenantA, fmt.Sprintf("req-burst-%d", i)), textTarget(aw.text), req, nil)
		})
	}
	waitFor(ev, "every burst call ends", doneWhen(wg.Wait))
	close(stopSampling)
	heapPeak := <-peak
	ev.Record("burst_duration", time.Since(start).String())

	r := pressureReport{Concurrency: c, Waiters: policy.MaxAdmissionWaiters, Wait: policy.AdmissionWait.String(), Typical: pressureTypical.String(),
		RuleOfThumb: float64(c) * policy.AdmissionWait.Seconds() / pressureTypical.Seconds(), Scale: pressureScale, Burst: burst,
		ReleasesDuringWait: released, RequestBytes: len(prompt),
		QueuedBytesBound: int64(policy.MaxAdmissionWaiters) * policy.MaxRequestBytes, HeapBaseline: baseline.HeapInuse, HeapPeak: heapPeak}
	for _, err := range outcomes {
		var ae *ai.Error
		switch {
		case err == nil:
			r.AdmittedAfterWait++
		// The refusal names the limit that applied (ADR-0008): the waiter
		// cap at once, or the concurrency limit after waiting. With
		// AdmissionWait 0 the concurrency refusal is immediate too.
		case errors.As(err, &ae) && ae.Code == ai.CodeAdmissionDenied && (strings.Contains(ae.Message, "MaxAdmissionWaiters") || policy.AdmissionWait == 0):
			r.RefusedAtOnce++
		case errors.As(err, &ae) && ae.Code == ai.CodeAdmissionDenied:
			r.RefusedAfterWait++
		default:
			ev.Check("burst calls only succeed or are refused admission", false, "got %v", err)
		}
	}
	n := float64(burst)
	r.RefusedAtOnceRatio, r.AdmittedRatio, r.RefusedAfterRatio = float64(r.RefusedAtOnce)/n, float64(r.AdmittedAfterWait)/n, float64(r.RefusedAfterWait)/n
	if waited := r.AdmittedAfterWait + r.RefusedAfterWait; waited > 0 {
		r.WaitedTimedOutRatio = float64(r.RefusedAfterWait) / float64(waited)
	}
	for _, end := range ends {
		end.fire()
	}
	checkAdmissionReleased(ev, aw)
	return r
}
