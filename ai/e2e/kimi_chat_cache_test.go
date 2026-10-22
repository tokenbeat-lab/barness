package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestKimiChatCache(t *testing.T) {
	keys := map[ai.ProviderID]string{}
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, retention := range []ai.CacheRetention{"", ai.CacheRetentionShort, ai.CacheRetentionLong} {
			for _, session := range []string{"", "private-session"} {
				t.Run(string(market)+"/"+string(retention)+"/"+session, func(t *testing.T) {
					ev := run.Case(t, kimiChatID(market, "K09-")+t.Name())
					w := newKimiChatWorld(t, market, nil)
					enqueue(ev, w, kimiChatReply(t))
					o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), ai.ChatOptions{CacheRetention: retention, SessionID: session}))
					o.record(ev)
					if !ev.Check("cache preference accepted", o.err == nil, "%v", o.err) {
						return
					}
					fields := bodyFields(t, w.provider.Requests()[0].Body)
					ttl := "5m"
					if retention == ai.CacheRetentionLong {
						ttl = "1h"
					}
					ev.Check("precise implicit TTL", jsonEqual(fields["prompt_cache_options"], json.RawMessage(`{"mode":"implicit","ttl":"`+ttl+`"}`)), "%s", w.provider.Requests()[0].Body)
					ev.Check("key only for session", (len(fields["prompt_cache_key"]) > 0) == (session != ""), "%s", w.provider.Requests()[0].Body)
					if session != "" {
						if prior := keys[market]; prior != "" {
							ev.Check("retention does not change routing scope", prior == string(fields["prompt_cache_key"]), "different keys")
						}
						keys[market] = string(fields["prompt_cache_key"])
					}
				})
			}
		}
	}
	if keys["moonshot"] == keys["moonshot-cn"] {
		t.Fatal("actual provider must separate routing keys")
	}
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, hook := range []bool{false, true} {
			t.Run(string(market)+"/narrow-long/"+map[bool]string{true: "hook", false: "named"}[hook], func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K09-narrow-long-")+t.Name())
				w := newKimiChatWorld(t, market, func(c *ai.Catalog) {
					for i := range c.Models {
						c.Models[i].Compat.SupportsLongCacheRetention = ai.Value(false)
					}
				})
				opts := ai.ChatOptions{}
				hooks := ai.Hooks{}
				phase := ai.PhaseCapability
				if hook {
					phase = ai.PhaseRequest
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						p.Body["prompt_cache_options"] = map[string]any{"mode": "implicit", "ttl": "1h"}
						return ai.KeepPayload(), nil
					}
				} else {
					opts.CacheRetention = ai.CacheRetentionLong
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), opts))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, phase)
				ev.Check("no send", len(w.provider.Requests()) == 0, "%v", o.err)
			})
		}
	}
}
