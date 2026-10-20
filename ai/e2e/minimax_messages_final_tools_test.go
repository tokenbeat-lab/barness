package e2e

import (
	"context"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxMessagesFinalToolFields(t *testing.T) {
	for _, c := range []struct{ name, tools string }{
		{"null", `null`},
		{"missing-name-schema", `[{}]`},
		{"name-number", `[{"name":7,"input_schema":{"type":"object"}}]`},
		{"name-blank", `[{"name":"  ","input_schema":{"type":"object"}}]`},
		{"schema-missing", `[{"name":"inspect"}]`},
		{"schema-null", `[{"name":"inspect","input_schema":null}]`},
		{"schema-array", `[{"name":"inspect","input_schema":[]}]`},
		{"schema-number", `[{"name":"inspect","input_schema":7}]`},
		{"schema-type-missing", `[{"name":"inspect","input_schema":{}}]`},
		{"schema-type-array", `[{"name":"inspect","input_schema":{"type":"array"}}]`},
		{"properties-array", `[{"name":"inspect","input_schema":{"type":"object","properties":[]}}]`},
		{"required-string", `[{"name":"inspect","input_schema":{"type":"object","required":"label"}}]`},
		{"required-number", `[{"name":"inspect","input_schema":{"type":"object","required":[7]}}]`},
		{"description-number", `[{"name":"inspect","description":7,"input_schema":{"type":"object"}}]`},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-final-tool-field-"+c.name)
			w := newMiniMaxSharedWorld(t, nil)
			reply := miniMaxMessagesReply(t, "MiniMax-M3")
			ev.Record("provider-script", reply)
			w.provider.EnqueueAt("/tenant-a/minimax-messages", reply)
			var tools any
			mustUnmarshal(t, []byte(c.tools), &tools)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["tools"] = tools
				ev.Record("hook-final-body", p.Body)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm17-tool-"+c.name), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("malformed final tool fields never reach Provider", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			w.released(ev)
		})
	}
}

func TestMiniMaxMessagesFinalBusinessHook(t *testing.T) {
	for _, c := range []struct{ name, messages, tools string }{
		{"string-content", `[{"role":"user","content":"Synthetic input."}]`, ""},
		{"business-containers", `[{"role":"user","content":[{"type":"text","text":"Synthetic input."}]},{"role":"assistant","content":[{"type":"tool_use","id":"call-17","name":"inspect","input":{"type":"file","file_id":"business-only","cache_control":{"strict":true},"input_schema":7,"properties":["anything"]}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-17","content":[{"type":"text","text":"Synthetic result."}],"is_error":false}]}]`, `[{"name":"inspect","description":"Synthetic business schema.","input_schema":{"type":"object","properties":{"type":{"type":"string"},"file_id":{"type":"string"},"cache_control":{"type":"object","properties":{"strict":{"type":"boolean"}}},"input_schema":{"type":"object"},"properties":{"type":"object"}},"required":["type"]}}]`},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-final-business-hook-"+c.name)
			w := newMiniMaxSharedWorld(t, nil)
			reply := miniMaxMessagesReply(t, "MiniMax-M3")
			ev.Record("provider-script", reply)
			w.provider.EnqueueAt("/tenant-a/minimax-messages", reply)
			var messages, tools any
			mustUnmarshal(t, []byte(c.messages), &messages)
			if c.tools != "" {
				mustUnmarshal(t, []byte(c.tools), &tools)
			}
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["messages"] = messages
				if c.tools != "" {
					p.Body["tools"] = tools
				}
				ev.Record("hook-final-body", p.Body)
				return ai.KeepPayload(), nil
			}}
			var o outcome
			o.result, o.err = w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("mm17-business-"+c.name), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil)
			o.record(ev)
			requests := w.provider.Requests()
			ev.Record("requests", requests)
			if ev.Check("legal final business containers reach Provider", o.err == nil && len(requests) == 1 && len(o.result.Metadata.Attempts) == 1, "%v %+v", o.err, o.result.Metadata) {
				body := bodyFields(t, requests[0].Body)
				ev.Check("final message business values unchanged", jsonEqual(body["messages"], []byte(c.messages)), "%s", requests[0].Body)
				if c.tools != "" {
					ev.Check("schema business properties unchanged", jsonEqual(body["tools"], []byte(c.tools)), "%s", requests[0].Body)
				}
			}
			w.released(ev)
		})
	}
}
