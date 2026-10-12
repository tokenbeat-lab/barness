package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestDecisionsRuntimeCancellation(t *testing.T) {
	for _, stage := range []string{"binding", "credential", "headers", "payload", "host-admission"} {
		for _, deadline := range []bool{false, true} {
			name := stage + "-cancel"
			code := ai.CodeCanceled
			if deadline {
				name = stage + "-deadline"
				code = ai.CodeDeadlineExceeded
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P10-OD42-runtime-"+name)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				end := func(ctx context.Context) {
					if deadline {
						<-ctx.Done()
					} else {
						cancel()
					}
				}
				grants, releases := 0, 0
				u := runtimeDecisionsWorld(t, func(c *ai.Config) {
					if deadline {
						fitCallTimeout(c.Policy, short)
					}
					switch stage {
					case "binding":
						original := c.Bindings
						c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
							b, e := original.ResolveBinding(ctx, s, id)
							end(ctx)
							return b, e
						})
					case "credential":
						original := c.Credentials
						c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
							cred, e := original.ResolveCredential(ctx, s, b)
							end(ctx)
							return cred, e
						})
					case "host-admission":
						c.Admission = unaryAdmissionFunc(func(ctx context.Context, _ ai.AdmissionRequest) (func(), error) {
							end(ctx)
							grants++
							return func() { releases++ }, nil
						})
					}
				})
				hooks := ai.Hooks{}
				if stage == "headers" {
					hooks.TransformHeaders = func(ctx context.Context, _ ai.CallScope, _ http.Header) error { end(ctx); return nil }
				}
				if stage == "payload" {
					hooks.OnPayload = func(ctx context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
						end(ctx)
						return ai.KeepPayload(), nil
					}
				}
				res, err := runtimeDecisionsCall(ctx, u, "req-cancel-"+name, decisionsInput(), hooks)
				phase, stop := ai.PhaseRequest, ai.StopReasonAborted
				switch stage {
				case "binding":
					phase = ai.PhaseBinding
					stop = ai.StopReasonError
				case "credential":
					phase = ai.PhaseCredential
					stop = ai.StopReasonError
				case "host-admission":
					phase = ai.PhaseAdmission
					stop = ai.StopReasonAborted
				}
				checkUnaryFailure(ev, res, err, code, phase, stop)
				ev.Check("no attempt after canceled boundary", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
				if stage == "binding" || stage == "credential" {
					ev.Check("late resolver success stays unresolved", !res.Metadata.Resolved, "resolved ended call")
				}
				if stage == "binding" {
					ev.Check("no credential after binding cancellation", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "credential read")
				}
				if stage == "host-admission" {
					ev.Check("late granted permit returned once", grants == 1 && releases == 1, "grants=%d releases=%d", grants, releases)
				}
				recordRuntimeConfig(ev, u)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsRuntimeTimeouts(t *testing.T) {
	for _, name := range []string{"connect", "header", "header-retry", "idle", "body-cancel", "body-host-deadline", "body-call-deadline"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD42-runtime-timeout-"+name)
			u := runtimeDecisionsWorld(t, func(c *ai.Config) {
				c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0)
				switch name {
				case "connect":
					c.Transport = provider.TrackBodies(provider.StalledDialTransport())
					c.Policy.ConnectTimeout = short
				case "header", "header-retry":
					c.Policy.ResponseHeaderTimeout = short
				case "idle":
					c.Policy.ReadIdleTimeout = short
				case "body-call-deadline":
					fitCallTimeout(c.Policy, short)
					c.Policy.ReadIdleTimeout = short
				}
			})
			reply := runtimeDecisionsReply(t)
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
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				enqueue(ev, u.world, runtimeDecisionsReply(t))
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
			res, err := runtimeDecisionsCall(ctx, u, "req-timeout-"+name, decisionsInput(), hooks)
			ev.Record("elapsed-ms", time.Since(start).Milliseconds())
			ev.Record("result", res)
			ev.Check("timeout bounded", time.Since(start) < long, "took %s", time.Since(start))
			if name == "header-retry" {
				ev.Check("initial header timeout may retry", err == nil && len(res.Metadata.Attempts) == 2 && res.Metadata.Attempts[0].Code == ai.CodeDeadlineExceeded, "got %v / %+v", err, res.Metadata)
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
				ev.Check("one sent attempt without invented usage", len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "got %+v", res.Metadata)
			}
			recordRuntimeConfig(ev, u)
			u.records(ev, res, err)
			u.released(ev)
			if name != "connect" {
				enqueue(ev, u.world, runtimeDecisionsReply(t))
				next, e := runtimeDecisionsCall(ctxFor(t), u, "req-timeout-recovery-"+name, decisionsInput(), ai.Hooks{})
				ev.Check("next call obtains released permit", e == nil && len(next.Answers) == 1, "got %v", e)
				u.records(ev, next, e)
				u.released(ev)
			}
		})
	}
}

func TestDecisionsRuntimeBackoffCancellation(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name, code := "cancel", ai.CodeCanceled
		if deadline {
			name, code = "deadline", ai.CodeDeadlineExceeded
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD42-runtime-backoff-"+name)
			clk := clock.NewFake(time.Unix(1790000000, 0), 0)
			sleeping := clk.HoldSleeps()
			u := runtimeDecisionsWorld(t, func(c *ai.Config) {
				c.Clock = clk
				if deadline {
					fitCallTimeout(c.Policy, short)
				}
			})
			u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
			enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after": "2"}, Body: `{"error":{"message":"limited"}}`}.script(t, decisionsProtocol, ""))
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			done := make(chan struct{})
			var res ai.ClassifierResult
			var err error
			go func() {
				res, err = runtimeDecisionsCall(ctx, u, "req-backoff-"+name, decisionsInput(), ai.Hooks{})
				close(done)
			}()
			select {
			case d := <-sleeping:
				ev.Check("explicit backoff", d == 2*time.Second, "got %s", d)
			case <-time.After(long):
				t.Fatal("no backoff")
			}
			ev.Check("backoff releases body and permits", u.bodies.Open() == 0 && u.gauges.Permits() == 0 && u.adm.Held() == 0, "resources retained")
			if !deadline {
				cancel()
			}
			waitFor(ev, "call ends", done)
			checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonAborted)
			ev.Check("no request after backoff cancellation", len(u.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].Code == ai.CodeRateLimited, "got %+v", res.Metadata)
			ev.Record("sleeps", clk.Sleeps())
			recordRuntimeConfig(ev, u)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
