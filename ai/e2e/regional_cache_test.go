package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"net/http"
	"os"
	"testing"
)

// Failure modes (before implementation): v1 lacks Provider; fields may use
// delimiters or rune lengths, normalize Unicode/case, rotate with keys, or
// accidentally derive headers for a provider that only supports routing.
func TestRegionalProviderCacheV2(t *testing.T) {
	vectors, raw := regionalRuntimeCacheVectors(t)
	for _, v := range vectors {
		if v.Provider == "deepseek" {
			continue
		}
		t.Run(v.ID, func(t *testing.T) {
			ev := run.Case(t, "RP11-cache-v2-"+v.ID)
			ev.Fixture("independent-vectors.json", raw)
			w := newRegionalRuntimeWorld(t, nil)
			p := ai.ProviderID(v.Provider)
			api := ai.APIOpenAIResponses
			regionalRuntimeInstall(w, v.Tenant, p, api, v.Account, tenantA)
			reply := regionalRuntimeReply(t, api)
			ev.Record("response-script", reply)
			w.provider.EnqueueAt(regionalRuntimePrefix(v.Tenant, p, api), reply)
			o := drain(w.client.Stream(ctxFor(t), ai.CallScope{TenantID: v.Tenant, RequestID: "regional-cache-" + v.ID}, regionalRuntimeTarget(p, api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic cache input.")}}, ai.ResponsesOptions{SessionID: v.Session}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("call succeeds", o.err == nil, "%v", o.err)
			captures := w.provider.Requests()
			ev.Check("one request", len(captures) == 1, "got %d", len(captures))
			if len(captures) == 1 {
				r := captures[0]
				ev.Check("exact versioned UTF8 length-prefixed cache key", regionalRuntimeBodyKey(t, r) == v.Key, "got %q want %q", regionalRuntimeBodyKey(t, r), v.Key)
				for _, name := range []string{"Session_id", "X-Client-Request-Id"} {
					want := ""
					if p == ai.ProviderOpenAI {
						want = v.Key
					}
					ev.Check("affinity capability "+name, r.Header[name] == want, "got %q want %q", r.Header[name], want)
				}
			}
			w.released(ev)
		})
	}
}

type regionalRuntimeCacheVector struct{ ID, Tenant, Provider, Account, Session, Key string }

func regionalRuntimeCacheVectors(t *testing.T) ([]regionalRuntimeCacheVector, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/regional-runtime/cache-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct{ Vectors []regionalRuntimeCacheVector }
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f.Vectors, raw
}

func TestRegionalProviderCacheEligibilityAndRotation(t *testing.T) {
	vectors, raw := regionalRuntimeCacheVectors(t)
	expected := map[ai.ProviderID]string{}
	for _, v := range vectors {
		if v.Tenant == "regional-a" && v.Account == "same-account" && v.Session == "same-session" {
			expected[ai.ProviderID(v.Provider)] = v.Key
		}
	}
	for _, p := range append(append([]ai.ProviderID{}, regionalRuntimeProviders...), ai.ProviderOpenAI, ai.ProviderDeepSeek) {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses} {
			t.Run(string(p)+"/"+string(api), func(t *testing.T) {
				ev := run.Case(t, "RP11-cache-eligibility-"+string(p)+"-"+string(api))
				ev.Fixture("independent-vectors.json", raw)
				w := newRegionalRuntimeWorld(t, nil)
				tenant := "regional-a"
				regionalRuntimeInstall(w, tenant, p, api, "same-account", tenantA)
				eligible := api == ai.APIOpenAIResponses && p != ai.ProviderDeepSeek || api == ai.APIOpenAICompletions && (p == ai.ProviderOpenAI || p == ai.ProviderDeepSeek)
				for _, step := range []struct {
					name, session string
					retention     ai.CacheRetention
					rotate        bool
				}{
					{"long", "same-session", ai.CacheRetentionLong, false},
					{"key-rotation", "same-session", ai.CacheRetentionLong, true},
					{"none", "same-session", ai.CacheRetentionNone, false},
					{"no-session", "", ai.CacheRetentionLong, false},
				} {
					if step.rotate {
						b := w.host.Binding(tenant, regionalRuntimeTarget(p, api).BindingID)
						b.CredentialRef += "-rotated"
						b.Version = "b2"
						w.host.PutBinding(b)
						w.host.PutCredential(ai.Credential{OwnerTenantID: tenant, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "v2", Active: true, APIKey: ai.NewSecret(tenantA2.secret)})
					}
					var opts ai.Options = ai.ChatOptions{SessionID: step.session, CacheRetention: step.retention}
					if api == ai.APIOpenAIResponses {
						opts = ai.ResponsesOptions{SessionID: step.session, CacheRetention: step.retention}
					}
					reply := regionalRuntimeReply(t, api)
					ev.Record(step.name+"-script", reply)
					w.provider.EnqueueAt(regionalRuntimePrefix(tenant, p, api), reply)
					id := "regional-cache-eligibility-" + string(p) + "-" + string(api) + "-" + step.name
					o := drain(w.client.Stream(ctxFor(t), ai.CallScope{TenantID: tenant, RequestID: id}, regionalRuntimeTarget(p, api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic cache rotation input.")}}, opts))
					ev.Record(step.name+"-result", o.result)
					ev.Check("cache option call succeeds", o.err == nil, "%v", o.err)
					requests := w.provider.Requests()
					if len(requests) == 0 {
						t.Fatal("missing request")
					}
					r := requests[len(requests)-1]
					want := ""
					if eligible && step.session != "" && step.retention != ai.CacheRetentionNone {
						want = expected[p]
					}
					ev.Check("exact v2 key or protocol suppression", regionalRuntimeBodyKey(t, r) == want, "got %q want %q", regionalRuntimeBodyKey(t, r), want)
					for _, name := range []string{"Session_id", "X-Client-Request-Id"} {
						affinity := ""
						if api == ai.APIOpenAIResponses && p == ai.ProviderOpenAI {
							affinity = want
						}
						ev.Check("affinity matches eligibility", r.Header[name] == affinity, "%s=%q want %q", name, r.Header[name], affinity)
					}
					if step.rotate {
						ev.Check("new credential selected without cache change", o.result.Metadata.CredentialVersion == "v2" && r.KeyAlias == tenantA2.alias, "%+v", o.result.Metadata)
					}
				}
				// A trusted callback still cannot replace this call's derived identity.
				if eligible {
					hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
						payload.Body["prompt_cache_key"] = "hook-assigned-key"
						return ai.KeepPayload(), nil
					}}
					var opts ai.Options = ai.ChatOptions{SessionID: "same-session", CacheRetention: ai.CacheRetentionLong}
					if api == ai.APIOpenAIResponses {
						opts = ai.ResponsesOptions{SessionID: "same-session"}
					}
					n := len(w.provider.Requests())
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), ai.CallScope{TenantID: tenant, RequestID: "regional-cache-tamper-" + string(p) + "-" + string(api)}, regionalRuntimeTarget(p, api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic cache hook input.")}}, opts))
					o.record(ev)
					checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
					ev.Check("hook cannot send another key", len(w.provider.Requests()) == n && len(o.result.Metadata.Attempts) == 0, "sent unauthorized key")
				}
				if p == ai.ProviderOpenAI && api == ai.APIOpenAIResponses {
					hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
						h.Set("Session_id", "hook-assigned-key")
						return nil
					}}
					n := len(w.provider.Requests())
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), ai.CallScope{TenantID: tenant, RequestID: "regional-affinity-tamper"}, regionalRuntimeTarget(p, api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic cache header input.")}}, ai.ResponsesOptions{SessionID: "same-session"}))
					o.record(ev)
					checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
					ev.Check("hook cannot replace affinity", len(w.provider.Requests()) == n, "sent unauthorized affinity")
				}
				ev.Record("requests", w.provider.Requests())
				w.released(ev)
			})
		}
	}
}
