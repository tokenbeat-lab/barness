package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"net/http"
	"testing"
)

func TestKimiResponsesText(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, hooked := range []bool{false, true} {
				t.Run(string(market)+"/"+entry+"/"+map[bool]string{true: "hooks", false: "plain"}[hooked], func(t *testing.T) {
					ev := run.Case(t, kimiResponsesID(market, "K03-")+t.Name())
					w := newKimiResponsesWorld(t, market, nil)
					enqueue(ev, w, kimiResponsesReply(t))
					ev.Fixture("frames.json", mustMarshal(t, kimiResponsesFrames()))
					n := 0
					hooks := ai.Hooks{}
					if hooked {
						hooks = ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, _ http.Header) error { n++; return nil }, OnPayload: func(_ context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
							n++
							return ai.KeepPayload(), nil
						}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error { n++; return nil }}
					}
					client := w.client.WithHooks(hooks)
					scope := textScope("kimi-responses-text-" + string(market) + entry + map[bool]string{true: "hooks", false: "plain"}[hooked])
					req, target := kimiResponsesRequest(), kimiResponsesTarget()
					var o outcome
					switch entry {
					case "complete":
						o.result, o.err = client.Complete(ctxFor(t), scope, target, req, ai.ResponsesOptions{SessionID: "private-session"})
					case "complete-simple":
						o.result, o.err = client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{SessionID: "private-session"})
					case "stream-simple":
						o = drain(client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{SessionID: "private-session"}))
					default:
						o = drain(client.Stream(ctxFor(t), scope, target, req, ai.ResponsesOptions{SessionID: "private-session"}))
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("one successful request", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
						return
					}
					r := w.provider.Requests()[0]
					f := bodyFields(t, r.Body)
					for _, key := range []string{"store", "include", "reasoning", "service_tier", "prompt_cache_retention", "previous_response_id", "conversation", "background"} {
						ev.Check("omits "+key, len(f[key]) == 0, "%s", r.Body)
					}
					ev.Check("cache implicit5m", jsonEqual(f["prompt_cache_options"], json.RawMessage(`{"mode":"implicit","ttl":"5m"}`)), "%s", r.Body)
					ev.Check("cache key derived", len(f["prompt_cache_key"]) > 0 && string(f["prompt_cache_key"]) != `"private-session"`, "%s", r.Body)
					var input []map[string]any
					mustUnmarshal(t, f["input"], &input)
					ev.Check("system role and endpoint", input[0]["role"] == "system" && r.Path == "/v1/responses" && r.KeyAlias == kimiResponsesKey.alias, "%+v", r)
					if entry == "stream-simple" || entry == "complete-simple" {
						ev.Check("simple default cap", string(f["max_output_tokens"]) == "131072", "%s", r.Body)
					} else {
						ev.Check("full default left to service", len(f["max_output_tokens"]) == 0, "%s", r.Body)
					}
					ev.Check("summary and text preserved", len(o.result.Message.Content) == 2 && o.result.Message.Content[0].(ai.Thinking).Thinking == "Check.\n\nThen answer." && o.result.Message.Content[1].(ai.Text).Text == "Hello Kimi.", "%+v", o.result.Message)
					ev.Check("null encryption retains actual reasoning item", o.result.Message.Content[0].(ai.Thinking).Signature == kimiResponsesReasoning, "%+v", o.result.Message.Content[0])
					ev.Check("usage normalized at terminal", o.result.Message.Usage.Input == 100 && o.result.Message.Usage.CacheWrite == 80 && o.result.Message.Usage.Output == 50 && o.result.Metadata.Attempts[0].CostEstimate.Status == ai.CostBounded, "%+v", o.result)
					if hooked {
						ev.Check("all hooks run", n == 3, "%d", n)
					}
				})
			}
		}
	}
}
