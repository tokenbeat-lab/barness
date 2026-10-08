package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestUnaryEndedBeforeStart(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "canceled"
		code := ai.CodeCanceled
		if deadline {
			name = "deadline"
			code = ai.CodeDeadlineExceeded
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E08-unary-before-"+name)
			u := newUnaryWorld(t, nil)
			ctx, cancel := context.WithCancel(ctxFor(t))
			if deadline {
				cancel()
				ctx, cancel = context.WithDeadline(ctxFor(t), time.Now().Add(-time.Second))
			}
			cancel()
			res, err := u.call(t, ctx, "req-before-"+name, ai.Hooks{})
			checkUnaryFailure(ev, res, err, code, ai.PhaseScope, ai.StopReasonError)
			reads := u.host.ReadsFor(res.Metadata.RequestID)
			ev.Check("no resolution or HTTP", reads.Bindings == 0 && reads.Credentials == 0 && len(u.provider.Requests()) == 0, "got %+v", reads)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestUnaryResolutionInterruption(t *testing.T) {
	for _, stage := range []string{"binding", "credential"} {
		for _, deadline := range []bool{false, true} {
			for _, success := range []bool{false, true} {
				name := stage + "-cancel"
				code := ai.CodeCanceled
				if deadline {
					name = stage + "-deadline"
					code = ai.CodeDeadlineExceeded
				}
				if success {
					name += "-resolver-returned-success"
				}
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "P07-E08-unary-resolve-"+name)
					ctx, cancel := context.WithCancel(ctxFor(t))
					defer cancel()
					end := func(ctx context.Context) {
						if deadline {
							<-ctx.Done()
						} else {
							cancel()
						}
					}
					u := newUnaryWorld(t, func(c *ai.Config) {
						if deadline {
							fitCallTimeout(c.Policy, short)
						}
						if stage == "binding" {
							original := c.Bindings
							c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
								b, err := original.ResolveBinding(ctx, s, id)
								end(ctx)
								if !success {
									return ai.Binding{}, ctx.Err()
								}
								return b, err
							})
						} else {
							original := c.Credentials
							c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
								cred, err := original.ResolveCredential(ctx, s, b)
								end(ctx)
								if !success {
									return ai.Credential{}, ctx.Err()
								}
								return cred, err
							})
						}
					})
					res, err := u.call(t, ctx, "req-resolve-"+name, ai.Hooks{})
					phase := ai.PhaseBinding
					if stage == "credential" {
						phase = ai.PhaseCredential
					}
					checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
					ev.Check("unresolved without attempts", !res.Metadata.Resolved && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
					if stage == "binding" {
						ev.Check("no credential read after termination", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "got %+v", u.host.ReadsFor(res.Metadata.RequestID))
					}
					u.records(ev, res, err)
					u.released(ev)
				})
			}
		}
	}
}

type unaryAdmissionFunc func(context.Context, ai.AdmissionRequest) (func(), error)

func (f unaryAdmissionFunc) Admit(ctx context.Context, r ai.AdmissionRequest) (func(), error) {
	return f(ctx, r)
}

func TestUnaryHostAdmissionInterruption(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		for _, grant := range []bool{false, true} {
			name := "cancel"
			code := ai.CodeCanceled
			if deadline {
				name = "deadline"
				code = ai.CodeDeadlineExceeded
			}
			if grant {
				name += "-late-grant"
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P07-E08-unary-host-admission-"+name)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				released := 0
				u := newUnaryWorld(t, func(c *ai.Config) {
					if deadline {
						fitCallTimeout(c.Policy, short)
					}
					c.Admission = unaryAdmissionFunc(func(ctx context.Context, _ ai.AdmissionRequest) (func(), error) {
						if deadline {
							<-ctx.Done()
						} else {
							cancel()
						}
						if grant {
							return func() { released++ }, nil
						}
						return nil, ctx.Err()
					})
				})
				res, err := u.call(t, ctx, "req-host-admission-"+name, ai.Hooks{})
				checkUnaryFailure(ev, res, err, code, ai.PhaseAdmission, ai.StopReasonAborted)
				ev.Check("no transmitted attempt", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
				want := 0
				if grant {
					want = 1
				}
				ev.Check("late permit returned once", released == want, "got %d", released)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}
