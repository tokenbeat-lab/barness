package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMultimodalRuntimeCancellation(t *testing.T) {
	for _, r := range multimodalRoutes {
		for _, entry := range []string{"plain", "hooked"} {
			for _, stage := range []string{"pre-send", "body", "backoff", "host"} {
				for _, deadline := range []bool{false, true} {
					name := r.id + "-" + entry + "-" + stage + "-cancel"
					code := ai.CodeCanceled
					if deadline {
						name = r.id + "-" + entry + "-" + stage + "-deadline"
						code = ai.CodeDeadlineExceeded
					}
					t.Run(name, func(t *testing.T) {
						ev := run.Case(t, "UC09-cancel-"+name)
						ctx, cancel := context.WithCancel(ctxFor(t))
						defer cancel()
						reached := make(chan struct{}, 1)
						clk := clock.NewFake(time.Unix(1790000000, 0), 0)
						sleeping := clk.HoldSleeps()
						grants, releases := 0, 0
						var meter *unaryMeter
						u := newMultimodalWorld(t, ev, func(c *ai.Config) {
							c.Clock = clk
							if stage == "body" {
								meter = &unaryMeter{next: c.Transport, held: func() bool { return c.Probe.Permits() == 1 }}
								c.Transport = meter
							}
							if deadline {
								fitCallTimeout(c.Policy, 250*time.Millisecond)
							}
							if stage == "pre-send" {
								previous := c.Credentials
								c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
									cred, e := previous.ResolveCredential(ctx, s, b)
									reached <- struct{}{}
									<-ctx.Done()
									return cred, e
								})
							}
							if stage == "host" {
								c.Admission = unaryAdmissionFunc(func(ctx context.Context, _ ai.AdmissionRequest) (func(), error) {
									reached <- struct{}{}
									<-ctx.Done()
									grants++
									return func() { releases++ }, nil
								})
							}
						})
						req := multimodalInput(t, r, tenantA)
						if stage == "body" {
							multimodalEnqueue(ev, u, r, tenantA, provider.Reply{Status: 200, Chunks: [][]byte{[]byte(`{"answers":`)}, Script: `{"answers":`, End: provider.EndHold, OnHold: func() { reached <- struct{}{} }})
						}
						if stage == "backoff" {
							b := u.host.Binding(tenantA.tenant, r.id)
							b.Retry.MaxRetries = 2
							u.host.PutBinding(b)
							multimodalEnqueue(ev, u, r, tenantA, provider.Reply{Status: 429, Header: map[string]string{"retry-after-ms": "10"}, Chunks: [][]byte{[]byte(`{"error":"synthetic private error"}`)}, Script: `{"error":"synthetic private error"}`})
						}
						var res ai.ClassifierResult
						var err error
						done := make(chan struct{})
						go func() {
							res, err = multimodalCall(ctx, u, r, tenantA, entry, "req-"+name, req, ai.Hooks{})
							close(done)
						}()
						if stage == "backoff" {
							select {
							case <-sleeping:
							case <-time.After(long):
								t.Fatal("no retry wait")
							}
							ev.Check("retry wait releases resources", u.bodies.Open() == 0 && u.gauges.Permits() == 0 && u.adm.Held() == 0, "retained retry resources")
						} else {
							waitFor(ev, "boundary reached", reached)
							if stage == "body" {
								eventually(ev, "body actually read", func() bool { return meter.read.Load() > 0 })
							}
						}
						if !deadline {
							cancel()
						}
						waitFor(ev, "canceled call finishes", done)
						phase, stop, attempts := ai.PhaseRequest, ai.StopReasonAborted, 0
						switch stage {
						case "pre-send":
							phase = ai.PhaseCredential
							stop = ai.StopReasonError
						case "body":
							phase = ai.PhaseResponse
							attempts = 1
						case "backoff":
							attempts = 1
						case "host":
							phase = ai.PhaseAdmission
						}
						checkUnaryFailure(ev, res, err, code, phase, stop)
						ev.Check("bounded sent history", len(res.Metadata.Attempts) == attempts && len(u.provider.Requests()) == attempts, "got %d", len(res.Metadata.Attempts))
						if stage == "backoff" {
							ev.Check("failed Attempt survives interrupted wait", res.Metadata.Attempts[0].Code == ai.CodeRateLimited && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "lost failure history")
						}
						if stage == "host" {
							ev.Check("late Host grant returned once", grants == 1 && releases == 1, "%d/%d", grants, releases)
						}
						recordMultimodalCall(ev, u, req, res, err)
						ev.Record("sleeps", clk.Sleeps())
						u.released(ev)
					})
				}
			}
		}
	}
}

func TestMultimodalRuntimeAdmissionIsolation(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		for _, limit := range []string{"tenant", "process"} {
			for _, deadline := range []bool{false, true} {
				name := entry + "-" + limit + "-cancel"
				code := ai.CodeCanceled
				if deadline {
					name = entry + "-" + limit + "-deadline"
					code = ai.CodeDeadlineExceeded
				}
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "UC09-admission-"+name)
					held, release := make(chan struct{}), make(chan struct{})
					var once sync.Once
					unblock := func() { once.Do(func() { close(release) }) }
					defer unblock()
					var eof *mixedEOFTransport
					u := newMultimodalWorld(t, ev, func(c *ai.Config) {
						c.Policy.MaxConcurrentPerTenant = 1
						if limit == "process" {
							c.Policy.MaxConcurrentProcess = 1
						}
						eof = &mixedEOFTransport{next: c.Transport, ended: func() { close(held) }, release: release}
						c.Transport = eof
					})
					ownerRoute, waiterRoute := multimodalRoutes[0], multimodalRoutes[1]
					ownerReq := multimodalInput(t, ownerRoute, tenantA)
					multimodalEnqueue(ev, u, ownerRoute, tenantA, multimodalReply(t, ev, ownerRoute, false))
					ownerDone := make(chan struct{})
					var owner ai.ClassifierResult
					var ownerErr error
					go func() {
						owner, ownerErr = multimodalCall(ctxFor(t), u, ownerRoute, tenantA, entry, "req-owner-"+name, ownerReq, ai.Hooks{})
						close(ownerDone)
					}()
					waitFor(ev, "owner holds complete body", held)
					ctx, cancel := context.WithCancel(ctxFor(t))
					if deadline {
						cancel()
						ctx, cancel = context.WithTimeout(ctxFor(t), 250*time.Millisecond)
					}
					defer cancel()
					k := tenantA
					if limit == "process" {
						k = tenantB
					}
					req := multimodalInput(t, waiterRoute, k)
					done := make(chan struct{})
					var res ai.ClassifierResult
					var err error
					go func() {
						res, err = multimodalCall(ctx, u, waiterRoute, k, entry, "req-waiter-"+name, req, ai.Hooks{})
						close(done)
					}()
					eventually(ev, "mixed caller waits", func() bool { return u.gauges.AdmissionWaiters() == 1 })
					if !deadline {
						cancel()
					}
					waitFor(ev, "waiter ends", done)
					ev.Check("waiter ends before any Attempt", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseAdmission}) && len(res.Metadata.Attempts) == 0, "got %v", err)
					ev.Check("waiter does not release owner", u.gauges.ActiveCalls() == 1 && u.gauges.Permits() == 1 && u.bodies.Open() == 1 && u.adm.Held() == 1 && u.gauges.AdmissionWaiters() == 0, "owner lease damaged")
					recordMultimodalCall(ev, u, req, res, err)
					unblock()
					waitFor(ev, "owner ends", ownerDone)
					checkMultimodalSuccess(t, ev, ownerRoute, ownerReq, owner, ownerErr)
					recordMultimodalCall(ev, u, ownerReq, owner, ownerErr)
					u.released(ev)
					// Recovery uses the other protocol through the same exhausted limiter.
					// The EOF barrier is now open; count callbacks may recur only once.
					eof.ended = func() {}
					multimodalEnqueue(ev, u, waiterRoute, k, multimodalReply(t, ev, waiterRoute, false))
					next, e := multimodalCall(ctxFor(t), u, waiterRoute, k, entry, "req-recovery-"+name, req, ai.Hooks{})
					checkMultimodalSuccess(t, ev, waiterRoute, req, next, e)
					recordMultimodalCall(ev, u, req, next, e)
					u.released(ev)
				})
			}
		}
	}
}
