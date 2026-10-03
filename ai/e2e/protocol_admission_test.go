package e2e

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

// abortedMessage is pi's text, on every scenario protocol, for a call
// canceled or timed out by its context once the stream started.
const abortedMessage = "Request was aborted"

// TestProtocolTimeouts is TestTimeouts on each scenario protocol
// (P02–P04/E08): response header, read idle, the protocol's timeoutMs, the
// call time limit and the host's deadline each end the call as
// deadline_exceeded, a header timeout is retried under the binding's retry
// policy (a connection failure without a response, as the Stainless SDKs
// retry it), and every path releases what it held. pi's Google path has no
// request timeout of its own, so there timeoutMs bounds nothing, and it never
// retries a request that got no response, so neither is a header timeout
// retried (ADR-0012).
func TestProtocolTimeouts(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolTimeouts)
}

func testProtocolTimeouts(t *testing.T, s *protocolSuite) {
	a := loadScenarioText(t, s.proto)
	call := func(aw admissionWorld, ctx context.Context, e outcomeEntry, requestID string) outcome {
		return e.invoke(ctx, aw.world, scopeFor(tenantA, requestID), a.target(), a.request(t))
	}

	t.Run("response-header", func(t *testing.T) {
		for _, e := range s.entries() {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-timeout-response-header-"+e.name))
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = short }, nil)
				enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent})
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-header-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, "exceeded the resource policy's ResponseHeaderTimeout (100ms)")
				checkTimedOutAttempt(ev, o)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	if s.timeoutOptions == nil {
		t.Run("timeout-ms-bounds-nothing", func(t *testing.T) {
			ev := run.Case(t, s.caseID("E08-timeout-ms-bounds-nothing"))
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
	} else {
		testProtocolTimeoutMs(t, s, call)
	}

	// A header timeout under a binding that retries once: retried on the
	// Stainless protocols, the call's end on pi's Google path.
	header, caseName := "header-timeout-not-retried", "header-not-retried"
	if s.headerTimeoutRetried {
		header, caseName = "header-timeout-retried", "header-retried"
	}
	t.Run(header, func(t *testing.T) {
		for _, e := range s.entries() {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-timeout-"+caseName+"-"+e.name))
				aw := retryingHeaderWorld(t, s)
				enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent}, a.reply(t))
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-"+header+"-"+e.name) })
				at := o.result.Metadata.Attempts
				if s.headerTimeoutRetried {
					checkScenarioTextSucceeded(ev, a, o)
					ev.Check("the timed-out attempt and its retry are recorded", len(at) == 2 && at[0].Code == ai.CodeDeadlineExceeded && at[0].HTTPStatus == 0 && at[1].Code == "",
						"attempts=%+v", at)
				} else {
					checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, "exceeded the resource policy's ResponseHeaderTimeout (100ms)")
					ev.Check("only the timed-out attempt", len(at) == 1 && len(aw.provider.Requests()) == 1,
						"attempts=%+v requests=%d", at, len(aw.provider.Requests()))
				}
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("read-idle", func(t *testing.T) {
		for _, e := range s.entries() {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-timeout-read-idle-"+e.name))
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
		for _, e := range s.entries() {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-timeout-call-"+e.name))
				const callTimeout = 300 * time.Millisecond
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { fitCallTimeout(p, callTimeout) }, nil)
				r, _ := s.longOutput(t, 200)
				r.AfterChunk = func(int) { time.Sleep(20 * time.Millisecond) }
				enqueue(ev, aw.world, r)
				start := time.Now()
				o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-call-timeout-"+e.name) })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream, abortedMessage)
				ev.Check("the call ended at CallTimeout", time.Since(start) >= callTimeout, "took %s", time.Since(start))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("host-deadline-earlier", func(t *testing.T) {
		for _, e := range s.entries() {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-timeout-host-deadline-"+e.name))
				aw := newAdmissionWorld(t, nil, nil)
				enqueue(ev, aw.world, heldReply(200, "text/event-stream", string(a.chunks(t)[0])))
				ctx, cancel := context.WithTimeout(ctxFor(t), short)
				defer cancel()
				o := timedCall(ev, long, func() outcome { return call(aw, ctx, e, "req-host-deadline-"+e.name) })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream, abortedMessage)
				checkAdmissionReleased(ev, aw)
			})
		}
	})
}

// testProtocolTimeoutMs: whichever of the protocol's timeoutMs and the
// policy's header timeout is earlier ends the call, and a negative timeoutMs
// is refused before anything is sent.
func testProtocolTimeoutMs(t *testing.T, s *protocolSuite, call func(admissionWorld, context.Context, outcomeEntry, string) outcome) {
	for _, c := range []struct {
		name      string
		policy    time.Duration
		timeoutMs int
		msg       string
	}{
		{"protocol-timeout-earlier", long, int(short / time.Millisecond), "Request timed out."},
		{"policy-earlier-than-protocol-timeout", short, int(long / time.Millisecond), "exceeded the resource policy's ResponseHeaderTimeout (100ms)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			for _, e := range protocolTimeoutEntries(s, c.timeoutMs) {
				t.Run(e.name, func(t *testing.T) {
					ev := run.Case(t, s.caseID("E08-timeout-"+c.name+"-"+e.name))
					aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = c.policy }, nil)
					enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent})
					o := timedCall(ev, long, func() outcome { return call(aw, ctxFor(t), e, "req-"+c.name+"-"+e.name) })
					checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, c.msg)
					checkAdmissionReleased(ev, aw)
				})
			}
		})
	}

	t.Run("invalid-protocol-timeout", func(t *testing.T) {
		ev := run.Case(t, s.caseID("E08-timeout-invalid-protocol-timeout"))
		aw := newAdmissionWorld(t, nil, nil)
		for _, e := range protocolTimeoutEntries(s, -1) {
			o := within(ev, "call ends", func() outcome { return call(aw, ctxFor(t), e, "req-negative-"+e.name) })
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
		}
		ev.Check("nothing reached the provider", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
	})
}

// retryingHeaderWorld is an admission world whose header timeout is short
// and whose binding retries once, on a fake clock so the backoff is
// instant.
func retryingHeaderWorld(t *testing.T, s *protocolSuite) admissionWorld {
	aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = short }, nil, func(c *ai.Config) {
		c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0)
	})
	aw.updateBindingOf(tenantA, s.proto.binding, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
	return aw
}

// protocolTimeoutEntries are the four entry points with the protocol's timeoutMs.
func protocolTimeoutEntries(s *protocolSuite, ms int) []outcomeEntry {
	full := s.timeoutOptions(ms)
	simple := ai.SimpleOptions{TimeoutMs: ms}
	return []outcomeEntry{
		{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.Stream(ctx, scope, target, req, full))
		}},
		{"stream-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.StreamSimple(ctx, scope, target, req, simple))
		}},
		{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.Complete(ctx, scope, target, req, full)
			return outcome{result: res, err: err}
		}},
		{"complete-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.CompleteSimple(ctx, scope, target, req, simple)
			return outcome{result: res, err: err}
		}},
	}
}

// TestProtocolAdmission is TestAdmission on each scenario protocol
// (P02–P04/E08, issue 13): a streaming attempt holds its permit until it
// ends, a waiting call is admitted when it is returned, the host's admission
// sees every attempt with the tenant and the protocol's account, preflight
// failures never ask it, and Close and cancel release everything.
func TestProtocolAdmission(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolAdmission)
}

func testProtocolAdmission(t *testing.T, s *protocolSuite) {
	a := loadScenarioText(t, s.proto)
	entries := s.entries()
	call := func(aw admissionWorld, ctx context.Context, e outcomeEntry, k tenantKey, requestID string) outcome {
		return e.invoke(ctx, aw.world, scopeFor(k, requestID), a.target(), a.request(t))
	}

	t.Run("tenant-limit", func(t *testing.T) {
		ev := run.Case(t, s.caseID("E08-admission-tenant-limit"))
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
			p.MaxConcurrentPerTenant, p.MaxConcurrentProcess, p.AdmissionWait = 1, 4, 0
		}, nil)
		held := holdScenario(t, ev, aw, a, ctxFor(t), tenantA, "req-hold-a")
		ev.Check("the streaming attempt holds its permit", aw.probe.Permits() == 1, "got %d", aw.probe.Permits())
		for _, e := range entries {
			o := within(ev, "tenant A's call ends", func() outcome { return call(aw, ctxFor(t), e, tenantA, "req-a-over-"+e.name) })
			o.record(ev)
			checkAdmissionDenied(ev, o, "MaxConcurrentPerTenant")
		}
		aw.provider.Enqueue(a.reply(t))
		o := within(ev, "tenant B's call ends", func() outcome { return call(aw, ctxFor(t), entries[2], tenantB, "req-b") })
		checkScenarioTextSucceeded(ev, a, o)
		_ = held.Close()
		checkAdmissionReleased(ev, aw)
	})

	t.Run("wait-then-admitted", func(t *testing.T) {
		for _, e := range entries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-admission-wait-admitted-"+e.name))
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
		for _, e := range entries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E08-admission-host-per-attempt-"+e.name))
				admission := host.NewAdmission()
				aw := newAdmissionWorld(t, nil, admission, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) })
				aw.updateBindingOf(tenantA, s.proto.binding, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
				aw.provider.Enqueue(s.overloaded, a.reply(t))
				id := "req-host-" + e.name
				o := within(ev, "call ends", func() outcome { return call(aw, ctxFor(t), e, tenantA, id) })
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				got := admission.Requests()
				ev.Record("admission_requests", got)
				want := []ai.AdmissionRequest{
					{TenantID: tenantA.tenant, AccountScopeID: s.proto.account, AttemptID: id + "#1"},
					{TenantID: tenantA.tenant, AccountScopeID: s.proto.account, AttemptID: id + "#2"},
				}
				ev.Check("the host's admission sees each attempt with tenant and the protocol's account", len(got) == 2 && got[0] == want[0] && got[1] == want[1],
					"got %+v want %+v", got, want)
				ev.Check("an attempt's permit is returned before the next attempt", admission.PeakHeld() == 1, "peak %d", admission.PeakHeld())
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("preflight-failure-skips-admission", func(t *testing.T) {
		ev := run.Case(t, s.caseID("E08-admission-preflight-skips"))
		admission := host.NewAdmission()
		aw := newAdmissionWorld(t, nil, admission)
		aw.host.DeleteCredential(tenantB.tenant, s.credential(tenantB).CredentialID)
		for _, e := range entries {
			for _, c := range []struct {
				name   string
				scope  ai.CallScope
				target ai.Target
				code   ai.Code
				phase  ai.Phase
			}{
				{"no-request-id", ai.CallScope{TenantID: tenantA.tenant}, a.target(), ai.CodeInvalidRequest, ai.PhaseScope},
				{"model-not-allowed", scopeFor(tenantA, "req-model"), ai.Target{BindingID: s.proto.binding, ModelID: s.deniedModel}, ai.CodeTenantDenied, ai.PhaseCapability},
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
		ev := run.Case(t, s.caseID("E08-admission-release"))
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
				results <- call(aw, ctxFor(t), entries[i], k, "req-waiting-"+entries[i].name)
			}()
		}
		eventually(ev, "every extra call waits", func() bool { return aw.probe.AdmissionWaiters() == waiting })
		var wg sync.WaitGroup
		for _, st := range closed {
			for range 3 {
				wg.Go(func() { _ = st.Close() })
			}
		}
		cancelB()
		waitFor(ev, "concurrent Close calls return", doneWhen(wg.Wait))
		for _, st := range append(closed, canceled) {
			_ = st.Close()
			r := resultWithin(ev, "held call ends", st)
			checkInterrupted(ev, outcome{result: r.res, err: r.err}, ai.StopReasonAborted, ai.CodeCanceled, ai.PhaseStream, abortedMessage)
		}
		for range waiting {
			checkScenarioTextSucceeded(ev, a, waitValue(ev, "waiting call ends", results))
		}
		checkAdmissionReleased(ev, aw)
	})
}
