package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
)

func TestGLMResponsesCacheRouting(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM05-cache-"+model)
			w := newGLMResponsesWorld(t, nil)
			b := glm.ResponsesBinding(tenantB.tenant, "same-account-alias", "cred-glm-responses", w.provider.URL()+"/api/v1")
			w.host.PutBinding(b)
			w.host.PutCredential(ai.Credential{OwnerTenantID: tenantB.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantB.secret)})
			scenarios, raw := glmResponsesFixtures(t)
			ev.Fixture("text.json", raw)
			var keys []string
			for _, sc := range []struct {
				id, tenant, account, session string
				retention                    ai.CacheRetention
			}{
				{"short", tenantA.tenant, "same-account-alias", "same-session", ai.CacheRetentionShort},
				{"long", tenantA.tenant, "same-account-alias", "same-session", ai.CacheRetentionLong},
				{"default", tenantA.tenant, "same-account-alias", "same-session", ""},
				{"tenant-isolation", tenantB.tenant, "same-account-alias", "same-session", ai.CacheRetentionShort},
				{"account-isolation", tenantA.tenant, "other-account-alias", "same-session", ai.CacheRetentionShort},
				{"none", tenantA.tenant, "other-account-alias", "same-session", ai.CacheRetentionNone},
				{"no-session", tenantA.tenant, "other-account-alias", "", ai.CacheRetentionLong},
			} {
				if sc.tenant == tenantA.tenant {
					w.updateBindingOf(tenantA, "glm-responses", func(b *ai.Binding) { b.AccountScopeID = sc.account })
					w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: "cred-glm-responses", Version: "v1", AccountScopeID: sc.account, Active: true, APIKey: ai.NewSecret(glmA.secret)})
				}
				enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
				scope := textScope("glm-cache-" + model + sc.id)
				scope.TenantID = sc.tenant
				o := drain(w.client.Stream(ctxFor(t), scope, glmResponsesTarget(model), glmResponsesRequest(), ai.ResponsesOptions{SessionID: sc.session, CacheRetention: sc.retention}))
				ev.Record(sc.id+"-result", o.result)
				ev.Record(sc.id+"-error", errString(o.err))
				ev.Check(sc.id+" succeeds", o.err == nil, "%v", o.err)
				reqs := w.provider.Requests()
				if len(reqs) == 0 {
					continue
				}
				req := reqs[len(reqs)-1]
				body := bodyFields(t, req.Body)
				var key string
				_ = json.Unmarshal(body["prompt_cache_key"], &key)
				keys = append(keys, key)
				want := sc.session != "" && sc.retention != ai.CacheRetentionNone
				ev.Check(sc.id+" key presence", (key != "") == want, "%s", req.Body)
				if want {
					ev.Check(sc.id+" hides raw session/tenant/account", !strings.Contains(key, sc.session) && !strings.Contains(key, sc.tenant) && !strings.Contains(key, sc.account), "%s", key)
				}
				for _, field := range []string{"prompt_cache_retention", "prompt_cache_options"} {
					_, has := body[field]
					ev.Check(sc.id+" no "+field, !has, "%s", req.Body)
				}
				for _, header := range []string{"Session_id", "X-Client-Request-Id"} {
					ev.Check(sc.id+" no affinity "+header, len(req.Header[header]) == 0, "%v", req.Header)
				}
			}
			ev.Record("requests", w.provider.Requests())
			if len(keys) == 7 {
				ev.Check("short long default share routing only", keys[0] == keys[1] && keys[0] == keys[2], "%v", keys)
				ev.Check("same tenant/account aliases never cross scope", keys[0] != keys[3] && keys[0] != keys[4] && keys[3] != keys[4], "%v", keys)
			}
		})
	}
}

func TestGLMResponsesCacheKeyAuthorization(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"model", "hook"} {
			t.Run(model+"/"+source, func(t *testing.T) {
				ev := run.Case(t, "GLM05-cache-key-auth-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				if source == "model" {
					cat := glm.ResponsesCatalog()
					for i := range cat.Models {
						cat.Models[i].SamplingParams = map[string]json.RawMessage{"prompt_cache_key": json.RawMessage(`"self-reported-session"`)}
					}
					_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
					ev.Record("configuration-error", errString(err))
					ev.Check("catalog cannot assign cache identity", err != nil && strings.Contains(err.Error(), "prompt_cache_key"), "%v", err)
					ev.Check("configuration refusal sends no request", len(w.provider.Requests()) == 0, "%v", w.provider.Requests())
					return
				}
				hooks := ai.Hooks{}
				if source == "hook" {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						p.Body["prompt_cache_key"] = "self-reported-session"
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-cache-auth"), glmResponsesTarget(model), glmResponsesRequest(), ai.ResponsesOptions{SessionID: "same-session"}))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
				ev.Check("zero requests and Attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}
