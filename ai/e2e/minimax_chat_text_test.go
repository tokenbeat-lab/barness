package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxChatText(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			t.Run(model+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "MM02-text-"+model+"-"+entry)
				scope := textScope("mm-chat-text-" + model + "-" + entry)
				rec := newRecorder()
				w := newMiniMaxChatWorld(t, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
				frames := miniMaxChatFrames(model)
				ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
				enqueue(ev, w, chatEvents(t, frames, ""))
				var order []string
				client := w.client.WithHooks(ai.Hooks{
					TransformHeaders: func(_ context.Context, _ ai.CallScope, _ http.Header) error {
						order = append(order, "header")
						return nil
					},
					OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						order = append(order, "payload")
						return ai.KeepPayload(), nil
					},
					OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
						order = append(order, "response")
						return nil
					},
				})
				var o outcome
				switch entry {
				case "complete":
					o.result, o.err = client.Complete(ctxFor(t), scope, miniMaxChatTarget(model), miniMaxChatRequest(), ai.ChatOptions{})
				case "complete-simple":
					o.result, o.err = client.CompleteSimple(ctxFor(t), scope, miniMaxChatTarget(model), miniMaxChatRequest(), ai.SimpleOptions{})
				case "stream-simple":
					o = drain(client.StreamSimple(ctxFor(t), scope, miniMaxChatTarget(model), miniMaxChatRequest(), ai.SimpleOptions{}))
				default:
					o = drain(client.Stream(ctxFor(t), scope, miniMaxChatTarget(model), miniMaxChatRequest(), nil))
				}
				o.record(ev)
				obs := rec.awaitCall(ev, scope.RequestID)
				ev.Check("attributed observer lifecycle", len(obs) == 4 && !obs[0].Call.Resolved && obs[3].Call.ProviderID == ai.ProviderMiniMaxCN && obs[3].Call.ModelID == model, "%+v", obs)
				ev.Record("requests", w.provider.Requests())
				ev.Record("candidate", w.client.Catalog())
				if !ev.Check("one successful attempt", o.err == nil && len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%v", o.err) {
					return
				}
				fields := bodyFields(t, w.provider.Requests()[0].Body)
				for _, key := range []string{"store", "prompt_cache_key", "prompt_cache_retention", "thinking", "reasoning_effort", "tool_stream", "max_tokens"} {
					ev.Check("omitted "+key, len(fields[key]) == 0, "%s", w.provider.Requests()[0].Body)
				}
				ev.Check("split reasoning and trailing usage requested", string(fields["reasoning_split"]) == "true" && jsonEqual(fields["stream_options"], json.RawMessage(`{"include_usage":true}`)), "%s", w.provider.Requests()[0].Body)
				var msgs []map[string]any
				mustUnmarshal(t, fields["messages"], &msgs)
				ev.Check("system role", len(msgs) == 2 && msgs[0]["role"] == "system", "%v", msgs)
				ev.Check("resource and credential", w.provider.Requests()[0].Path == "/v1/chat/completions" && w.provider.Requests()[0].KeyAlias == minimaxChatKey.alias, "%+v", w.provider.Requests()[0])
				ev.Check("all callbacks", slices.Equal(order, []string{"header", "payload", "response"}), "%v", order)
				m := o.result.Message
				ev.Check("reasoning and text", len(m.Content) == 2 && m.Content[0].(ai.Thinking).Thinking == "Check." && m.Content[1].(ai.Text).Text == "Hello MiniMax.", "%+v", m.Content)
				ev.Check("trailing usage normalized", m.Usage.Input == 10 && m.Usage.CacheRead == 2 && m.Usage.Output == 8 && m.Usage.TotalTokens == 20 && o.result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "%+v", m.Usage)
			})
		}
	}
}
