package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxMessagesFinalMessageContainers(t *testing.T) {
	for _, c := range []struct {
		name     string
		messages any
	}{
		{"number", json.Number("7")},
		{"null", nil},
		{"object", map[string]any{"role": "user", "content": "Synthetic input."}},
		{"empty", []any{}},
		{"scalar-element", []any{"Synthetic input."}},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-final-message-container-"+c.name)
			w := newMiniMaxSharedWorld(t, nil)
			reply := miniMaxMessagesReply(t, "MiniMax-M3")
			ev.Record("provider-script", reply)
			w.provider.EnqueueAt("/tenant-a/minimax-messages", reply)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["messages"] = c.messages
				ev.Record("hook-final-body", p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm17-container-"+c.name), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("malformed final messages never reach Provider", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			w.released(ev)
		})
	}
}

func TestMiniMaxMessagesFinalMessageFields(t *testing.T) {
	for _, c := range []struct{ name, messages string }{
		{"role-missing", `[{"content":"Synthetic input."}]`},
		{"role-number", `[{"role":7,"content":"Synthetic input."}]`},
		{"role-unknown", `[{"role":"system","content":"Synthetic input."}]`},
		{"content-missing", `[{"role":"user"}]`},
		{"content-null", `[{"role":"user","content":null}]`},
		{"content-number", `[{"role":"user","content":7}]`},
		{"content-object", `[{"role":"user","content":{"type":"text","text":"Synthetic input."}}]`},
		{"content-empty", `[{"role":"user","content":[]}]`},
		{"block-scalar", `[{"role":"user","content":["Synthetic input."]}]`},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-final-message-field-"+c.name)
			w := newMiniMaxSharedWorld(t, nil)
			reply := miniMaxMessagesReply(t, "MiniMax-M3")
			ev.Record("provider-script", reply)
			w.provider.EnqueueAt("/tenant-a/minimax-messages", reply)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["messages"] = json.RawMessage(c.messages)
				ev.Record("hook-final-body", p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm17-fields-"+c.name), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("malformed final message fields never reach Provider", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			w.released(ev)
		})
	}
}

func TestMiniMaxMessagesFinalBlockFields(t *testing.T) {
	for _, c := range []struct{ name, role, block string }{
		{"type-missing", "user", `{"text":"Synthetic input."}`},
		{"text-missing", "user", `{"type":"text"}`},
		{"text-number", "user", `{"type":"text","text":7}`},
		{"thinking-number", "assistant", `{"type":"thinking","thinking":7,"signature":"synthetic-signed"}`},
		{"tool-use-id-missing", "assistant", `{"type":"tool_use","name":"inspect","input":{}}`},
		{"tool-use-name-number", "assistant", `{"type":"tool_use","id":"call-17","name":7,"input":{}}`},
		{"tool-use-input-null", "assistant", `{"type":"tool_use","id":"call-17","name":"inspect","input":null}`},
		{"tool-use-input-array", "assistant", `{"type":"tool_use","id":"call-17","name":"inspect","input":[]}`},
		{"tool-use-user", "user", `{"type":"tool_use","id":"call-17","name":"inspect","input":{}}`},
		{"tool-result-id-missing", "user", `{"type":"tool_result","content":"Synthetic result."}`},
		{"tool-result-content-number", "user", `{"type":"tool_result","tool_use_id":"call-17","content":7}`},
		{"tool-result-error-string", "user", `{"type":"tool_result","tool_use_id":"call-17","content":"Synthetic result.","is_error":"false"}`},
		{"tool-result-assistant", "assistant", `{"type":"tool_result","tool_use_id":"call-17","content":"Synthetic result."}`},
		{"tool-result-nested-use", "user", `{"type":"tool_result","tool_use_id":"call-17","content":[{"type":"tool_use","id":"nested","name":"inspect","input":{}}]}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-final-block-field-"+c.name)
			w := newMiniMaxSharedWorld(t, nil)
			reply := miniMaxMessagesReply(t, "MiniMax-M3")
			ev.Record("provider-script", reply)
			w.provider.EnqueueAt("/tenant-a/minimax-messages", reply)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["messages"] = json.RawMessage(`[{"role":"` + c.role + `","content":[` + c.block + `]}]`)
				ev.Record("hook-final-body", p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm17-block-"+c.name), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("malformed final block fields never reach Provider", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			w.released(ev)
		})
	}
}
