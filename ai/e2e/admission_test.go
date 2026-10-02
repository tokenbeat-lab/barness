package e2e

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestAdmission covers E08's admission (spec I9, User Stories 29–30, 42):
// every attempt runs under a permit of the built-in per-tenant and process
// concurrency limits and, when the host injects one, of the host's
// admission, which receives the tenant and the vendor account. At capacity
// an attempt is refused at once or after a bounded, cancelable wait; one
// tenant at its own limit never keeps another from a permit; and every path
// returns its permits and waiters.
func TestAdmission(t *testing.T) {
	t.Run("tenant-limit-isolates-tenants", func(t *testing.T) {
		ev := run.Case(t, "E08-admission-tenant-limit")
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
			p.MaxConcurrentPerTenant, p.MaxConcurrentProcess, p.AdmissionWait = 2, 4, 0
		}, nil)
		holds := []*ai.Stream{
			aw.hold(t, ev, ctxFor(t), tenantA, "req-hold-a1"),
			aw.hold(t, ev, ctxFor(t), tenantA, "req-hold-a2"),
		}
		ev.Check("tenant A holds its two permits", aw.probe.Permits() == 2, "got %d", aw.probe.Permits())
		for _, e := range outcomeEntries {
			o := within(ev, "tenant A's call ends", func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-a-over-"+e.name) })
			o.record(ev)
			checkAdmissionDenied(ev, o, "MaxConcurrentPerTenant")
		}
		ev.Check("denied calls send nothing", len(aw.provider.Requests()) == 2, "got %d requests", len(aw.provider.Requests()))
		for _, e := range outcomeEntries {
			aw.provider.Enqueue(sseReply(t, aw.text, provider.FramingLF))
			o := within(ev, "tenant B's call ends", func() outcome { return aw.call(ctxFor(t), e, tenantB, "req-b-"+e.name) })
			checkTextSucceeded(ev, aw.text, o)
		}
		reqs := aw.provider.Requests()
		ev.Check("tenant B's calls reached the provider with B's key", len(reqs) == 6 && reqs[5].KeyAlias == tenantB.alias, "got %d requests", len(reqs))
		for _, s := range holds {
			_ = s.Close()
		}
		checkAdmissionReleased(ev, aw)
	})

	t.Run("process-limit", func(t *testing.T) {
		ev := run.Case(t, "E08-admission-process-limit")
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
			p.MaxConcurrentPerTenant, p.MaxConcurrentProcess, p.AdmissionWait = 2, 3, 0
		}, nil)
		aw.hold(t, ev, ctxFor(t), tenantA, "req-hold-a1")
		aw.hold(t, ev, ctxFor(t), tenantA, "req-hold-a2")
		b := aw.hold(t, ev, ctxFor(t), tenantB, "req-hold-b1")
		for _, e := range outcomeEntries {
			o := within(ev, "tenant B's call ends", func() outcome { return aw.call(ctxFor(t), e, tenantB, "req-b-over-"+e.name) })
			o.record(ev)
			checkAdmissionDenied(ev, o, "MaxConcurrentProcess")
		}
		// A permit returned to the process serves the next call at once.
		_ = b.Close()
		aw.provider.Enqueue(sseReply(t, aw.text, provider.FramingLF))
		o := within(ev, "tenant B's next call ends", func() outcome { return aw.call(ctxFor(t), outcomeEntries[2], tenantB, "req-b-after") })
		checkTextSucceeded(ev, aw.text, o)
	})

	t.Run("wait-then-admitted", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-admission-wait-admitted-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
					p.MaxConcurrentPerTenant, p.AdmissionWait = 1, 3*time.Second
				}, nil)
				held := aw.hold(t, ev, ctxFor(t), tenantA, "req-hold")
				aw.provider.Enqueue(sseReply(t, aw.text, provider.FramingLF))
				done := make(chan outcome, 1)
				go func() { done <- aw.call(ctxFor(t), e, tenantA, "req-waiting") }()
				eventually(ev, "the call waits for a permit", func() bool { return aw.probe.AdmissionWaiters() == 1 })
				ev.Check("a waiting call sends nothing", len(aw.provider.Requests()) == 1, "got %d", len(aw.provider.Requests()))
				_ = held.Close()
				o := waitValue(ev, "the waiting call ends", done)
				o.record(ev)
				checkTextSucceeded(ev, aw.text, o)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("wait-expires", func(t *testing.T) {
		ev := run.Case(t, "E08-admission-wait-expires")
		const wait = 100 * time.Millisecond
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.MaxConcurrentPerTenant, p.AdmissionWait = 1, wait }, nil)
		aw.hold(t, ev, ctxFor(t), tenantA, "req-hold")
		for _, e := range outcomeEntries {
			start := time.Now()
			o := within(ev, "call ends", func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-expires-"+e.name) })
			waited := time.Since(start)
			o.record(ev)
			checkAdmissionDenied(ev, o, "MaxConcurrentPerTenant")
			ev.Check("the call waited for AdmissionWait first", waited >= wait, "waited %s", waited)
			ev.Check("no attempt is left waiting", aw.probe.AdmissionWaiters() == 0, "%d waiting", aw.probe.AdmissionWaiters())
		}
	})

	t.Run("wait-interrupted", func(t *testing.T) {
		for _, e := range outcomeEntries {
			for _, how := range []struct {
				name string
				code ai.Code
				msg  string
				// start runs the call and returns how to interrupt it.
				ctx func(t *testing.T) (context.Context, func())
			}{
				{"cancel", ai.CodeCanceled, "Request aborted", func(t *testing.T) (context.Context, func()) {
					return context.WithCancel(ctxFor(t))
				}},
				{"deadline", ai.CodeDeadlineExceeded, "Request timed out.", func(t *testing.T) (context.Context, func()) {
					ctx, cancel := context.WithTimeout(ctxFor(t), 150*time.Millisecond)
					t.Cleanup(cancel)
					return ctx, func() { <-ctx.Done() }
				}},
			} {
				t.Run(how.name+"/"+e.name, func(t *testing.T) {
					ev := run.Case(t, "E08-admission-wait-"+how.name+"-"+e.name)
					aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
						p.MaxConcurrentPerTenant, p.AdmissionWait = 1, 3*time.Second
					}, nil)
					held := aw.hold(t, ev, ctxFor(t), tenantA, "req-hold")
					ctx, interrupt := how.ctx(t)
					done := make(chan outcome, 1)
					go func() { done <- aw.call(ctx, e, tenantA, "req-interrupted") }()
					eventually(ev, "the call waits for a permit", func() bool { return aw.probe.AdmissionWaiters() == 1 })
					interrupt()
					interrupted := time.Now()
					o := waitValue(ev, "the waiting call ends", done)
					o.record(ev)
					ev.Check("the wait ends with the call, not at AdmissionWait", time.Since(interrupted) < time.Second, "took %s", time.Since(interrupted))
					checkInterrupted(ev, o, ai.StopReasonAborted, how.code, ai.PhaseAdmission, how.msg)
					ev.Check("no attempt is left waiting", aw.probe.AdmissionWaiters() == 0, "%d waiting", aw.probe.AdmissionWaiters())
					ev.Check("only the held call reached the provider", len(aw.provider.Requests()) == 1, "got %d", len(aw.provider.Requests()))
					_ = held.Close()
					checkAdmissionReleased(ev, aw)
				})
			}
		}
	})

	t.Run("host-admission-per-attempt", func(t *testing.T) {
		// A retried call asks the host's admission once per attempt, with the
		// tenant and the vendor account, and returns each permit before the
		// next attempt.
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-admission-host-per-attempt-"+e.name)
				admission := host.NewAdmission()
				aw := newAdmissionWorld(t, nil, admission, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) })
				aw.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
				failing := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte("{}")}, Script: "{}"}
				aw.provider.Enqueue(failing, sseReply(t, aw.text, provider.FramingLF))
				scope := "req-host-" + e.name
				o := within(ev, "call ends", func() outcome { return aw.call(ctxFor(t), e, tenantA, scope) })
				o.record(ev)
				checkTextSucceeded(ev, aw.text, o)
				got := admission.Requests()
				ev.Record("admission_requests", got)
				want := []ai.AdmissionRequest{
					{TenantID: tenantA.tenant, AccountScopeID: "acct-" + tenantA.tenant, AttemptID: scope + "#1"},
					{TenantID: tenantA.tenant, AccountScopeID: "acct-" + tenantA.tenant, AttemptID: scope + "#2"},
				}
				ev.Check("the host's admission sees each attempt with tenant and account", len(got) == 2 && got[0] == want[0] && got[1] == want[1],
					"got %+v want %+v", got, want)
				ev.Check("an attempt's permit is returned before the next attempt", admission.PeakHeld() == 1, "peak %d", admission.PeakHeld())

				o = within(ev, "tenant B's call ends", func() outcome {
					aw.provider.Enqueue(sseReply(t, aw.text, provider.FramingLF))
					return aw.call(ctxFor(t), e, tenantB, "req-host-b-"+e.name)
				})
				checkTextSucceeded(ev, aw.text, o)
				got = admission.Requests()
				ev.Check("tenant B's attempt is attributed to B's tenant and account", len(got) == 3 &&
					got[2] == ai.AdmissionRequest{TenantID: tenantB.tenant, AccountScopeID: "acct-" + tenantB.tenant, AttemptID: "req-host-b-" + e.name + "#1"},
					"got %+v", got)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("host-admission-denies", func(t *testing.T) {
		ev := run.Case(t, "E08-admission-host-denies")
		admission := host.NewAdmission()
		admission.Deny(errHostAdmission)
		aw := newAdmissionWorld(t, nil, admission)
		for _, e := range outcomeEntries {
			o := within(ev, "call ends", func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-host-denied-"+e.name) })
			o.record(ev)
			checkAdmissionDenied(ev, o, "admission")
			ev.Check("the host's error is kept for diagnosis", errors.Is(o.err, errHostAdmission), "err=%v", o.err)
			ev.Check("the host's error text is not published", !strings.Contains(o.result.Message.ErrorMessage, "acct-tenant-a"),
				"got %q", o.result.Message.ErrorMessage)
		}
		ev.Check("denied calls send nothing", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
		checkAdmissionReleased(ev, aw)
	})

	t.Run("preflight-failure-skips-admission", func(t *testing.T) {
		ev := run.Case(t, "E08-admission-preflight-skips")
		admission := host.NewAdmission()
		aw := newAdmissionWorld(t, nil, admission)
		aw.host.DeleteCredential(tenantB.tenant, "cred-"+tenantB.tenant)
		for _, e := range outcomeEntries {
			for _, c := range []struct {
				name   string
				scope  ai.CallScope
				target ai.Target
				code   ai.Code
				phase  ai.Phase
			}{
				{"no-request-id", ai.CallScope{TenantID: tenantA.tenant}, textTarget(aw.text), ai.CodeInvalidRequest, ai.PhaseScope},
				{"unknown-binding", scopeFor(tenantA, "req-unknown"), ai.Target{BindingID: "missing", ModelID: aw.text.Model}, ai.CodeBindingNotFound, ai.PhaseBinding},
				{"model-not-allowed", scopeFor(tenantA, "req-model"), ai.Target{BindingID: "primary", ModelID: "gpt-4.1"}, ai.CodeTenantDenied, ai.PhaseCapability},
				{"credential-missing", scopeFor(tenantB, "req-cred"), textTarget(aw.text), ai.CodeCredentialUnavailable, ai.PhaseCredential},
			} {
				o := within(ev, c.name+" ends", func() outcome { return e.invoke(ctxFor(t), aw.world, c.scope, c.target, textRequest(aw.text)) })
				checkRejected(ev, o, c.code, c.phase)
			}
		}
		ev.Check("no preflight failure asks for admission", len(admission.Requests()) == 0, "got %+v", admission.Requests())
		ev.Check("no preflight failure sends a request", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
		checkAdmissionReleased(ev, aw)
	})

	t.Run("release-under-close-and-cancel", func(t *testing.T) {
		// Both tenants fill the process; more calls wait. Holds then end by
		// concurrent and repeated Close, and by the host canceling the call
		// as a downstream disconnect does. Every waiter is admitted in turn
		// and completes, and nothing is left held, waiting or open.
		ev := run.Case(t, "E08-admission-release")
		admission := host.NewAdmission()
		aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) {
			p.MaxConcurrentPerTenant, p.MaxConcurrentProcess, p.AdmissionWait = 3, 4, 3*time.Second
		}, admission)
		baseline := runtime.NumGoroutine()
		closed := []*ai.Stream{
			aw.hold(t, ev, ctxFor(t), tenantA, "req-hold-a1"),
			aw.hold(t, ev, ctxFor(t), tenantA, "req-hold-a2"),
		}
		ctxA, cancelA := context.WithCancel(ctxFor(t))
		ctxB, cancelB := context.WithCancel(ctxFor(t))
		canceled := []*ai.Stream{
			aw.hold(t, ev, ctxA, tenantA, "req-hold-a3"),
			aw.hold(t, ev, ctxB, tenantB, "req-hold-b1"),
		}
		const waiting = 4
		for range waiting {
			aw.provider.Enqueue(sseReply(t, aw.text, provider.FramingLF))
		}
		results := make(chan outcome, waiting)
		for i := range waiting {
			k := []tenantKey{tenantA, tenantB}[i%2]
			go func() { results <- aw.call(ctxFor(t), outcomeEntries[i], k, "req-waiting-"+outcomeEntries[i].name) }()
		}
		eventually(ev, "every extra call waits", func() bool { return aw.probe.AdmissionWaiters() == waiting })

		var wg sync.WaitGroup
		for _, s := range closed {
			for range 3 {
				wg.Add(1)
				go func() { defer wg.Done(); _ = s.Close() }()
			}
		}
		cancelA()
		cancelB()
		waitFor(ev, "concurrent Close calls return", doneWhen(wg.Wait))
		for _, s := range append(closed, canceled...) {
			_ = s.Close() // again, after the call ended
			r := resultWithin(ev, "held call ends", s)
			checkInterrupted(ev, outcome{result: r.res, err: r.err}, ai.StopReasonAborted, ai.CodeCanceled, ai.PhaseStream, "OpenAI Responses stream ended before a terminal response event")
		}
		for range waiting {
			o := waitValue(ev, "waiting call ends", results)
			checkTextSucceeded(ev, aw.text, o)
		}
		checkAdmissionReleased(ev, aw)
		// Idle keep-alive connections are pooled by the transport, not held
		// by any call.
		aw.transport.CloseIdleConnections()
		checkGoroutinesSettle(ev, baseline)
	})
}
