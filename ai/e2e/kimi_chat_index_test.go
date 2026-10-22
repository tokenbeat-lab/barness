package e2e

import (
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func TestKimiChatExactToolIndex(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, entry := range []string{"stream", "complete"} {
			for _, sc := range []struct {
				name, index string
				valid       bool
			}{
				{"null", "null", false}, {"rounded-fraction", "1.00000000000000001", false}, {"fraction", "0.5", false}, {"negative", "-1", false}, {"overflow", "9223372036854775808", false}, {"unsafe-float", "9007199254740992", false}, {"missing", "", false}, {"string", `"0"`, false}, {"zero", "0", true}, {"nonzero", "3", true}, {"max-exact", "9007199254740991", true},
			} {
				t.Run(string(market)+"/"+entry+"/"+sc.name, func(t *testing.T) {
					ev := run.Case(t, kimiChatID(market, "K12-exact-index-")+entry+"-"+sc.name)
					w := newKimiChatWorld(t, market, nil)
					prefix := `"index":` + sc.index + `,`
					if sc.index == "" {
						prefix = ""
					}
					frames := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{` + prefix + `"id":"opaque|α","function":{"name":"lookup","arguments":"{\"q\":"}}]}}]}`), json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{` + prefix + `"function":{"arguments":"\"value\"}"}}]},"finish_reason":"tool_calls"}]}`)}
					ev.Fixture("synthetic.json", mustMarshal(t, frames))
					enqueue(ev, w, chatEvents(t, frames, ""))
					var o outcome
					if entry == "complete" {
						o.result, o.err = w.client.Complete(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatToolRequest(), nil)
					} else {
						o = drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatToolRequest(), nil))
					}
					o.record(ev)
					if sc.valid {
						ev.Check("legal split index remains executable", o.err == nil && o.result.Message.StopReason == ai.StopReasonToolUse && strings.Contains(o.result.Message.Content[0].(ai.ToolCall).RawArguments, "value"), "%v", o.err)
					} else {
						ev.Check("invalid raw index cannot create executable success", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
					}
				})
			}
		}
	}
}
