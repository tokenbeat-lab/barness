package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMixedCancellationIsolation(t *testing.T) {
	for i, r := range mixedRoutes {
		for _, mode := range []string{"cancel", "deadline", "http-failure"} {
			t.Run(r.id+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "E08-mixed-"+mode+"-isolation-"+r.id)
				w := newMixedWorld(t, nil, tenantA, tenantB)
				other := mixedRoutes[(i+1)%4]
				held, release := make(chan struct{}), make(chan struct{})
				var once sync.Once
				defer once.Do(func() { close(release) })
				w.provider.EnqueueAt(other.prefix(tenantB), other.reply(t))
				sibling := make(chan mixedOutcome, 1)
				go func() {
					sibling <- other.invokeHooked(ctxFor(t), w.client, scopeFor(tenantB, "req-held-sibling"), ai.Target{BindingID: other.id, ModelID: mixedModel}, ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
						close(held)
						select {
						case <-release:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}})
				}()
				waitFor(ev, "sibling owns body and permit", held)
				arrived := make(chan struct{})
				reply := provider.Reply{End: provider.EndSilent, OnReceive: func() { close(arrived) }}
				if mode == "http-failure" {
					reply.End = provider.EndClose
					reply.Status = 422
					reply.Header = map[string]string{"Content-Type": "application/json"}
					reply.Chunks = [][]byte{[]byte(`{"error":"synthetic failure"}`)}
				}
				w.provider.EnqueueAt(r.prefix(tenantA), reply)
				ctx, cancel := context.WithCancel(ctxFor(t))
				if mode == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(ctxFor(t), short)
				}
				defer cancel()
				done := make(chan mixedOutcome, 1)
				go func() {
					done <- r.invoke(ctx, w.client, textScope("req-ended-subject"), ai.Target{BindingID: r.id, ModelID: mixedModel})
				}()
				waitFor(ev, "subject sent", arrived)
				if mode == "cancel" {
					cancel()
				}
				var o mixedOutcome
				select {
				case o = <-done:
				case <-time.After(long):
					t.Fatal("subject failed to end")
				}
				ev.Check("subject failed without published unary output", o.err != nil && len(o.images.Content) == 0 && len(o.classifier.Answers) == 0, "got %v", o.err)
				if mode != "http-failure" {
					code := ai.CodeCanceled
					if mode == "deadline" {
						code = ai.CodeDeadlineExceeded
					}
					ev.Check("correct termination classification", errors.Is(o.err, &ai.Error{Code: code}), "got %v", o.err)
				}
				ev.Check("only subject resources released", w.gauges.Permits() == 1 && w.bodies.Open() == 1 && w.gauges.ActiveCalls() == 1, "permits=%d bodies=%d calls=%d", w.gauges.Permits(), w.bodies.Open(), w.gauges.ActiveCalls())
				w.record(ev, o)
				once.Do(func() { close(release) })
				kept := <-sibling
				w.record(ev, kept)
				ev.Check("sibling result not polluted", kept.err == nil && kept.metadata.TenantID == tenantB.tenant, "got %v", kept.err)
				w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
				next := r.invoke(ctxFor(t), w.client, textScope("req-after-ended"), ai.Target{BindingID: r.id, ModelID: mixedModel})
				w.record(ev, next)
				ev.Check("same operation reacquires returned permit", next.err == nil, "got %v", next.err)
				ev.Record("requests", w.provider.Requests())
				w.released(ev)
			})
		}
	}
}

func TestMixedAdmissionLimits(t *testing.T) {
	for i, owner := range mixedRoutes {
		for _, mode := range []string{"tenant-cancel", "process-cancel", "wait-timeout"} {
			t.Run(owner.id+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "E08-mixed-admission-"+owner.id+"-"+mode)
				w := newMixedWorld(t, func(c *ai.Config) {
					c.Policy.MaxConcurrentPerTenant = 1
					c.Policy.MaxConcurrentProcess = 2
					c.Policy.MaxAdmissionWaiters = 1
					if mode == "process-cancel" {
						c.Policy.MaxConcurrentProcess = 1
					}
					if mode == "wait-timeout" {
						c.Policy.AdmissionWait = short
					}
				}, tenantA, tenantB)
				held, release := make(chan struct{}), make(chan struct{})
				var once sync.Once
				defer once.Do(func() { close(release) })
				reply := owner.reply(t)
				reply.End = provider.EndHold
				reply.OnHold = func() { close(held) }
				// Hold before EOF: the complete body is available, but a unary
				// response has not been fully read or validated. Chat uses a
				// response callback barrier because its terminal can finish early.
				hooks := ai.Hooks{}
				if owner.op == ai.OperationChat {
					reply.End = provider.EndClose
					reply.OnHold = nil
					hooks.OnResponse = func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
						close(held)
						select {
						case <-release:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}
				w.provider.EnqueueAt(owner.prefix(tenantA), reply)
				ownerCtx, endOwner := context.WithCancel(ctxFor(t))
				defer endOwner()
				old := make(chan mixedOutcome, 1)
				go func() {
					old <- owner.invokeHooked(ownerCtx, w.client, textScope("req-admission-owner"), ai.Target{BindingID: owner.id, ModelID: mixedModel}, hooks)
				}()
				waitFor(ev, "owner holds complete response and permit", held)
				waiter := mixedRoutes[(i+1)%4]
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				waiting := make(chan mixedOutcome, 1)
				scope := textScope("req-mixed-waiter")
				if mode == "process-cancel" {
					scope.TenantID = tenantB.tenant
				}
				go func() {
					waiting <- waiter.invoke(ctx, w.client, scope, ai.Target{BindingID: waiter.id, ModelID: mixedModel})
				}()
				eventually(ev, "one bounded mixed waiter", func() bool { return w.gauges.AdmissionWaiters() == 1 })
				overflow := mixedRoutes[(i+2)%4].invoke(ctxFor(t), w.client, scopeFor(tenantA, "req-waiter-overflow"), ai.Target{BindingID: mixedRoutes[(i+2)%4].id, ModelID: mixedModel})
				ev.Check("waiter cap rejects without request", errors.Is(overflow.err, &ai.Error{Code: ai.CodeAdmissionDenied}) && len(overflow.metadata.Attempts) == 0, "got %v", overflow.err)
				if mode != "wait-timeout" {
					cancel()
				}
				var denied mixedOutcome
				select {
				case denied = <-waiting:
				case <-time.After(long):
					t.Fatal("waiter failed to end")
				}
				code := ai.CodeCanceled
				if mode == "wait-timeout" {
					code = ai.CodeAdmissionDenied
				}
				ev.Check("wait terminates within bounded policy", errors.Is(denied.err, &ai.Error{Code: code, Phase: ai.PhaseAdmission}) && len(denied.metadata.Attempts) == 0, "got %v", denied.err)
				ev.Check("waiter never released owner", w.gauges.Permits() == 1 && w.bodies.Open() == 1 && w.gauges.AdmissionWaiters() == 0 && len(w.provider.Requests()) == 1, "held=%d bodies=%d waiters=%d sent=%d", w.gauges.Permits(), w.bodies.Open(), w.gauges.AdmissionWaiters(), len(w.provider.Requests()))
				if mode != "process-cancel" {
					w.provider.EnqueueAt(waiter.prefix(tenantB), waiter.reply(t))
					other := waiter.invoke(ctxFor(t), w.client, scopeFor(tenantB, "req-other-tenant-admitted"), ai.Target{BindingID: waiter.id, ModelID: mixedModel})
					ev.Check("other tenant uses its process permit", other.err == nil, "got %v", other.err)
					w.record(ev, other)
				}
				if owner.op == ai.OperationChat {
					once.Do(func() { close(release) })
				} else {
					endOwner()
				}
				o := <-old
				w.record(ev, o)
				w.record(ev, denied)
				w.record(ev, overflow)
				w.released(ev)
			})
		}
	}
}
