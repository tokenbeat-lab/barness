package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMResponsesNativeHistory(t *testing.T) {
	scenarios, raw := glmResponsesToolFixtures(t)
	for _, sc := range scenarios {
		if sc.Name != "schema-native-array" {
			continue
		}
		for _, mode := range []string{"trusted-restored", "untrusted", "cross-model", "cross-provider", "cross-api", "cross-account", "cross-model-account"} {
			t.Run(sc.Model+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "GLM05-native-"+t.Name())
				ev.Fixture("tools.json", raw)
				w := newGLMResponsesWorld(t, nil)
				enqueue(ev, w, glmResponsesReply(t, sc.Tools))
				enqueue(ev, w, glmResponsesReply(t, sc.Answer))
				req, target := glmResponsesToolRequest(), glmResponsesTarget(sc.Model)
				first, err := w.client.Complete(ctxFor(t), textScope("glm-native-first"), target, req, nil)
				if !ev.Check("first turn complete", err == nil, "%v", err) {
					return
				}
				stored := persist(t, first.Message)
				ev.Record("stored-record", json.RawMessage(stored))
				message, env, _ := restore(t, stored)
				_, trusted := message.NativeState.Envelope()
				ev.Check("deserialize requires Host trust", !trusted && env != nil, "%v", trusted)
				if env == nil {
					return
				}
				want := ai.NativeStateDowngrades{}
				switch mode {
				case "untrusted":
					want.NoEnvelope = 1
				case "cross-model", "cross-model-account":
					want.CrossModel = 1
					target.ModelID = map[string]string{"glm-5.3": "glm-5.2", "glm-5.2": "glm-5.3"}[sc.Model]
				case "cross-provider":
					want.CrossModel = 1
					message.Provider, env.ProviderID = ai.ProviderOpenAI, ai.ProviderOpenAI
				case "cross-api":
					want.CrossModel = 1
					message.API, env.API = ai.APIOpenAICompletions, ai.APIOpenAICompletions
				case "cross-account":
					want.AccountMismatch = 1
				}
				if mode == "cross-account" || mode == "cross-model-account" {
					env.AccountScopeID = "another-authorized-history-account"
				}
				if mode != "untrusted" {
					state, err := ai.TrustNativeState(textScope("glm-native-restore"), *env)
					ev.Check("Host trust restored", err == nil, "%v", err)
					message.NativeState = state
				}
				before := persist(t, message)
				req.Messages = append(req.Messages, message)
				for _, block := range message.Content {
					if call, ok := block.(ai.ToolCall); ok {
						req.Messages = append(req.Messages, ai.ToolResultText(call, "BRN42", false))
					}
				}
				second := drain(w.client.Stream(ctxFor(t), textScope("glm-native-second"), target, req, nil))
				second.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("legitimate history continues with scoped degradation", second.err == nil && second.result.Metadata.NativeStateDowngrades == want, "%v %+v", second.err, second.result.Metadata)
				ev.Check("input state is immutable", string(before) == string(persist(t, message)), "mutated")
				if len(w.provider.Requests()) != 2 {
					return
				}
				var input []map[string]any
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
				native := 0
				calls, results := []string{}, []string{}
				for _, item := range input {
					switch item["type"] {
					case "reasoning":
						native++
					case "function_call":
						calls = append(calls, item["call_id"].(string))
					case "function_call_output":
						results = append(results, item["call_id"].(string))
					}
				}
				expected := 0
				if mode == "trusted-restored" {
					expected = 1
				}
				ev.Check("only scoped trusted state is native", native == expected, "%+v", input)
				ev.Check("visible text and thinking retained", strings.Contains(string(w.provider.Requests()[1].Body), "Check both.") && strings.Contains(string(w.provider.Requests()[1].Body), "Checking."), "%+v", input)
				ev.Check("downgrade keeps two valid tool pairs", len(calls) == 2 && len(results) == 2 && calls[0] == results[0] && calls[1] == results[1] && calls[0] != calls[1], "%v %v", calls, results)
			})
		}
	}
}

func TestGLMResponsesVisibleThinkingFallback(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, mode := range []string{"done-only", "empty-native"} {
			t.Run(model+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "GLM05-visible-thinking-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				events := []json.RawMessage{json.RawMessage(`{"type":"response.reasoning_text.done","output_index":0,"text":"Visible reasoning without native content."}`)}
				if mode == "empty-native" {
					events = append(events, json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"rs_empty","content":[]}}`))
				}
				events = append(events, json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`))
				enqueue(ev, w, glmResponsesReply(t, events))
				enqueue(ev, w, glmResponsesReply(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}))
				req := glmResponsesRequest()
				first, err := w.client.Complete(ctxFor(t), textScope("glm-visible-first"), glmResponsesTarget(model), req, nil)
				ev.Record("first-result", first)
				ev.Check("visible thinking available", err == nil && len(first.Message.Content) == 1, "%v %+v", err, first.Message)
				req.Messages = append(req.Messages, first.Message, ai.UserText("Continue."))
				second := drain(w.client.Stream(ctxFor(t), textScope("glm-visible-second"), glmResponsesTarget(model), req, nil))
				second.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("visible-only history continues", second.err == nil, "%v", second.err)
				if len(w.provider.Requests()) == 2 {
					var input []map[string]any
					mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
					native := 0
					for _, item := range input {
						if item["type"] == "reasoning" {
							native++
						}
					}
					ev.Check("ordinary visible history survives without invented native item", native == 0 && strings.Contains(string(w.provider.Requests()[1].Body), "Visible reasoning without native content."), "%+v", input)
				}
			})
		}
	}
}
