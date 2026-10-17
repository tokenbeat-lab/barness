package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxChatUnknownReasoningFormat(t *testing.T) {
	for _, raw := range []string{
		`"reasoning_details":{"text":"unknown"}`,
		`"reasoning_details":[{"type":"unknown","text":"opaque"}]`,
		`"reasoning_content":"visible","reasoning_details":[{"type":"reasoning.text","text":"other"}]`,
		`"reasoning_content":true`,
		`"reasoning":"unknown field"`,
	} {
		t.Run(raw, func(t *testing.T) {
			ev := run.Case(t, "MM02-unknown-reasoning-"+t.Name())
			w := newMiniMaxChatWorld(t, nil)
			frames := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{` + raw + `},"finish_reason":"stop"}]}`)}
			ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
			enqueue(ev, w, chatEvents(t, frames, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("unknown native format is never silently lost", o.err != nil && o.result.Message.StopReason == ai.StopReasonError && len(w.provider.Requests()) == 1, "%v %+v", o.err, o.result.Message)
		})
	}
}
