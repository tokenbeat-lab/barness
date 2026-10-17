package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGLMChatOriginNonemptyResponseID(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, destination := range []string{"openai", "deepseek"} {
			t.Run(model+"/"+destination, func(t *testing.T) {
				ev := run.Case(t, "GLM05-empty-origin-id-"+t.Name())
				w := newGLMSharedWorld(t, func(cfg *ai.Config) { cfg.Catalog.Models = append(cfg.Catalog.Models, ai.BuiltinCatalog().Models...) })
				w.host.PutBinding(primaryBinding(tenantA, w.provider.URL()))
				w.host.PutCredential(primaryCredential(tenantA, "v1"))
				w.host.PutBinding(deepseekBinding(tenantA, w.provider.URL()))
				w.host.PutCredential(deepseekCredential(tenantA))
				source := glmSharedRoute{"chat", ai.APIOpenAICompletions}
				events := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"中文","type":"function","function":{"name":"lookup_probe_code","arguments":"{\"label\":\"source\"}"}}]},"finish_reason":"tool_calls"}]}`), json.RawMessage(`"[DONE]"`)}
				ev.Fixture("chat-unicode-id.json", mustMarshal(t, events))
				w.provider.EnqueueAt(source.prefix(glmA), chatEvents(t, events, provider.FramingLF))
				req := glmResponsesToolRequest()
				first, err := w.client.Complete(ctxFor(t), textScope("glm-empty-origin-first"), source.target(model), req, nil)
				ev.Record("first-result", first)
				if !ev.Check("Chat returns own opaque Unicode ID", err == nil && len(first.Message.Content) == 1 && first.Message.Content[0].(ai.ToolCall).ID == "中文", "%v %+v", err, first.Message) {
					return
				}
				call := first.Message.Content[0].(ai.ToolCall)
				req.Messages = append(req.Messages, first.Message, ai.ToolResultText(call, "unicode-result", false))
				target := ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}
				if destination == "deepseek" {
					target = ai.Target{BindingID: "deepseek", ModelID: "deepseek-flash"}
				}
				w.provider.Enqueue(glmResponsesReply(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}))
				o := drain(w.client.Stream(ctxFor(t), textScope("glm-empty-origin-second"), target, req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("downgraded history completes", o.err == nil, "%v", o.err)
				qs := w.provider.Requests()
				if len(qs) != 2 {
					return
				}
				_, calls, results := glmSharedWirePairs(t, glmSharedRoute{"responses", ai.APIOpenAIResponses}, qs[1].Body)
				ev.Check("single source call gets a nonempty target ID and valid result link", len(calls) == 1 && len(results) == 1 && calls[0] != "" && calls[0] == results[0], "%v %v", calls, results)
				w.released(ev)
			})
		}
	}
}
