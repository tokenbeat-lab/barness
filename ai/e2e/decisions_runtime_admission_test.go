package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsRuntimeAdmissionIsolation(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name, code := "cancel", ai.CodeCanceled
		if deadline {
			name, code = "deadline", ai.CodeDeadlineExceeded
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD42-runtime-mixed-admission-"+name)
			w := newRuntimeMixedWorld(t, true, func(c *ai.Config) { c.Policy.MaxConcurrentPerTenant = 1; c.Policy.MaxConcurrentProcess = 2 })
			image, luna, chat := runtimeMixedRoutes[1], runtimeMixedRoutes[3], runtimeMixedRoutes[0]
			runtimeMixedEnqueue(ev, w, image, tenantA, runtimeMixedReply(t, image, tenantA))
			held, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			defer unblock()
			owner := make(chan mixedOutcome, 1)
			go func() {
				owner <- runtimeMixedInvoke(t, ctxFor(t), w, image, tenantA, "req-owner", ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
					close(held)
					select {
					case <-release:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				}}, false)
			}()
			waitFor(ev, "image holds shared permit", held)
			ctx, cancel := context.WithCancel(ctxFor(t))
			if deadline {
				cancel()
				ctx, cancel = context.WithTimeout(ctxFor(t), short)
			}
			defer cancel()
			waiting := make(chan mixedOutcome, 1)
			go func() { waiting <- runtimeMixedInvoke(t, ctx, w, luna, tenantA, "req-waiter", ai.Hooks{}, true) }()
			eventually(ev, "Luna waits behind image operation", func() bool { return w.gauges.AdmissionWaiters() == 1 })
			if !deadline {
				cancel()
			}
			var waiter mixedOutcome
			select {
			case waiter = <-waiting:
			case <-time.After(long):
				t.Fatal("waiter never ended")
			}
			waiter.recordResult(ev, "waiter-result")
			w.record(ev, waiter)
			ev.Check("waiter cancellation creates no attempt", errors.Is(waiter.err, &ai.Error{Code: code, Phase: ai.PhaseAdmission}) && len(waiter.metadata.Attempts) == 0, "got %v", waiter.err)
			ev.Record("owner-only-resources", map[string]any{"calls": w.gauges.ActiveCalls(), "permits": w.gauges.Permits(), "bodies": w.bodies.Open(), "waiters": w.gauges.AdmissionWaiters(), "hostPermits": w.adm.Held()})
			ev.Check("waiter only releases own resources", w.gauges.ActiveCalls() == 1 && w.gauges.Permits() == 1 && w.bodies.Open() == 1 && w.adm.Held() == 1 && w.gauges.AdmissionWaiters() == 0, "owner lease damaged")
			runtimeMixedEnqueue(ev, w, chat, tenantB, runtimeMixedReply(t, chat, tenantB))
			other := runtimeMixedInvoke(t, ctxFor(t), w, chat, tenantB, "req-other-tenant", ai.Hooks{}, false)
			other.recordResult(ev, "other-result")
			w.record(ev, other)
			ev.Check("shared provider account does not merge tenant admission", other.err == nil, "got %v", other.err)
			unblock()
			var old mixedOutcome
			select {
			case old = <-owner:
			case <-time.After(long):
				t.Fatal("owner never ended")
			}
			w.record(ev, old)
			old.recordResult(ev, "owner-result")
			ev.Check("owner succeeds", old.err == nil, "got %v", old.err)
			runtimeMixedEnqueue(ev, w, luna, tenantA, runtimeMixedReply(t, luna, tenantA))
			next := runtimeMixedInvoke(t, ctxFor(t), w, luna, tenantA, "req-after-waiter", ai.Hooks{}, true)
			w.record(ev, next)
			next.recordResult(ev, "recovery-result")
			ev.Check("Luna reacquires after owner finishes", next.err == nil, "got %v", next.err)
			recordRuntimeMixedConfig(ev, w)
			ev.Record("requests", w.provider.Requests())
			w.released(ev)
		})
	}
}
