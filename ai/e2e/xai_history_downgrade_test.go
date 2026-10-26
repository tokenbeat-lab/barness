package e2e

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIHistoryDowngradesToOpenAIResponses(t *testing.T) {
	for _, mode := range []string{"parallel", "separate-turns", "occupied-legacy-id"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-review-foreign-history-"+mode)
			w := newWorld(t, tenantA)
			enqueue(ev, w, xaiReply(t))
			a := ai.ToolCall{ID: xaiPublicPair(t, "call|甲", "item|甲"), Name: "probe", Arguments: ai.UncheckedArguments(`{"label":"first"}`)}
			b := ai.ToolCall{ID: xaiPublicPair(t, "call|乙", "item|乙"), Name: "probe", Arguments: ai.UncheckedArguments(`{"label":"second"}`)}
			msg := xaiTrustedMessage(t, ai.Thinking{Thinking: "visible reasoning", Signature: `{"type":"reasoning","id":"private-reasoning","summary":[],"encrypted_content":"private-cipher"}`}, ai.Text{Text: "visible answer", Signature: `{"v":1,"id":"private-message"}`}, a)
			req := xaiRequest()
			if mode == "separate-turns" {
				req.Messages = append(req.Messages, msg, ai.ToolResultText(a, "first-result", false), xaiTrustedMessage(t, b), ai.ToolResultText(b, "second-result", false))
			} else {
				msg.Content = append(msg.Content, b)
				req.Messages = append(req.Messages, msg, ai.ToolResultText(a, "first-result", false), ai.ToolResultText(b, "second-result", false))
			}
			if mode == "occupied-legacy-id" {
				legacy := ai.ToolCall{ID: "xar1|fc_legacy", Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
				req.Messages = append(req.Messages, ai.AssistantMessage{Provider: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, Model: "gpt-4.1-mini", StopReason: ai.StopReasonToolUse, Content: []ai.AssistantContent{legacy}}, ai.ToolResultText(legacy, "legacy-result", false))
			}
			before := persist(t, msg)
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}, req, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("Host-selected destination accepts downgraded history", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			raw := w.provider.Requests()[0].Body
			var input []map[string]json.RawMessage
			mustUnmarshal(t, bodyFields(t, raw)["input"], &input)
			var calls, results, outputs []string
			for _, item := range input {
				switch string(item["type"]) {
				case `"function_call"`:
					var call string
					mustUnmarshal(t, item["call_id"], &call)
					calls = append(calls, call)
					ev.Check("foreign account item IDs stripped", item["id"] == nil, "%s", mustMarshal(t, item))
				case `"function_call_output"`:
					var call, output string
					mustUnmarshal(t, item["call_id"], &call)
					mustUnmarshal(t, item["output"], &output)
					results, outputs = append(results, call), append(outputs, output)
				}
			}
			want := []string{"first-result", "second-result"}
			if mode == "occupied-legacy-id" {
				want = append(want, "legacy-result")
			}
			ev.Check("opaque calls retain unique target associations", len(calls) == len(want) && len(slices.Compact(slices.Sorted(slices.Values(calls)))) == len(want), "calls=%v", calls)
			ev.Check("each visible result answers its own call", slices.Equal(calls, results) && slices.Equal(outputs, want), "calls=%v results=%v outputs=%v", calls, results, outputs)
			ev.Check("visible content survives without private state", strings.Contains(string(raw), "visible reasoning") && strings.Contains(string(raw), "visible answer") && !strings.Contains(string(raw), "private-cipher") && !strings.Contains(string(raw), "private-message") && !strings.Contains(string(raw), "item|"), "%s", raw)
			ev.Check("source history immutable", string(before) == string(persist(t, msg)), "history changed")
		})
	}
}
