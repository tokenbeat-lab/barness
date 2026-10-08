package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestUnaryTransportTimeouts(t *testing.T) {
	for _, name := range []string{"connect", "header", "header-retry", "idle", "body-cancel", "body-host-deadline", "body-call-deadline"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E08-unary-timeout-"+name)
			u := newUnaryWorld(t, func(c *ai.Config) {
				c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0)
				switch name {
				case "connect":
					c.Policy.ConnectTimeout = short
					c.Transport = provider.TrackBodies(provider.StalledDialTransport())
				case "header", "header-retry":
					c.Policy.ResponseHeaderTimeout = short
				case "idle":
					c.Policy.ReadIdleTimeout = short
				case "body-call-deadline":
					fitCallTimeout(c.Policy, short)
					c.Policy.ReadIdleTimeout = short
				}
			})
			if name == "header-retry" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			}
			reply := u.sc.replies(t)[0]
			switch name {
			case "header", "header-retry":
				reply = provider.Reply{End: provider.EndSilent}
			case "idle", "body-cancel", "body-host-deadline", "body-call-deadline":
				reply.End = provider.EndHold
			}
			if name != "connect" {
				enqueue(ev, u.world, reply)
			}
			if name == "header-retry" {
				enqueue(ev, u.world, u.sc.replies(t)[0])
			}
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			hooks := ai.Hooks{}
			if name == "body-cancel" {
				hooks.OnResponse = func(context.Context, ai.CallScope, ai.ResponseInfo) error { time.AfterFunc(short, cancel); return nil }
			}
			if name == "body-host-deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(ctxFor(t), short)
				defer cancel()
			}
			start := time.Now()
			res, err := u.call(t, ctx, "req-timeout-"+name, hooks)
			ev.Check("bounded completion", time.Since(start) < long, "took %s", time.Since(start))
			if name == "header-retry" {
				ev.Check("header timeout can retry", err == nil && len(res.Metadata.Attempts) == 2 && res.Metadata.Attempts[0].Code == ai.CodeDeadlineExceeded, "got %v %+v", err, res.Metadata)
			} else {
				code, phase, stop := ai.CodeDeadlineExceeded, ai.PhaseRequest, ai.StopReasonError
				if name == "idle" || name == "body-cancel" || name == "body-host-deadline" || name == "body-call-deadline" {
					phase = ai.PhaseResponse
				}
				if name == "body-cancel" {
					code = ai.CodeCanceled
				}
				if name == "body-cancel" || name == "body-host-deadline" || name == "body-call-deadline" {
					stop = ai.StopReasonAborted
				}
				checkUnaryFailure(ev, res, err, code, phase, stop)
				ev.Check("one actual attempt", len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata)
			}
			u.records(ev, res, err)
			u.released(ev)
			if name != "connect" {
				enqueue(ev, u.world, u.sc.replies(t)[0])
				next, e := u.call(t, ctxFor(t), "req-after-"+name, ai.Hooks{})
				ev.Check("subsequent call acquires permit", e == nil && len(next.Answers) == 1, "got %v", e)
				u.released(ev)
			}
		})
	}
}

func TestUnaryBackoffInterruption(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancel"
		code := ai.CodeCanceled
		if deadline {
			name = "deadline"
			code = ai.CodeDeadlineExceeded
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E05-unary-backoff-"+name)
			clk := clock.NewFake(time.Unix(1790000000, 0), 0)
			sleeping := clk.HoldSleeps()
			u := newUnaryWorld(t, func(c *ai.Config) {
				c.Clock = clk
				if deadline {
					fitCallTimeout(c.Policy, short)
				}
			})
			u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
			enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after": "2"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""))
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			done := make(chan struct{})
			var res ai.ClassifierResult
			var err error
			go func() { res, err = u.call(t, ctx, "req-backoff-"+name, ai.Hooks{}); close(done) }()
			select {
			case d := <-sleeping:
				ev.Check("requested backoff", d == 2*time.Second, "got %s", d)
			case <-time.After(long):
				t.Fatal("no backoff")
			}
			ev.Check("backoff owns no body or permit", u.bodies.Open() == 0 && u.gauges.Permits() == 0 && u.adm.Held() == 0, "bodies=%d permits=%d", u.bodies.Open(), u.gauges.Permits())
			if !deadline {
				cancel()
			}
			select {
			case <-done:
			case <-time.After(long):
				t.Fatal("call did not end")
			}
			checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonAborted)
			ev.Check("only failed attempt remains", len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].Code == ai.CodeRateLimited && len(u.provider.Requests()) == 1, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestUnaryAdmissionWaitIsolation(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancel"
		code := ai.CodeCanceled
		if deadline {
			name = "deadline"
			code = ai.CodeDeadlineExceeded
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E06-unary-admission-wait-"+name)
			u := newUnaryWorld(t, func(c *ai.Config) { c.Policy.MaxConcurrentPerTenant = 1; c.Policy.MaxConcurrentProcess = 2 })
			enqueue(ev, u.world, u.sc.replies(t)[0], u.sc.replies(t)[0], u.sc.replies(t)[0])
			held, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var owner ai.ClassifierResult
			var ownerErr error
			go func() {
				owner, ownerErr = u.call(t, ctxFor(t), "req-owner-"+name, ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
					close(held)
					select {
					case <-release:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				}})
				close(done)
			}()
			select {
			case <-held:
			case <-time.After(long):
				t.Fatal("owner not holding response")
			}
			ctx, cancel := context.WithCancel(ctxFor(t))
			if deadline {
				cancel()
				ctx, cancel = context.WithTimeout(ctxFor(t), short)
			}
			defer cancel()
			waitingDone := make(chan struct{})
			var waiter ai.ClassifierResult
			var waiterErr error
			go func() { waiter, waiterErr = u.call(t, ctx, "req-waiter-"+name, ai.Hooks{}); close(waitingDone) }()
			eventually(ev, "waiter entered bounded admission", func() bool { return u.gauges.AdmissionWaiters() == 1 })
			if !deadline {
				cancel()
			}
			select {
			case <-waitingDone:
			case <-time.After(long):
				t.Fatal("waiter did not end")
			}
			checkUnaryFailure(ev, waiter, waiterErr, code, ai.PhaseAdmission, ai.StopReasonAborted)
			ev.Check("no attempt invented for waiting call", len(waiter.Metadata.Attempts) == 0, "got %+v", waiter.Metadata)
			ev.Check("only owner resources remain", u.gauges.Permits() == 1 && u.bodies.Open() == 1 && u.gauges.ActiveCalls() == 1 && u.gauges.AdmissionWaiters() == 0, "calls=%d permits=%d bodies=%d waiters=%d", u.gauges.ActiveCalls(), u.gauges.Permits(), u.bodies.Open(), u.gauges.AdmissionWaiters())
			scope := textScope("req-other-" + name)
			scope.TenantID = tenantB.tenant
			other, e := u.client.Classify(ctxFor(t), scope, u.sc.target(), classifierInput(t, u.sc), nil)
			ev.Check("other tenant unaffected", e == nil && len(other.Answers) == 1, "got %v", e)
			close(release)
			select {
			case <-done:
			case <-time.After(long):
				t.Fatal("owner did not finish")
			}
			ev.Check("owner completes after waiter cancellation", ownerErr == nil && len(owner.Answers) == 1, "got %v", ownerErr)
			u.records(ev, waiter, waiterErr)
			u.records(ev, owner, ownerErr)
			u.released(ev)
			next, e := u.call(t, ctxFor(t), "req-followup-"+name, ai.Hooks{})
			ev.Check("same tenant can acquire again", e == nil && len(next.Answers) == 1, "got %v", e)
			u.released(ev)
		})
	}
}
