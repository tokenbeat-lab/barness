package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxChatProtocolContainers(t *testing.T) {
	for _, sc := range []struct {
		name   string
		change func(map[string]any)
		code   ai.Code
	}{
		{"stream", func(b map[string]any) { b["stream"] = false }, ai.CodeInvalidRequest},
		{"strict", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "function", "function": map[string]any{"name": "lookup", "parameters": map[string]any{"type": "object"}, "strict": true}}}
		}, ai.CodeInvalidRequest},
		{"legacy", func(b map[string]any) { b["functions"] = []any{} }, ai.CodeInvalidRequest},
		{"developer", func(b map[string]any) { b["messages"] = []any{map[string]any{"role": "developer", "content": "Hi"}} }, ai.CodeInvalidRequest},
		{"image", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://example.test/i.png"}}}}}
		}, ai.CodeInvalidRequest},
		{"no-user-content", func(b map[string]any) { b["messages"] = []any{map[string]any{"role": "user"}} }, ai.CodeInvalidRequest},
		{"bad-reasoning", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "assistant", "content": nil, "reasoning_content": true}}
		}, ai.CodeInvalidRequest},
		{"bad-details", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "assistant", "content": nil, "reasoning_details": []any{map[string]any{"type": "unknown", "data": "opaque"}}}}
		}, ai.CodeInvalidRequest},
		{"bad-call", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "assistant", "content": nil, "tool_calls": []any{map[string]any{"type": "function", "id": "a", "function": map[string]any{"name": "lookup", "arguments": map[string]any{}}}}}}
		}, ai.CodeInvalidRequest},
		{"cache", func(b map[string]any) { b["prompt_cache_key"] = "foreign" }, ai.CodeTenantDenied},
		{"store-true", func(b map[string]any) { b["store"] = true }, ai.CodeTenantDenied},
		{"store-false", func(b map[string]any) { b["store"] = false }, ai.CodeInvalidRequest},
		{"near-temp", func(b map[string]any) { b["temperature"] = json.Number("2.00000000000000001") }, ai.CodeInvalidRequest},
		{"near-top-p", func(b map[string]any) { b["top_p"] = json.Number("1.00000000000000001") }, ai.CodeInvalidRequest},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM02-container-"+sc.name)
			w := newMiniMaxChatWorld(t, nil)
			enqueue(ev, w, miniMaxChatReply(t, "MiniMax-M3"))
			o := drain(w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				sc.change(p.Body)
				return ai.KeepPayload(), nil
			}}).Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, sc.code, ai.PhaseRequest)
			ev.Check("no attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}
