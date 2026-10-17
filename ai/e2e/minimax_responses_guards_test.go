package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxResponsesInputAndToolBoundary(t *testing.T) {
	for _, sc := range []struct {
		name   string
		change func(map[string]any)
		valid  bool
	}{
		{"strict-schema-business-key", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "function", "name": "probe", "parameters": map[string]any{"type": "object", "properties": map[string]any{"strict": map[string]any{"type": "boolean"}, "image_url": map[string]any{"type": "string"}}}}}
		}, true},
		{"input-image", func(b map[string]any) {
			b["input"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_image", "image_url": "data:image/png;base64,AA=="}}}}
		}, false},
		{"tool-result-image", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call_output", "call_id": "opaque", "output": []any{map[string]any{"type": "input_image", "image_url": "data:image/png;base64,AA=="}}}}
		}, false},
		{"invalid-call-id-type", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call", "call_id": json.Number("1"), "name": "probe", "arguments": "{}"}}
		}, false},
		{"invalid-argument-type", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call", "call_id": "opaque|%", "name": "probe", "arguments": map[string]any{}}}
		}, false},
		{"invisible-reasoning-extra", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "reasoning", "summary": []any{}, "encrypted_content": "opaque"}}
		}, false},
		{"input-extra-state", func(b map[string]any) {
			b["input"] = []any{map[string]any{"role": "user", "content": "hello", "id": "other-account-resource"}}
		}, false},
		{"metadata-not-string", func(b map[string]any) { b["metadata"] = map[string]any{"synthetic": true} }, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM03-input-boundary-"+sc.name)
			w := newMiniMaxResponsesWorld(t, nil)
			enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				sc.change(p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if sc.valid {
				ev.Check("business schema properties remain legal", o.err == nil, "%v", o.err)
			} else {
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("zero attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			}
		})
	}
}
