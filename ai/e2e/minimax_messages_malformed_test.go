package e2e

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxMessagesMalformedContent(t *testing.T) {
	for _, c := range []struct{ name, block, delta string }{
		{"text-null", `{"type":"text","text":null}`, ``},
		{"text-delta-missing", `{"type":"text","text":""}`, `{"type":"text_delta"}`},
		{"thinking-delta-null", `{"type":"thinking","thinking":""}`, `{"type":"thinking_delta","thinking":null}`},
		{"signature-null", `{"type":"thinking","thinking":""}`, `{"type":"signature_delta","signature":null}`},
		{"tool-id-missing", `{"type":"tool_use","name":"inspect","input":{}}`, ``},
		{"tool-input-array", `{"type":"tool_use","id":"c","name":"inspect","input":[]}`, ``},
		{"tool-delta-missing", `{"type":"tool_use","id":"c","name":"inspect","input":{}}`, `{"type":"input_json_delta"}`},
		{"redacted-missing-data", `{"type":"redacted_thinking"}`, ``},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-malformed-"+c.name)
			w := newMiniMaxMessagesWorld(t, nil)
			frames := []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"synthetic-malformed","model":"MiniMax-M3"}}`), json.RawMessage(`{"type":"content_block_start","index":0,"content_block":` + c.block + `}`)}
			if c.delta != "" {
				frames = append(frames, json.RawMessage(`{"type":"content_block_delta","index":0,"delta":`+c.delta+`}`))
			}
			frames = append(frames, json.RawMessage(`{"type":"content_block_stop","index":0}`), json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"}}`), json.RawMessage(`{"type":"message_stop"}`))
			enqueue(ev, w, sseEvents(t, frames, provider.FramingLF))
			o := drain(w.client.Stream(ctxFor(t), textScope("mm04-malformed-"+c.name), miniMaxMessagesTarget("MiniMax-M3"), miniMaxToolRequest(), nil))
			o.record(ev)
			ev.Check("malformed required content is a protocol failure", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
		})
	}
}

func TestMiniMaxMessagesIncompleteToolJSON(t *testing.T) {
	ev := run.Case(t, "MM04-incomplete-tool-json")
	w := newMiniMaxMessagesWorld(t, nil)
	frames := []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"synthetic-incomplete","model":"MiniMax-M3"}}`), json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"c","name":"inspect","input":{}}}`), json.RawMessage(`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"label\":"}}`), json.RawMessage(`{"type":"content_block_stop","index":0}`), json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"tool_use"}}`), json.RawMessage(`{"type":"message_stop"}`)}
	enqueue(ev, w, sseEvents(t, frames, provider.FramingLF))
	o := drain(w.client.Stream(ctxFor(t), textScope("mm04-incomplete-tool-json"), miniMaxMessagesTarget("MiniMax-M3"), miniMaxToolRequest(), nil))
	o.record(ev)
	ev.Check("completed block cannot bless unfinished tool JSON", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
}
