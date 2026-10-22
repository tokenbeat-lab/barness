package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"net/http"
	"strings"
	"testing"
)

func TestKimiResponsesCache(t *testing.T) {
	keys := map[ai.ProviderID]string{}
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, retention := range []ai.CacheRetention{"", ai.CacheRetentionShort, ai.CacheRetentionLong} {
			// Case labels must remain canonical evidence paths; the request
			// keeps the zero retention value to exercise protocol defaults.
			retentionLabel := string(retention)
			if retentionLabel == "" {
				retentionLabel = "default"
			}
			for _, session := range []string{"", "private-session"} {
				t.Run(string(market)+"/"+retentionLabel+"/"+session, func(t *testing.T) {
					ev := run.Case(t, kimiResponsesID(market, "K09-")+t.Name())
					w := newKimiResponsesWorld(t, market, nil)
					enqueue(ev, w, kimiResponsesReply(t))
					o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), ai.ResponsesOptions{CacheRetention: retention, SessionID: session}))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("cache preference accepted", o.err == nil, "%v", o.err) {
						return
					}
					fields := bodyFields(t, w.provider.Requests()[0].Body)
					ttl := "5m"
					if retention == ai.CacheRetentionLong {
						ttl = "1h"
					}
					ev.Check("protocol implicit TTL", jsonEqual(fields["prompt_cache_options"], json.RawMessage(`{"mode":"implicit","ttl":"`+ttl+`"}`)), "%s", w.provider.Requests()[0].Body)
					ev.Check("key only with session", (len(fields["prompt_cache_key"]) > 0) == (session != ""), "%s", w.provider.Requests()[0].Body)
					if session != "" {
						if prior := keys[market]; prior != "" {
							ev.Check("TTL does not change routing scope", prior == string(fields["prompt_cache_key"]), "routing scope changed")
						}
						keys[market] = string(fields["prompt_cache_key"])
					}
				})
			}
		}
	}
	if keys[ai.ProviderMoonshot] == keys[ai.ProviderMoonshotCN] {
		t.Fatal("actual Provider must isolate derived cache keys")
	}
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, hook := range []bool{false, true} {
			t.Run(string(market)+"/narrow-long/"+map[bool]string{true: "hook", false: "named"}[hook], func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K09-narrow-")+t.Name())
				w := newKimiResponsesWorld(t, market, func(c *ai.Catalog) {
					for i := range c.Models {
						c.Models[i].Compat.SupportsLongCacheRetention = ai.Value(false)
					}
				})
				opts := ai.ResponsesOptions{}
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
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), opts))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, phase)
				ev.Check("no network", len(w.provider.Requests()) == 0, "%v", o.err)
			})
		}
	}
}
func TestKimiResponsesSnapshots(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, kimiResponsesID(market, "K14-snapshot"))
			w := newKimiResponsesWorld(t, market, nil)
			b := w.host.Binding(tenantA.tenant, "kimi-responses")
			b.Retry.MaxRetries = 1
			w.host.PutBinding(b)
			enqueue(ev, w, provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":{"type":"internal_server_error"}}`)}})
			enqueue(ev, w, kimiResponsesReply(t))
			entered, resume := make(chan struct{}), make(chan struct{})
			var retained map[string]any
			headers, payloads, responses := 0, 0, 0
			hooks := ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, _ http.Header) error {
				headers++
				close(entered)
				select {
				case <-resume:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				payloads++
				retained = p.Body
				p.Body["max_output_tokens"] = 4096
				return ai.KeepPayload(), nil
			}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
				responses++
				retained["model"] = "after-serialization"
				return nil
			}}
			req := kimiResponsesRequest()
			params := map[string]json.RawMessage{"reasoning": json.RawMessage(`{"effort":"high"}`)}
			s := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("kimi-r-snapshot-"+string(market)), kimiResponsesTarget(), req, ai.ResponsesOptions{SamplingParams: params})
			defer s.Close()
			waitFor(ev, "header barrier entered", entered)
			req.Messages[0] = ai.UserText("caller-mutated")
			params["reasoning"][0] = '['
			b.Retry.MaxRetries = 0
			w.host.PutBinding(b)
			close(resume)
			o := drain(s)
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("frozen retry and hooks succeed", o.err == nil && headers == 1 && payloads == 1 && responses == 1 && len(w.provider.Requests()) == 2, "%v", o.err)
			if len(w.provider.Requests()) == 2 {
				a, b := w.provider.Requests()[0], w.provider.Requests()[1]
				ev.Check("identical pinned bytes", string(a.Body) == string(b.Body) && !strings.Contains(string(a.Body), "caller-mutated") && !strings.Contains(string(a.Body), "after-serialization"), "snapshot escaped")
			}
			if len(o.result.Metadata.Attempts) == 2 {
				o.result.Metadata.Attempts[1].CostEstimate.Status = ai.CostPartial
			}
			o.result.Message.Content = nil
			again, err := s.Result()
			ev.Record("second-result", again)
			ev.Check("Result owns content and estimate", err == nil && len(again.Message.Content) > 0 && again.Metadata.Attempts[1].CostEstimate.Status == ai.CostBounded, "%v", err)
		})
	}
}
func TestKimiResponsesHookAuthority(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, mode := range []string{"credential", "cache-key", "model", "native-reference"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K14-authority-")+mode)
				w := newKimiResponsesWorld(t, market, nil)
				hooks := ai.Hooks{}
				if mode == "credential" {
					hooks.TransformHeaders = func(_ context.Context, _ ai.CallScope, h http.Header) error {
						h.Set("Authorization", "Bearer synthetic-replacement")
						return nil
					}
				} else {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						switch mode {
						case "cache-key":
							p.Body["prompt_cache_key"] = "raw-session"
						case "model":
							p.Body["model"] = "unauthorized"
						case "native-reference":
							p.Body["input"] = []any{map[string]any{"type": "item_reference", "id": "foreign-id"}}
						}
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), nil))
				o.record(ev)
				checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
				ev.Check("no HTTP or Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%v", o.err)
			})
		}
	}
}
