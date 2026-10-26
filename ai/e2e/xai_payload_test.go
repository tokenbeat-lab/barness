package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIInputAndToolsSubset(t *testing.T) {
	for _, sc := range []struct {
		name   string
		change func(map[string]any)
		valid  bool
		denied bool
	}{
		{"business-schema", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "function", "name": "probe", "parameters": map[string]any{"type": "object", "properties": map[string]any{"file_id": map[string]any{"type": "string"}, "store": map[string]any{"type": "boolean"}, "model": map[string]any{"type": "string"}}, "required": []any{"file_id"}, "additionalProperties": true}}}
		}, true, false},
		{"business-arguments", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call", "name": "probe", "call_id": "c", "arguments": `{"file_id":"business","store":true,"model":"x"}`}, map[string]any{"type": "function_call_output", "call_id": "c", "output": `{"file_id":"business"}`}}
		}, true, false},
		{"strict", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "function", "name": "probe", "parameters": map[string]any{"type": "object"}, "strict": false}}
		}, false, false},
		{"combination-root", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "function", "name": "probe", "parameters": map[string]any{"oneOf": []any{map[string]any{"type": "object"}}}}}
		}, false, false},
		{"hosted-allowed", func(b map[string]any) { b["tools"] = []any{map[string]any{"type": "web_search"}} }, false, false},
		{"input-string", func(b map[string]any) { b["input"] = "shortcut" }, false, false},
		{"developer", func(b map[string]any) {
			b["input"] = []any{map[string]any{"role": "developer", "content": "unexpected"}}
		}, false, false},
		{"native-added", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "reasoning", "id": "foreign", "summary": []any{}, "encrypted_content": "foreign"}}
		}, false, true},
		{"native-id-added", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "message", "role": "assistant", "id": "foreign", "content": []any{map[string]any{"type": "output_text", "text": "history"}}}}
		}, false, true},
		{"orphan-output", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call_output", "call_id": "foreign", "output": "result"}}
		}, false, false},
		{"duplicate-calls", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call", "call_id": "c", "name": "probe", "arguments": "{}"}, map[string]any{"type": "function_call", "call_id": "c", "name": "probe", "arguments": "{}"}, map[string]any{"type": "function_call_output", "call_id": "c", "output": "result"}}
		}, false, false},
		{"arguments-array", func(b map[string]any) {
			b["input"] = []any{map[string]any{"type": "function_call", "call_id": "c", "name": "probe", "arguments": "[]"}, map[string]any{"type": "function_call_output", "call_id": "c", "output": "result"}}
		}, false, false},
		{"unsupported-item", func(b map[string]any) { b["input"] = []any{map[string]any{"type": "web_search_call", "id": "w"}} }, false, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E03-input-tools-"+sc.name)
			w := newXAIWorld(t, nil)
			binding := xaiBinding()
			binding.AllowedHostedTools = []string{"web_search"}
			w.host.PutBinding(binding)
			enqueue(ev, w, xaiReply(t))
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				sc.change(p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if sc.valid {
				ev.Check("business JSON remains legal", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
			} else {
				code := ai.CodeCallbackFailed
				if sc.denied {
					code = ai.CodeTenantDenied
				}
				checkRejected(ev, o, code, ai.PhaseRequest)
				ev.Check("zero HTTP", len(w.provider.Requests()) == 0, "requests=%+v", w.provider.Requests())
			}
		})
	}
}

func TestXAINativeHookCannotChangeSource(t *testing.T) {
	ev := run.Case(t, "XAI-E03-native-reference-protected")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	native := `{"type":"reasoning","id":"source","summary":[],"encrypted_content":"source-cipher"}`
	req := xaiRequest()
	req.Messages = append(req.Messages, xaiTrustedMessage(t, ai.Thinking{Signature: native}))
	h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		for _, raw := range p.Body["input"].([]any) {
			item := raw.(map[string]any)
			if item["type"] == "reasoning" {
				item["encrypted_content"] = "changed"
			}
		}
		return ai.KeepPayload(), nil
	}}
	o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
	o.record(ev)
	checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
	ev.Check("zero HTTP", len(w.provider.Requests()) == 0, "%d", len(w.provider.Requests()))
}

func TestXAIToolsCountBoundary(t *testing.T) {
	for _, count := range []int{350, 351} {
		t.Run(string(rune('a'+count-350)), func(t *testing.T) {
			ev := run.Case(t, "XAI-E04-tools-count-"+t.Name())
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			req := xaiRequest()
			for i := 0; i < count; i++ {
				req.Tools = append(req.Tools, ai.Tool{Name: "tool" + string(rune(0x400+i)), Parameters: json.RawMessage(`{"type":"object"}`)})
			}
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
			o.record(ev)
			if count == 350 {
				ev.Check("350 functions accepted", o.err == nil, "%v", o.err)
			} else {
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("351 functions rejected before credential", w.host.ReadsFor(t.Name()).Credentials == 0, "reads=%+v", w.host.ReadsFor(t.Name()))
			}
		})
	}
}
