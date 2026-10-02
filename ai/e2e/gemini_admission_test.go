package e2e

// The suites in this file parallel their Anthropic counterparts case for
// case on the shared helpers, with Gemini's data and its deliberate
// differences (ADR-0012); sharing one protocol-parametric suite is issue 30.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// geminiAborted is pi's text for a Gemini call canceled or timed out by its
// context once the stream started.
const geminiAborted = "Request was aborted"

// TestGeminiTimeouts is TestTimeouts on the Gemini Developer API (P03/E08):
// response header, read idle, the call time limit and the host's deadline
// each end the call as deadline_exceeded, and every path releases what it
// held. pi's Google path has no request timeout of its own, so timeoutMs
// bounds nothing, and it never retries a request that got no response, so
// neither is a header timeout retried (ADR-0012).
func TestGeminiTimeouts(t *testing.T) {
	a := loadScenarioText(t, geminiProtocol)
	call := func(aw admissionWorld, ctx context.Context, e outcomeEntry, requestID string) outcome {
		return e.invoke(ctx, aw.world, scopeFor(tenantA, requestID), a.target(), a.request(t))
	}

	t.Run("response-header", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-timeout-response-header-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = short }, nil)
				enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent})
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-header-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, "exceeded the resource policy's ResponseHeaderTimeout (100ms)")
				checkTimedOutAttempt(ev, o)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("timeout-ms-bounds-nothing", func(t *testing.T) {
		ev := run.Case(t, "P03-E08-timeout-ms-bounds-nothing")
		aw := newAdmissionWorld(t, nil, nil)
		r := a.reply(t)
		r.OnReceive = func() { time.Sleep(3 * short) }
		enqueue(ev, aw.world, r)
		o := timedCall(ev, long, func() outcome {
			res, err := aw.client.CompleteSimple(ctxFor(t), scopeFor(tenantA, "req-timeout-ms"), a.target(), a.request(t),
				ai.SimpleOptions{TimeoutMs: int(short / time.Millisecond)})
			return outcome{result: res, err: err}
		})
		checkScenarioTextSucceeded(ev, a, o)
		checkAdmissionReleased(ev, aw)
	})

	t.Run("header-timeout-not-retried", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-timeout-header-not-retried-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = short }, nil, func(c *ai.Config) {
					c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0)
				})
				aw.updateBindingOf(tenantA, "gemini", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
				enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent}, a.reply(t))
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-not-retried-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, "exceeded the resource policy's ResponseHeaderTimeout (100ms)")
				ev.Check("only the timed-out attempt", len(o.result.Metadata.Attempts) == 1 && len(aw.provider.Requests()) == 1,
					"attempts=%+v requests=%d", o.result.Metadata.Attempts, len(aw.provider.Requests()))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("read-idle", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-timeout-read-idle-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ReadIdleTimeout = short }, nil)
				enqueue(ev, aw.world, heldReply(200, "text/event-stream", joinChunks(a.chunks(t)[:a.firstDelta(t)+1])))
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-idle-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseStream, "exceeded the resource policy's ReadIdleTimeout (100ms)")
				ev.Check("text received before the stall is kept", textOf(o.result.Message) == "Hello", "got %q", textOf(o.result.Message))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("call-timeout", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-timeout-call-"+e.name)
				const callTimeout = 300 * time.Millisecond
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { fitCallTimeout(p, callTimeout) }, nil)
				replies, _ := geminiLongOutput(t, 200)
				r := replies[0]
				r.AfterChunk = func(int) { time.Sleep(20 * time.Millisecond) }
				enqueue(ev, aw.world, r)
				start := time.Now()
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-call-timeout-"+e.name) })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream, geminiAborted)
				ev.Check("the call ended at CallTimeout", time.Since(start) >= callTimeout, "took %s", time.Since(start))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("host-deadline-earlier", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-timeout-host-deadline-"+e.name)
				aw := newAdmissionWorld(t, nil, nil)
				enqueue(ev, aw.world, heldReply(200, "text/event-stream", string(a.chunks(t)[0])))
				ctx, cancel := context.WithTimeout(ctxFor(t), short)
				defer cancel()
				o := timedCall(ev, long, func() outcome { return call(aw, ctx, e, "req-host-deadline-"+e.name) })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream, geminiAborted)
				checkAdmissionReleased(ev, aw)
			})
		}
	})
}

// TestGeminiAdmission is TestAdmission on the Gemini Developer API (P03/E08,
// issue 13): a streaming attempt holds its permit until it ends, a waiting
// call is admitted when it is returned, the host's admission sees every
// attempt with the tenant and the Google account, preflight failures never
// ask it, and Close and cancel release everything.
func TestGeminiAdmission(t *testing.T) {
	a := loadScenarioText(t, geminiProtocol)
	call := func(aw admissionWorld, ctx context.Context, e outcomeEntry, k tenantKey, requestID string) outcome {
		return e.invoke(ctx, aw.world, scopeFor(k, requestID), a.target(), a.request(t))
	}

	t.Run("tenant-limit", func(t *testing.T) {
		ev := run.Case(t, "P03-E08-admission-tenant-limit")
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
			p.MaxConcurrentPerTenant, p.MaxConcurrentProcess, p.AdmissionWait = 1, 4, 0
		}, nil)
		held := holdScenario(t, ev, aw, a, ctxFor(t), tenantA, "req-hold-a")
		ev.Check("the streaming attempt holds its permit", aw.probe.Permits() == 1, "got %d", aw.probe.Permits())
		for _, e := range geminiEntries {
			o := within(ev, "tenant A's call ends", func() outcome { return call(aw, ctxFor(t), e, tenantA, "req-a-over-"+e.name) })
			o.record(ev)
			checkAdmissionDenied(ev, o, "MaxConcurrentPerTenant")
		}
		aw.provider.Enqueue(a.reply(t))
		o := within(ev, "tenant B's call ends", func() outcome { return call(aw, ctxFor(t), geminiEntries[2], tenantB, "req-b") })
		checkScenarioTextSucceeded(ev, a, o)
		_ = held.Close()
		checkAdmissionReleased(ev, aw)
	})

	t.Run("wait-then-admitted", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-admission-wait-admitted-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.MaxConcurrentPerTenant, p.AdmissionWait = 1, 3*time.Second }, nil)
				held := holdScenario(t, ev, aw, a, ctxFor(t), tenantA, "req-hold")
				aw.provider.Enqueue(a.reply(t))
				done := make(chan outcome, 1)
				go func() { done <- call(aw, ctxFor(t), e, tenantA, "req-waiting") }()
				eventually(ev, "the call waits for a permit", func() bool { return aw.probe.AdmissionWaiters() == 1 })
				ev.Check("a waiting call sends nothing", len(aw.provider.Requests()) == 1, "got %d", len(aw.provider.Requests()))
				_ = held.Close()
				o := waitValue(ev, "the waiting call ends", done)
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("host-admission-per-attempt", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E08-admission-host-per-attempt-"+e.name)
				admission := host.NewAdmission()
				aw := newAdmissionWorld(t, nil, admission, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) })
				aw.updateBindingOf(tenantA, "gemini", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
				overloaded := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"},
					Chunks: [][]byte{[]byte(`{"error":{"code":503,"message":"The model is overloaded.","status":"UNAVAILABLE"}}`)}}
				aw.provider.Enqueue(overloaded, a.reply(t))
				id := "req-host-" + e.name
				o := within(ev, "call ends", func() outcome { return call(aw, ctxFor(t), e, tenantA, id) })
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				got := admission.Requests()
				ev.Record("admission_requests", got)
				want := []ai.AdmissionRequest{
					{TenantID: tenantA.tenant, AccountScopeID: "acct-tenant-a-google", AttemptID: id + "#1"},
					{TenantID: tenantA.tenant, AccountScopeID: "acct-tenant-a-google", AttemptID: id + "#2"},
				}
				ev.Check("the host's admission sees each attempt with tenant and Google account", len(got) == 2 && got[0] == want[0] && got[1] == want[1],
					"got %+v want %+v", got, want)
				ev.Check("an attempt's permit is returned before the next attempt", admission.PeakHeld() == 1, "peak %d", admission.PeakHeld())
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("preflight-failure-skips-admission", func(t *testing.T) {
		ev := run.Case(t, "P03-E08-admission-preflight-skips")
		admission := host.NewAdmission()
		aw := newAdmissionWorld(t, nil, admission)
		aw.host.DeleteCredential(tenantB.tenant, "cred-tenant-b-google")
		for _, e := range geminiEntries {
			for _, c := range []struct {
				name   string
				scope  ai.CallScope
				target ai.Target
				code   ai.Code
				phase  ai.Phase
			}{
				{"no-request-id", ai.CallScope{TenantID: tenantA.tenant}, a.target(), ai.CodeInvalidRequest, ai.PhaseScope},
				{"model-not-allowed", scopeFor(tenantA, "req-model"), ai.Target{BindingID: "gemini", ModelID: "gemini-3.1-flash-live-preview"}, ai.CodeTenantDenied, ai.PhaseCapability},
				{"credential-missing", scopeFor(tenantB, "req-cred"), a.target(), ai.CodeCredentialUnavailable, ai.PhaseCredential},
			} {
				o := within(ev, c.name+" ends", func() outcome { return e.invoke(ctxFor(t), aw.world, c.scope, c.target, a.request(t)) })
				checkRejected(ev, o, c.code, c.phase)
			}
		}
		ev.Check("no preflight failure asks for admission", len(admission.Requests()) == 0, "got %+v", admission.Requests())
		ev.Check("no preflight failure sends a request", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
		checkAdmissionReleased(ev, aw)
	})

	t.Run("release-under-close-and-cancel", func(t *testing.T) {
		// Holds end by concurrent and repeated Close and by cancellation;
		// the waiters are admitted in turn and nothing is left held,
		// waiting or open.
		ev := run.Case(t, "P03-E08-admission-release")
		admission := host.NewAdmission()
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
			p.MaxConcurrentPerTenant, p.MaxConcurrentProcess, p.AdmissionWait = 2, 3, 3*time.Second
		}, admission)
		closed := []*ai.Stream{
			holdScenario(t, ev, aw, a, ctxFor(t), tenantA, "req-hold-a1"),
			holdScenario(t, ev, aw, a, ctxFor(t), tenantA, "req-hold-a2"),
		}
		ctxB, cancelB := context.WithCancel(ctxFor(t))
		canceled := holdScenario(t, ev, aw, a, ctxB, tenantB, "req-hold-b1")
		const waiting = 3
		for range waiting {
			aw.provider.Enqueue(a.reply(t))
		}
		results := make(chan outcome, waiting)
		for i := range waiting {
			k := []tenantKey{tenantA, tenantB}[i%2]
			go func() {
				results <- call(aw, ctxFor(t), geminiEntries[i], k, "req-waiting-"+geminiEntries[i].name)
			}()
		}
		eventually(ev, "every extra call waits", func() bool { return aw.probe.AdmissionWaiters() == waiting })
		var wg sync.WaitGroup
		for _, s := range closed {
			for range 3 {
				wg.Go(func() { _ = s.Close() })
			}
		}
		cancelB()
		waitFor(ev, "concurrent Close calls return", doneWhen(wg.Wait))
		for _, s := range append(closed, canceled) {
			_ = s.Close()
			r := resultWithin(ev, "held call ends", s)
			checkInterrupted(ev, outcome{result: r.res, err: r.err}, ai.StopReasonAborted, ai.CodeCanceled, ai.PhaseStream, geminiAborted)
		}
		for range waiting {
			checkScenarioTextSucceeded(ev, a, waitValue(ev, "waiting call ends", results))
		}
		checkAdmissionReleased(ev, aw)
	})
}
