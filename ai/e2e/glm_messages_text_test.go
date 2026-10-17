package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Actual Messages SSE is replayed unchanged at the confirmed public Client seam.
// Each model has its own expected content/usage from its own authorized capture.
func TestGLMMessagesText(t *testing.T) {
	scenarios, prompt, raw := loadGLMMessages(t)
	for _, sc := range scenarios {
		for _, entry := range []string{"stream-nil", "stream-full", "complete-nil", "complete-full", "stream-simple", "complete-simple"} {
			for _, hooked := range []bool{false, true} {
				name := sc.Model + "/" + entry
				if hooked {
					name += "/hooks"
				}
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "GLM06-text-"+name)
					w := newGLMMessagesWorld(t, nil)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, glmMessagesReply(t, sc))
					client := w.client.WithHooks(ai.Hooks{})
					if hooked {
						client = w.client.WithHooks(ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
							h.Set("X-Host-Probe", "basic-messages")
							return nil
						}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							p.Body["max_tokens"] = 4096
							return ai.KeepPayload(), nil
						}})
					}
					scope := textScope("glm-messages-" + name)
					target := ai.Target{BindingID: "glm-messages", ModelID: sc.Model}
					req := ai.Request{Messages: []ai.Message{ai.UserText(prompt)}}
					var o outcome
					switch entry {
					case "stream-nil":
						o = drain(client.Stream(ctxFor(t), scope, target, req, nil))
					case "stream-full":
						o = drain(client.Stream(ctxFor(t), scope, target, req, ai.AnthropicOptions{MaxTokens: ai.Value(4096)}))
					case "complete-nil":
						o.result, o.err = client.Complete(ctxFor(t), scope, target, req, nil)
					case "complete-full":
						o.result, o.err = client.Complete(ctxFor(t), scope, target, req, ai.AnthropicOptions{MaxTokens: ai.Value(4096)})
					case "stream-simple":
						o = drain(client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(4096)}))
					case "complete-simple":
						o.result, o.err = client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(4096)})
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("candidate-catalog", w.client.Catalog())
					ev.Check("text call succeeds", o.err == nil, "err=%v", o.err)
					msg := o.result.Message
					ev.Check("exact native text/thinking/signature content preserved", jsonEqual(mustMarshal(t, msg.Content), sc.Content), "got %+v", msg.Content)
					ev.Check("Messages stop and response ID preserved", msg.StopReason == ai.StopReasonStop && msg.RawStopReason == "end_turn" && msg.ResponseID == sc.ResponseID, "%+v", msg)
					ev.Check("call attributed to its Messages account", o.result.Metadata.ProviderID == ai.ProviderZhipu && o.result.Metadata.API == ai.APIAnthropicMessages && o.result.Metadata.AccountScopeID == "acct-tenant-a-glm-messages", "%+v", o.result.Metadata)
					a := o.result.Metadata.Attempts
					if ev.Check("one attempt", len(a) == 1, "%+v", a) {
						ev.Check("actual final usage replaces initial zero and stays partial", a[0].UsageReporting == ai.UsagePartial && msg.Usage.Input == sc.Input && msg.Usage.Output == sc.Output && msg.Usage.TotalTokens == sc.Input+sc.Output && msg.Usage.CacheRead == 0 && msg.Usage.CacheWrite == 0 && msg.Usage.Reasoning.IsZero(), "%+v", a[0])
						ev.Check("GLM X-Request-Id attributed", a[0].ProviderRequestID == sc.RequestID[0], "%+v", a[0])
					}
					if !ev.Check("one actual request", len(w.provider.Requests()) == 1, "%d", len(w.provider.Requests())) {
						return
					}
					r := w.provider.Requests()[0]
					ev.Check("standard Messages path, accepted SDK query and x-api-key", r.Method == "POST" && r.Path == "/api/anthropic/v1/messages" && r.Query == "beta=true" && r.KeyAlias == glmA.alias && r.Header["Authorization"] == "" && r.Header["X-Api-Key"] == glmA.alias, "%+v", r)
					ev.Check("SDK protocol version retained without vendor beta/browser defaults", r.Header["Anthropic-Version"] == "2023-06-01" && r.Header["Anthropic-Beta"] == "" && r.Header["Anthropic-Dangerous-Direct-Browser-Access"] == "", "%+v", r.Header)
					var body map[string]any
					mustUnmarshal(t, r.Body, &body)
					ev.Check("basic body only; omitted thinking remains provider default", len(body) == 4 && body["model"] == sc.Model && body["stream"] == true && body["thinking"] == nil, "%s", r.Body)
					var history []map[string]any
					mustUnmarshal(t, mustMarshal(t, body["messages"]), &history)
					blocks, _ := history[0]["content"].([]any)
					ev.Check("no unproven cache marker", len(blocks) == 1 && len(blocks[0].(map[string]any)) == 2 && blocks[0].(map[string]any)["type"] == "text" && blocks[0].(map[string]any)["text"] == prompt, "%s", r.Body)
				})
			}
		}
	}
}
