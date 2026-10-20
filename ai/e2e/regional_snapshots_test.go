package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Failure modes: concurrent calls read global service/key state; a disabled
// binding changes an already pinned call; retry re-resolves credentials or
// reserializes mutable request bytes; event and Observer attribution diverge.
func TestRegionalProviderConcurrentFrozenRetries(t *testing.T) {
	for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
		t.Run(string(api), func(t *testing.T) {
			ev := run.Case(t, "RP13-frozen-"+string(api))
			var source *ai.Catalog
			w := newRegionalRuntimeWorld(t, func(c *ai.Config) { source = c.Catalog })
			hash := catalogHash(t, w.client.Catalog())
			type pending struct {
				tenant    string
				provider  ai.ProviderID
				id        string
				stream    *ai.Stream
				entered   chan struct{}
				request   ai.Request
				sampling  map[string]json.RawMessage
				envelopes []ai.EventEnvelope
				result    outcome
			}
			var calls []*pending
			resume := make(chan struct{})
			networkEntered, networkRelease := make(chan struct{}, 8), make(chan struct{})
			var releaseNetwork sync.Once
			defer releaseNetwork.Do(func() { close(networkRelease) })
			var release sync.Once
			defer release.Do(func() { close(resume) })
			for _, p := range regionalRuntimeProviders {
				for _, tenant := range []string{"regional-a", "regional-b"} {
					target := regionalRuntimeTarget(p, api)
					b := w.host.Binding(tenant, target.BindingID)
					b.Retry = ai.RetryPolicy{MaxRetries: 1}
					w.host.PutBinding(b)
					bad := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic temporary failure"}}`)}, Script: `{"error":{"message":"synthetic temporary failure"}}`}
					bad.OnReceive = func() { networkEntered <- struct{}{}; <-networkRelease }
					marker := "Synthetic response " + tenant + " " + string(p)
					reply := regionalRuntimeMarkedReply(t, api, marker)
					w.provider.EnqueueAt(regionalRuntimePrefix(tenant, p, api), bad, reply)
					id := fmt.Sprintf("regional-frozen-%s-%s-%s", api, p, tenant)
					ev.Record(id+"-script", []provider.Reply{bad, reply})
					q := &pending{tenant: tenant, provider: p, id: id, entered: make(chan struct{}), request: ai.Request{Messages: []ai.Message{ai.UserText("Synthetic frozen " + tenant + " " + string(p))}}, sampling: map[string]json.RawMessage{"temperature": json.RawMessage(`0.25`)}}
					var retained map[string]any
					hooks := ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, h http.Header) error {
						ev.Check("hook hides credentials", h.Get("Authorization") == "[REDACTED]" || h.Get("X-Api-Key") == "[REDACTED]", "missing redaction")
						close(q.entered)
						select {
						case <-resume:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}, OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
						ev.Check("payload identifies pinned service", payload.ProviderID == p && payload.API == api && payload.ModelID == mixedModel, "%+v", payload)
						retained = payload.Body
						return ai.KeepPayload(), nil
					}, OnResponse: func(_ context.Context, _ ai.CallScope, info ai.ResponseInfo) error {
						ev.Check("response metadata identifies pinned service", info.ProviderID == p && info.API == api && info.ModelID == mixedModel, "%+v", info)
						retained["model"] = "after-serialization"
						return nil
					}}
					q.stream = w.client.WithHooks(hooks).StreamSimple(ctxFor(t), ai.CallScope{TenantID: tenant, RequestID: id}, target, q.request, ai.SimpleOptions{SessionID: "same-session", SamplingParams: q.sampling})
					t.Cleanup(func() { _ = q.stream.Close() })
					calls = append(calls, q)
				}
			}
			for _, q := range calls {
				waitFor(ev, "snapshot pinned "+q.id, q.entered)
			}
			// Every call is now pinned, before admission. Change all Host records and
			// caller-owned inputs while no adapter is reading them.
			for i := range source.Models {
				source.Models[i].Cost.Input = 99
				source.Models[i].ID = "caller-mutated"
			}
			for _, q := range calls {
				b := w.host.Binding(q.tenant, regionalRuntimeTarget(q.provider, api).BindingID)
				b.Version, b.Enabled, b.Endpoint, b.CredentialRef = "b2", false, w.provider.URL()+"/revoked-target", b.CredentialRef+"-rotated"
				b.Retry = ai.RetryPolicy{}
				w.host.PutBinding(b)
				w.host.PutCredential(ai.Credential{OwnerTenantID: q.tenant, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "v2", Active: true, APIKey: ai.NewSecret(tenantA2.secret)})
				q.request.Messages[0] = ai.UserText("caller-mutated")
				q.sampling["temperature"][0] = '9'
				id := q.id + "-fresh"
				fresh := drain(w.client.Stream(ctxFor(t), ai.CallScope{TenantID: q.tenant, RequestID: id}, regionalRuntimeTarget(q.provider, api), ai.Request{Messages: []ai.Message{ai.UserText("fresh revoked call")}}, nil))
				ev.Record(id+"-result", fresh.result)
				checkRejected(ev, fresh, ai.CodeTenantDenied, ai.PhaseBinding)
				ev.Check("fresh revoked call reads no credential", w.host.ReadsFor(id).Credentials == 0 && len(fresh.result.Metadata.Attempts) == 0, "%+v", fresh.result.Metadata)
			}
			ev.Check("no requests before releasing pinned calls", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
			release.Do(func() { close(resume) })
			for range calls {
				waitFor(ev, "all eight initial attempts overlap at Provider", networkEntered)
			}
			ev.Check("all tenants/services concurrently hold distinct attempts", w.gauges.Permits() == 8 && w.adm.Held() == 8, "local=%d host=%d", w.gauges.Permits(), w.adm.Held())
			releaseNetwork.Do(func() { close(networkRelease) })
			var wg sync.WaitGroup
			for _, q := range calls {
				wg.Go(func() {
					q.envelopes, q.result.result, q.result.err = drainEnvelopes(ev, q.stream)
					q.result.streamed = true
					for _, env := range q.envelopes {
						q.result.events = append(q.result.events, env.Event)
					}
				})
			}
			wg.Wait()
			keys := map[string]bool{}
			for _, q := range calls {
				o := q.result
				regionalRuntimeRecord(ev, w, o)
				ev.Record(q.id+"-envelopes", envelopesJSON(q.envelopes))
				m := o.result.Metadata
				ev.Check("response belongs to its own tenant/service", textOf(o.result.Message) == "Synthetic response "+q.tenant+" "+string(q.provider), "got %q", textOf(o.result.Message))
				ev.Check("pinned snapshot succeeds despite Host changes", o.err == nil && m.ProviderID == q.provider && m.API == api && m.TenantID == q.tenant && m.AccountScopeID == "same-account" && m.BindingVersion == "b1" && m.CredentialVersion == "v1" && m.CatalogHash == hash && len(m.Attempts) == 2, "%v %+v", o.err, m)
				for _, env := range q.envelopes {
					ev.Check("every event keeps actual attribution", env.Call == m.CallAttribution, "got %+v want %+v", env.Call, m.CallAttribution)
				}
				miniMaxSharedOneTerminal(ev, o)
				regionalRuntimeAdmitted(ev, w, q.tenant, q.provider, q.id, 2)
				ev.Check("resolvers called once per logical call", w.host.ReadsFor(q.id).Bindings == 1 && w.host.ReadsFor(q.id).Credentials == 1, "%+v", w.host.ReadsFor(q.id))
				var captures []provider.Request
				for _, r := range w.provider.Requests() {
					if strings.HasPrefix(r.Path, regionalRuntimePrefix(q.tenant, q.provider, api)+"/") {
						captures = append(captures, r)
					}
				}
				ev.Check("exactly two attempts at original endpoint", len(captures) == 2, "got %d", len(captures))
				if len(captures) == 2 {
					a, b := captures[0], captures[1]
					ev.Check("retry reuses original credentials and immutable bytes", a.KeyAlias == regionalRuntimeKey(q.tenant, q.provider).alias && b.KeyAlias == a.KeyAlias && a.Path == b.Path && reflect.DeepEqual(a.Body, b.Body) && (api == ai.APIAnthropicMessages || strings.Contains(string(a.Body), `"temperature":0.25`)) && strings.Contains(string(a.Body), "Synthetic frozen "+q.tenant+" "+string(q.provider)) && !strings.Contains(string(a.Body), "caller-mutated") && !strings.Contains(string(a.Body), "after-serialization"), "request snapshot changed")
					if api == ai.APIOpenAIResponses {
						key := regionalRuntimeBodyKey(t, a)
						ev.Check("eight tenant/provider cache scopes stay separate", key != "" && !keys[key], "colliding key %q", key)
						keys[key] = true
					}
				}
			}
			ev.Check("only bounded original attempts sent", len(w.provider.Requests()) == 16, "got %d", len(w.provider.Requests()))
			wire := observationText(t, w.rec.all())
			for _, private := range []string{"Synthetic frozen", "Synthetic response", "caller-mutated", "after-serialization"} {
				ev.Check("Observer contains no private body", !strings.Contains(wire, private), "private body in Observer")
			}
			w.released(ev)
		})
	}
}
