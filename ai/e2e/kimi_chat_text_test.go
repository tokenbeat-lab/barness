package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"net/http"
	"slices"
	"testing"
)

func TestKimiChatText(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, hooked := range []bool{false, true} {
				t.Run(string(market)+"/"+entry+"/"+map[bool]string{true: "hooks", false: "plain"}[hooked], func(t *testing.T) {
					ev := run.Case(t, kimiChatID(market, "K03-")+t.Name())
					w := newKimiChatWorld(t, market, nil)
					enqueue(ev, w, kimiChatReply(t))
					ev.Fixture("frames.json", mustMarshal(t, kimiChatFrames()))
					var order []string
					hooks := ai.Hooks{}
					if hooked {
						hooks = ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, _ http.Header) error {
							order = append(order, "headers")
							return nil
						}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							order = append(order, "payload")
							return ai.KeepPayload(), nil
						}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
							order = append(order, "response")
							return nil
						}}
					}
					client := w.client.WithHooks(hooks)
					scope := textScope(t.Name())
					req := kimiChatRequest()
					target := kimiChatTarget()
					var o outcome
					switch entry {
					case "complete":
						o.result, o.err = client.Complete(ctxFor(t), scope, target, req, ai.ChatOptions{SessionID: "private-session"})
					case "complete-simple":
						o.result, o.err = client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{SessionID: "private-session"})
					case "stream-simple":
						o = drain(client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{SessionID: "private-session"}))
					default:
						o = drain(client.Stream(ctxFor(t), scope, target, req, ai.ChatOptions{SessionID: "private-session"}))
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("one successful request", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
						return
					}
					r := w.provider.Requests()[0]
					f := bodyFields(t, r.Body)
					for _, key := range []string{"store", "developer", "thinking", "max_tokens", "prompt_cache_retention", "reasoning_split"} {
						ev.Check("omits "+key, len(f[key]) == 0, "%s", r.Body)
					}
					ev.Check("implicit five minute cache", jsonEqual(f["prompt_cache_options"], json.RawMessage(`{"mode":"implicit","ttl":"5m"}`)), "%s", r.Body)
					ev.Check("private derived cache key", len(f["prompt_cache_key"]) > 0 && string(f["prompt_cache_key"]) != `"private-session"`, "%s", r.Body)
					ev.Check("stream usage enabled", jsonEqual(f["stream_options"], json.RawMessage(`{"include_usage":true}`)), "%s", r.Body)
					var msgs []map[string]any
					mustUnmarshal(t, f["messages"], &msgs)
					ev.Check("system message and explicit auth", msgs[0]["role"] == "system" && r.Path == "/v1/chat/completions" && r.KeyAlias == kimiChatKey.alias, "%+v", r)
					if entry == "complete-simple" || entry == "stream-simple" {
						ev.Check("simple uses supported default", string(f["max_completion_tokens"]) == "131072", "%s", r.Body)
					} else {
						ev.Check("full preserves server output default", len(f["max_completion_tokens"]) == 0, "%s", r.Body)
					}
					ev.Check("default max effort retained", len(f["reasoning_effort"]) == 0, "%s", r.Body)
					ev.Check("reasoning text and trailing usage", len(o.result.Message.Content) == 2 && o.result.Message.Usage.Input == 100 && o.result.Message.Usage.CacheWrite == 80 && o.result.Message.Usage.Output == 50, "%+v", o.result.Message)
					if hooked {
						ev.Check("callbacks run in order", slices.Equal(order, []string{"headers", "payload", "response"}), "%v", order)
					}
				})
			}
		}
	}
}
