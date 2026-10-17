package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxChatSamplingLayers(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, simple := range []bool{false, true} {
			t.Run(model+map[bool]string{false: "/full", true: "/simple"}[simple], func(t *testing.T) {
				ev := run.Case(t, "MM02-sampling-"+t.Name())
				w := newMiniMaxChatWorld(t, func(cat *ai.Catalog) {
					for i := range cat.Models {
						m := &cat.Models[i]
						if m.ID == model {
							m.SamplingParams = map[string]json.RawMessage{"temperature": json.RawMessage(`0.1`), "top_p": json.RawMessage(`0.9`)}
							m.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{m.DefaultThinkingLevel: {"temperature": json.RawMessage(`0.2`), "top_p": json.RawMessage(`0.8`)}, ai.ThinkingLow: {"top_p": json.RawMessage(`0.1`)}}
						}
					}
				})
				enqueue(ev, w, miniMaxChatReply(t, model))
				client := w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					ev.Check("named then general/profile/call then hook", p.Body["temperature"] == json.Number("0.3") && p.Body["top_p"] == json.Number("0.8") && p.Body["thinking"] == nil && p.Body["reasoning_effort"] == nil, "%v", p.Body)
					p.Body["temperature"] = json.Number("1e-100")
					if model == "MiniMax-M3" {
						p.Body["thinking"] = map[string]any{"type": "disabled"}
					} else {
						p.Body["reasoning_effort"] = "low"
					}
					return ai.KeepPayload(), nil
				}})
				params := map[string]json.RawMessage{"temperature": json.RawMessage(`0.3`)}
				var o outcome
				if simple {
					o = drain(client.StreamSimple(ctxFor(t), textScope(t.Name()), miniMaxChatTarget(model), miniMaxChatRequest(), ai.SimpleOptions{MaxTokens: ai.Value(1), Temperature: ai.Value(0.7), SamplingParams: params}))
				} else {
					o = drain(client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget(model), miniMaxChatRequest(), ai.ChatOptions{MaxTokens: ai.Value(1), Temperature: ai.Value(0.7), SamplingParams: params}))
				}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Record("candidate", w.client.Catalog())
				if !ev.Check("small budget and valid small sampling accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
					return
				}
				f := bodyFields(t, w.provider.Requests()[0].Body)
				ev.Check("one-token budget and profile unchanged after wire rewrite", string(f["max_completion_tokens"]) == "1" && string(f["temperature"]) == "1e-100" && string(f["top_p"]) == "0.8", "%s", w.provider.Requests()[0].Body)
			})
		}
	}
}
