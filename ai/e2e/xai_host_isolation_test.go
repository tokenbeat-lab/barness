package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIHostConcurrentIsolation(t *testing.T) {
	ev := run.Case(t, "XAI-E13-E16-same-label-concurrent-isolation")
	rec := newRecorder()
	w := newXAIRuntimeWorld(t, func(c *ai.Catalog) {
		m, ok := ai.BuiltinCatalog().Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4.1-mini")
		if !ok {
			t.Fatal("missing old Responses model")
		}
		c.Models = append(c.Models, m)
	}, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
	for _, k := range []tenantKey{xaiKey, tenantB} {
		b := xaiBinding()
		b.TenantID, b.AccountScopeID = k.tenant, "same-account-label"
		w.host.PutBinding(b)
		w.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(k.secret)})
	}
	legacy := primaryBinding(tenantA, w.provider.URL())
	legacy.AccountScopeID = "same-account-label"
	w.host.PutBinding(legacy)
	cred := primaryCredential(tenantA, "v1")
	cred.AccountScopeID = legacy.AccountScopeID
	w.host.PutCredential(cred)
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("old-responses.json", raw)
	type call struct {
		tenant, id, marker, alias string
		target                    ai.Target
		provider                  ai.ProviderID
	}
	calls := []call{
		{tenantA.tenant, "xai-isolated-a", "xai-a-visible", xaiKey.alias, xaiTarget(), ai.ProviderXAI},
		{tenantB.tenant, "xai-isolated-b", "xai-b-visible", tenantB.alias, xaiTarget(), ai.ProviderXAI},
		{tenantA.tenant, "old-isolated-a", f.Expect.Text, tenantA.alias, textTarget(f), ai.ProviderOpenAI},
	}
	resume := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(resume) }) }
	defer unblock()
	type completedCall struct {
		outcome      outcome
		attributions []ai.CallAttribution
	}
	dones := make([]chan completedCall, len(calls))
	for i, c := range calls {
		ready := make(chan struct{})
		reply := xaiReply(t, xaiTerminal(strings.ReplaceAll(xaiMessage, "Hello Grok.", c.marker), xaiUsage))
		req := ai.Request{Messages: []ai.Message{ai.UserText(c.id)}}
		if c.provider == ai.ProviderOpenAI {
			reply = sseEvents(t, f.Events, "")
			req = textRequest(f)
		}
		enqueue(ev, w.world, reply)
		h := ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
			close(ready)
			select {
			case <-resume:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}}
		dones[i] = make(chan completedCall, 1)
		go func() {
			s := w.client.WithHooks(h).Stream(ctxFor(t), ai.CallScope{TenantID: c.tenant, RequestID: c.id}, c.target, req, ai.ResponsesOptions{SessionID: "same-private-session"})
			defer s.Close()
			var o outcome
			o.streamed = true
			var attributions []ai.CallAttribution
			for s.Next() {
				o.events = append(o.events, s.Event())
				attributions = append(attributions, s.Envelope().Call)
			}
			o.result, o.err = s.Result()
			o.streamErr = s.Err()
			dones[i] <- completedCall{outcome: o, attributions: attributions}
		}()
		waitFor(ev, "call holds its own response and permit "+c.id, ready)
	}
	ev.Record("overlap", map[string]int64{"bodies": w.bodies.Open(), "permits": w.gauges.Permits(), "host-permits": int64(w.adm.Held())})
	ev.Check("all three routes overlap", w.bodies.Open() == 3 && w.gauges.Permits() == 3 && w.adm.Held() == 3, "%d %d %d", w.bodies.Open(), w.gauges.Permits(), w.adm.Held())
	unblock()
	for i, c := range calls {
		completed := waitValue(ev, "isolated call settles "+c.id, dones[i])
		o := completed.outcome
		ev.Record(c.id+"-result", o.result)
		ev.Record(c.id+"-events", eventsJSON(o.events))
		ev.Record(c.id+"-event-attributions", completed.attributions)
		for _, attribution := range completed.attributions {
			ev.Check("event retains immutable attribution "+c.id, attribution == o.result.Metadata.CallAttribution, "%+v", attribution)
		}
		ev.Check("result retains attribution "+c.id, o.err == nil && textOf(o.result.Message) == c.marker && o.result.Metadata.TenantID == c.tenant && o.result.Metadata.RequestID == c.id && o.result.Metadata.ProviderID == c.provider, "%v %+v", o.err, o.result)
		obs := rec.awaitCall(ev, c.id)
		checkCallRecords(ev, obs, ai.CallScope{TenantID: c.tenant, RequestID: c.id}, c.target.BindingID, o, 1)
	}
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	keys := map[string]bool{}
	for i, q := range reqs {
		if i >= len(calls) {
			break
		}
		var key string
		_ = json.Unmarshal(bodyFields(t, q.Body)["prompt_cache_key"], &key)
		ev.Check("bound bearer remains isolated "+calls[i].id, q.KeyAlias == calls[i].alias, "%s", q.KeyAlias)
		ev.Check("cache scope is distinct and hides session "+calls[i].id, key != "" && !keys[key] && !strings.Contains(key, "same-private-session"), "%s", key)
		keys[key] = true
	}
	ev.Check("exactly three independently attributed requests", len(reqs) == 3, "%d", len(reqs))
	w.released(ev)
}
