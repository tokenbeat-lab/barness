package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAICacheScope(t *testing.T) {
	ev := run.Case(t, "XAI-E13-cache-scope")
	w := newXAIWorld(t, nil)
	type scenario struct {
		name, tenant, account, session string
		retention                      ai.CacheRetention
	}
	scenarios := []scenario{{"short", tenantA.tenant, "same-account", "session-sensitive", ai.CacheRetentionShort}, {"long", tenantA.tenant, "same-account", "session-sensitive", ai.CacheRetentionLong}, {"default", tenantA.tenant, "same-account", "session-sensitive", ""}, {"tenant", tenantB.tenant, "same-account", "session-sensitive", ""}, {"account", tenantA.tenant, "other-account", "session-sensitive", ""}, {"none", tenantA.tenant, "same-account", "session-sensitive", ai.CacheRetentionNone}, {"no-session", tenantA.tenant, "same-account", "", ai.CacheRetentionLong}}
	var keys []string
	for _, sc := range scenarios {
		b := xaiBinding()
		b.TenantID = sc.tenant
		b.AccountScopeID = sc.account
		w.host.PutBinding(b)
		secret := xaiKey.secret
		if sc.tenant == tenantB.tenant {
			secret = tenantB.secret
		}
		w.host.PutCredential(ai.Credential{OwnerTenantID: sc.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: sc.account, Active: true, APIKey: ai.NewSecret(secret)})
		enqueue(ev, w, xaiReply(t))
		scope := textScope(sc.name)
		scope.TenantID = sc.tenant
		res, err := w.client.Complete(ctxFor(t), scope, xaiTarget(), xaiRequest(), ai.ResponsesOptions{SessionID: sc.session, CacheRetention: sc.retention})
		ev.Record(sc.name+"-result", res)
		if !ev.Check(sc.name+" succeeds", err == nil && len(w.provider.Requests()) == len(keys)+1, "%v", err) {
			return
		}
		req := w.provider.Requests()[len(keys)]
		fields := bodyFields(t, req.Body)
		var key string
		_ = json.Unmarshal(fields["prompt_cache_key"], &key)
		keys = append(keys, key)
		want := sc.session != "" && sc.retention != ai.CacheRetentionNone
		ev.Check(sc.name+" key presence", (key != "") == want, "key=%s", key)
		if want {
			ev.Check(sc.name+" hides identifiers", !strings.Contains(key, sc.session) && !strings.Contains(key, sc.tenant) && !strings.Contains(key, sc.account), "%s", key)
		}
		ev.Check(sc.name+" no TTL or affinity", fields["prompt_cache_retention"] == nil && fields["prompt_cache_options"] == nil && len(req.Header["Session_id"]) == 0 && len(req.Header["X-Client-Request-Id"]) == 0, "%s %+v", req.Body, req.Header)
	}
	ev.Record("requests", w.provider.Requests())
	ev.Check("default short long preserve one scope", keys[0] == keys[1] && keys[0] == keys[2], "%v", keys)
	ev.Check("tenant and account scopes differ", keys[0] != keys[3] && keys[0] != keys[4] && keys[3] != keys[4], "%v", keys)
}

func TestXAIHeadersAndHooksOrder(t *testing.T) {
	ev := run.Case(t, "XAI-E02-E19-hooks-order")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	var order []string
	h := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, header http.Header) error {
		order = append(order, "headers")
		header.Set("X-Synthetic-Trace", "hook")
		return nil
	}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		order = append(order, "payload")
		p.Body["temperature"] = 0.5
		return ai.KeepPayload(), nil
	}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
		order = append(order, "response")
		return nil
	}}
	o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	ev.Record("hook-order", order)
	ev.Check("hooks complete in documented order", o.err == nil && strings.Join(order, ",") == "headers,payload,response", "err=%v order=%v", o.err, order)
	if len(w.provider.Requests()) == 1 {
		r := w.provider.Requests()[0]
		ev.Check("one bound bearer and independent trace", r.KeyAlias == xaiKey.alias && r.Header["X-Synthetic-Trace"] == "hook", "%+v", r.Header)
	}
}

func TestXAICacheProviderIsolation(t *testing.T) {
	ev := run.Case(t, "XAI-E13-cache-provider-isolation")
	w := newXAIWorld(t, func(c *ai.Catalog) {
		m, ok := ai.BuiltinCatalog().Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4.1-mini")
		if !ok {
			t.Fatal("missing builtin comparison model")
		}
		c.Models = append(c.Models, m)
	})
	primary := primaryBinding(tenantA, w.provider.URL())
	primary.AccountScopeID = "acct-xai"
	w.host.PutBinding(primary)
	cred := primaryCredential(tenantA, "v1")
	cred.AccountScopeID = "acct-xai"
	w.host.PutCredential(cred)
	var keys []string
	for _, target := range []ai.Target{xaiTarget(), {BindingID: "primary", ModelID: "gpt-4.1-mini"}} {
		enqueue(ev, w, xaiReply(t))
		res, err := w.client.Complete(ctxFor(t), textScope(t.Name()+target.BindingID), target, xaiRequest(), ai.ResponsesOptions{SessionID: "same-session"})
		ev.Record(target.BindingID+"-result", res)
		if !ev.Check(target.BindingID+" succeeds", err == nil && len(w.provider.Requests()) == len(keys)+1, "%v", err) {
			return
		}
		var key string
		mustUnmarshal(t, bodyFields(t, w.provider.Requests()[len(keys)].Body)["prompt_cache_key"], &key)
		keys = append(keys, key)
	}
	ev.Record("requests", w.provider.Requests())
	ev.Check("same tenant/account/session differ by Provider", keys[0] != "" && keys[1] != "" && keys[0] != keys[1], "%v", keys)
}
