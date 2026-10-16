package e2e

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMultimodalRuntimeRetry(t *testing.T) {
	for _, r := range multimodalRoutes {
		for _, entry := range []string{"plain", "hooked"} {
			for _, mode := range []string{"default-429", "default-503", "finite-success", "exhausted", "bad-2xx"} {
				t.Run(r.id+"/"+entry+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "UC09-retry-"+r.id+"-"+entry+"-"+mode)
					clk := clock.NewFake(time.Unix(1790000000, 0), 0)
					u := newMultimodalWorld(t, ev, func(c *ai.Config) { c.Clock = clk })
					req := multimodalInput(t, r, tenantA)
					b := u.host.Binding(tenantA.tenant, r.id)
					attempts := 1
					if mode == "finite-success" || mode == "bad-2xx" {
						b.Retry.MaxRetries = 1
					}
					if mode == "finite-success" {
						attempts = 2
					}
					if mode == "exhausted" {
						b.Retry.MaxRetries = 2
						attempts = 3
					}
					u.host.PutBinding(b)
					ev.Record("retry-binding", b)
					status := 429
					if mode == "default-503" {
						status = 503
					}
					for i := range attempts {
						reply := provider.Reply{Status: status, Header: map[string]string{"retry-after-ms": "10", "x-request-id": "up-failed", "x-typesafe-request-id": "up-failed"}, Chunks: [][]byte{[]byte(`{"error":"synthetic private error"}`)}, Script: `{"error":"synthetic private error"}`}
						if mode == "bad-2xx" || (mode == "finite-success" && i == attempts-1) {
							reply = multimodalReply(t, ev, r, mode == "bad-2xx")
						}
						multimodalEnqueue(ev, u, r, tenantA, reply)
					}
					counts := [3]int{}
					hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
						counts[0]++
						h.Set("X-Frozen", "original")
						return nil
					}, OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
						counts[1]++
						return ai.KeepPayload(), nil
					}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { counts[2]++; return nil }}
					res, err := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-retry-"+r.id+entry+mode, req, hooks)
					ev.Check("explicit finite Attempts", len(res.Metadata.Attempts) == attempts && len(u.provider.Requests()) == attempts, "wrong replay count")
					if mode == "finite-success" {
						checkMultimodalSuccess(t, ev, r, req, res, err)
					} else {
						code, phase := ai.CodeRateLimited, ai.PhaseRequest
						if mode == "default-503" {
							code = ai.CodeUpstreamError
						}
						if mode == "bad-2xx" {
							code, phase = ai.CodeProtocol, ai.PhaseResponse
						}
						checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
					}
					if mode == "bad-2xx" {
						ev.Check("legal usage survives atomic failure", res.Usage.Input > 0 && res.Metadata.Attempts[0].UsageReporting != ai.UsageUnreported, "usage lost")
					}
					if entry == "hooked" {
						response := 0
						if mode == "bad-2xx" || mode == "finite-success" {
							response = 1
						}
						ev.Check("Hooks once per Logical Call", counts == [3]int{1, 1, response}, "got %v", counts)
					}
					for i, a := range res.Metadata.Attempts {
						if i < attempts-1 || mode == "exhausted" || mode == "default-429" || mode == "default-503" {
							ev.Check("failure history retains status and no invented usage", a.HTTPStatus == status && a.ProviderRequestID == "up-failed" && a.UsageReporting == ai.UsageUnreported && a.Usage == (ai.Usage{}), "lost history")
						}
						if i > 0 {
							ev.Check("retry bytes header key credential fixed", reflect.DeepEqual(u.provider.Requests()[0], u.provider.Requests()[i]), "changed Attempt")
						}
					}
					ev.Check("one credential snapshot", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 1 && res.Metadata.CredentialVersion == "v1", "re-resolved credential")
					ev.Record("hooks", counts)
					ev.Record("sleeps", clk.Sleeps())
					recordMultimodalCall(ev, u, req, res, err)
					u.released(ev)
				})
			}
		}
	}
}

func TestMultimodalRuntimeRevocation(t *testing.T) {
	for _, r := range multimodalRoutes {
		for _, entry := range []string{"plain", "hooked"} {
			for _, mode := range []string{"binding", "model", "images"} {
				t.Run(r.id+"/"+entry+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "UC09-revocation-"+r.id+"-"+entry+"-"+mode)
					var source *ai.Catalog
					u := newMultimodalWorld(t, ev, func(c *ai.Config) { source = c.Catalog; c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) })
					req := multimodalInput(t, r, tenantA)
					var original ai.ClassifierRequest
					mustUnmarshal(t, mustMarshal(t, req), &original)
					catalog := u.client.Catalog()
					hash, _ := catalog.Hash()
					b := u.host.Binding(tenantA.tenant, r.id)
					b.Retry.MaxRetries = 1
					u.host.PutBinding(b)
					var borrowed map[string]any
					var headers http.Header
					fail := provider.Reply{Status: 429, Header: map[string]string{"retry-after-ms": "10"}, Chunks: [][]byte{[]byte(`{"error":"synthetic private error"}`)}, Script: `{"error":"synthetic private error"}`, OnReceive: func() {
						changed := b
						changed.Version = "b2"
						changed.Retry.MaxRetries = 0
						switch mode {
						case "binding":
							changed.Enabled = false
						case "model":
							changed.AllowedModels = nil
						case "images":
							changed.AllowClassifierImages = false
						}
						u.host.PutBinding(changed)
						ev.Record("revoked-binding", changed)
						u.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "v2", Active: true, APIKey: ai.NewSecret(tenantA2.secret)})
						for i := range source.ClassifierModels {
							source.ClassifierModels[i].Cost.Input = 999
						}
						req.Images[0] = req.Images[1]
						req.State[1] = 'X'
						req.Questions["intent"] = nil
						req.ImageReferences[0].ImageIndex = 0
						if borrowed != nil {
							borrowed["state"] = "changed"
							borrowed["input"] = "changed"
							borrowed["questions"] = nil
							borrowed["assets"] = nil
						}
						if headers != nil {
							headers.Set("X-Frozen", "changed")
						}
					}}
					multimodalEnqueue(ev, u, r, tenantA, fail, multimodalReply(t, ev, r, false))
					counts := [3]int{}
					hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
						counts[0]++
						h.Set("X-Frozen", "original")
						headers = h
						return nil
					}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						counts[1]++
						borrowed = p.Body
						return ai.KeepPayload(), nil
					}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { counts[2]++; return nil }}
					res, err := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-pinned-"+r.id+entry+mode, req, hooks)
					checkMultimodalSuccess(t, ev, r, original, res, err)
					captures := u.provider.Requests()
					ev.Check("pinned replay unchanged", len(captures) == 2 && reflect.DeepEqual(captures[0], captures[1]) && captures[0].KeyAlias == tenantA.alias, "changed retry")
					ev.Check("pinned catalog price binding credential", res.Metadata.CatalogHash == hash && res.Metadata.BindingVersion == "b1" && res.Metadata.CredentialVersion == "v1" && res.Usage.Cost.Input < .001 && u.host.ReadsFor(res.Metadata.RequestID).Credentials == 1, "changed snapshot")
					if entry == "hooked" {
						ev.Check("pinned callbacks once", counts == [3]int{1, 1, 1}, "got %v", counts)
					}
					recordMultimodalCall(ev, u, original, res, err)
					next, e := multimodalCall(ctxFor(t), u, r, tenantA, entry, "req-denied-"+r.id+entry+mode, original, ai.Hooks{})
					phase := ai.PhaseCapability
					if mode == "binding" {
						phase = ai.PhaseBinding
					}
					ev.Check("new image call denied without fallback", errors.Is(e, &ai.Error{Code: ai.CodeTenantDenied, Phase: phase}) && len(next.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 2 && u.host.ReadsFor(next.Metadata.RequestID).Credentials == 0, "got %v", e)
					recordMultimodalCall(ev, u, original, next, e)
					otherReq := multimodalInput(t, r, tenantB)
					multimodalEnqueue(ev, u, r, tenantB, multimodalReply(t, ev, r, false))
					other, e := multimodalCall(ctxFor(t), u, r, tenantB, entry, "req-unaffected-"+r.id+entry+mode, otherReq, ai.Hooks{})
					checkMultimodalSuccess(t, ev, r, otherReq, other, e)
					recordMultimodalCall(ev, u, otherReq, other, e)
					u.released(ev)
				})
			}
		}
	}
}
