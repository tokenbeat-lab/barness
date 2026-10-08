package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestUnaryInitialFailures(t *testing.T) {
	for _, status := range []int{401, 403, 422, 429, 529, 0} {
		for _, retries := range []int{0, 2} {
			name := fmt.Sprintf("status-%d-retries-%d", status, retries)
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P07-E05-unary-"+name)
				clk := clock.NewFake(time.Unix(1790000000, 0), 0)
				u := newUnaryWorld(t, func(c *ai.Config) { c.Clock = clk })
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = retries })
				reply := fixtureReply{Status: status, Header: map[string]string{"retry-after-ms": "10", "x-typesafe-request-id": "vendor-fail"}, Body: `{"error":"fixture failure"}`}.script(t, classifierProtocol, "")
				if status == 0 {
					reply = provider.Reply{End: provider.EndDrop}
				}
				n := 1
				if status == 429 || status == 529 || status == 0 {
					n += retries
				}
				for range n {
					enqueue(ev, u.world, reply)
				}
				headers, payloads, responses := 0, 0, 0
				res, err := u.call(t, ctxFor(t), "req-"+name, ai.Hooks{
					TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { headers++; return nil },
					OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
						payloads++
						return ai.KeepPayload(), nil
					},
					OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { responses++; return nil },
				})
				code := ai.CodeUpstreamAuth
				switch status {
				case 422:
					code = ai.CodeInvalidRequest
				case 429:
					code = ai.CodeRateLimited
				case 529:
					code = ai.CodeUpstreamError
				case 0:
					code = ai.CodeTransport
				}
				checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
				ev.Check("exact network and attempt count", len(u.provider.Requests()) == n && len(res.Metadata.Attempts) == n, "requests=%d attempts=%+v", len(u.provider.Requests()), res.Metadata.Attempts)
				ev.Check("callbacks once per logical request", headers == 1 && payloads == 1 && responses == 0, "got %d/%d/%d", headers, payloads, responses)
				admissions := u.adm.Requests()
				ev.Record("admissions", admissions)
				ev.Record("sleeps", clk.Sleeps())
				ev.Check("independent admission per attempt", len(admissions) == n, "got %+v", admissions)
				for i, a := range res.Metadata.Attempts {
					ev.Check("failed attempt classification", a.Code == code && a.HTTPStatus == status && a.UsageReporting == ai.UsageUnreported, "got %+v", a)
					if i < len(admissions) {
						ev.Check("admission uses trusted snapshot", admissions[i].AttemptID == a.AttemptID && admissions[i].TenantID == tenantA.tenant && admissions[i].AccountScopeID == "acct-tenant-a", "got %+v", admissions[i])
					}
				}
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestUnaryFrozenRetry(t *testing.T) {
	ev := run.Case(t, "P07-E05-unary-frozen-snapshot")
	clk := clock.NewFake(time.Unix(1790000000, 0), 0)
	var source *ai.Catalog
	u := newUnaryWorld(t, func(c *ai.Config) { c.Clock = clk; source = c.Catalog })
	u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
	var retainedBody map[string]any
	var retainedHeaders http.Header
	headers, payloads, responses := 0, 0, 0
	failed := fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "10"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, "")
	failed.OnReceive = func() {
		u.host.PutCredential(primaryCredential(tenantA2, "v2"))
		u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Version = "b2"; b.Endpoint += "/new"; b.Retry.MaxRetries = 0 })
		source.Version = "changed"
		source.ClassifierModels[0].Cost.Input = 999
		retainedBody["state"] = "changed after freeze"
		retainedHeaders.Set("X-Frozen", "changed after freeze")
	}
	enqueue(ev, u.world, failed, u.sc.replies(t)[0])
	res, err := u.call(t, ctxFor(t), "req-frozen", ai.Hooks{
		TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			headers++
			h.Set("X-Frozen", "frozen")
			retainedHeaders = h
			return nil
		},
		OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			payloads++
			retainedBody = p.Body
			return ai.KeepPayload(), nil
		},
		OnResponse: func(_ context.Context, _ ai.CallScope, r ai.ResponseInfo) error {
			responses++
			ev.Check("only response metadata", r.Status == 200 && r.Operation == ai.OperationClassifier && r.ModelID == "host-jev", "got %+v", r)
			r.Header.Set("x-typesafe-request-id", "host-mutated")
			return nil
		},
	})
	ev.Record("result", res)
	reqs := u.provider.Requests()
	ev.Check("retry succeeds with one set of callbacks", err == nil && len(reqs) == 2 && headers == 1 && payloads == 1 && responses == 1, "err=%v callbacks=%d/%d/%d", err, headers, payloads, responses)
	if len(reqs) == 2 {
		ev.Check("same target credential headers and frozen bytes", reflect.DeepEqual(reqs[0], reqs[1]) && reqs[0].Path == "/v1/systemone" && reqs[0].KeyAlias == tenantA.alias && reqs[0].Header["X-Frozen"] == "frozen", "got %+v", reqs)
	}
	ev.Check("pinned versions catalog and price", res.Metadata.BindingVersion == "b1" && res.Metadata.CredentialVersion == "v1" && res.Metadata.CatalogVersion != "changed" && res.Usage.Cost.Input == 0.000296, "got %+v", res)
	ev.Check("one configuration resolution", u.host.ReadsFor("req-frozen").Bindings == 1 && u.host.ReadsFor("req-frozen").Credentials == 1, "got %+v", u.host.ReadsFor("req-frozen"))
	if len(res.Metadata.Attempts) == 2 {
		ev.Check("callback cannot mutate vendor metadata", res.Metadata.Attempts[1].ProviderRequestID == "req-typesafe", "got %+v", res.Metadata.Attempts)
	}
	u.records(ev, res, err)
	u.released(ev)
}

func TestUnaryRetryAdmissionDenied(t *testing.T) {
	for _, retry := range []bool{false, true} {
		name := "initial"
		if retry {
			name = "retry"
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E08-unary-admission-denied-"+name)
			u := newUnaryWorld(t, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) })
			sentinel := errors.New("synthetic admission failure")
			n := 0
			if retry {
				n = 1
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
				r := fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "10"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, "")
				r.OnReceive = func() { u.adm.Deny(sentinel) }
				enqueue(ev, u.world, r)
			} else {
				u.adm.Deny(sentinel)
			}
			res, err := u.call(t, ctxFor(t), "req-deny-"+name, ai.Hooks{})
			checkUnaryFailure(ev, res, err, ai.CodeAdmissionDenied, ai.PhaseAdmission, ai.StopReasonError)
			ev.Check("admission cause retained", errors.Is(err, sentinel), "got %v", err)
			ev.Check("refusal invents no sent attempt", len(u.provider.Requests()) == n && len(res.Metadata.Attempts) == n, "got %+v", res.Metadata)
			if retry {
				ev.Check("rate limit history retained", res.Metadata.Attempts[0].Code == ai.CodeRateLimited, "got %+v", res.Metadata.Attempts)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestUnaryBindingPolicySnapshot(t *testing.T) {
	ev := run.Case(t, "P07-E07-unary-binding-policy-snapshot")
	delay := 20 * time.Millisecond
	clk := clock.NewFake(time.Unix(1790000000, 0), 0)
	u := newUnaryWorld(t, func(c *ai.Config) {
		c.Clock = clk
		bindings := c.Bindings
		c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
			b, err := bindings.ResolveBinding(ctx, s, id)
			b.Retry = ai.RetryPolicy{MaxRetries: 1, MaxRetryDelay: &delay}
			return b, err
		})
		credentials := c.Credentials
		c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
			delay = time.Millisecond
			*b.Retry.MaxRetryDelay = time.Millisecond
			return credentials.ResolveCredential(ctx, s, b)
		})
	})
	enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "10"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""), u.sc.replies(t)[0])
	res, err := u.call(t, ctxFor(t), "req-policy-snapshot", ai.Hooks{})
	ev.Record("result", res)
	ev.Record("error", errString(err))
	ev.Check("resolved binding policy independently pinned", err == nil && len(res.Metadata.Attempts) == 2 && len(clk.Sleeps()) == 1 && clk.Sleeps()[0] == 10*time.Millisecond, "got %v attempts=%+v sleeps=%v", err, res.Metadata.Attempts, clk.Sleeps())
	u.records(ev, res, err)
	u.released(ev)
}
