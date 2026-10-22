package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func kimiResponsesToolRequest() ai.Request {
	r := kimiResponsesRequest()
	r.Tools = []ai.Tool{{Name: "lookup", Description: "Find a value.", Parameters: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`)}}
	return r
}

const kimiResponsesCallA = `{"id":"item|a","type":"function_call","call_id":"opaque|α/a","name":"lookup","arguments":"{\"q\":\"a\"}","status":"completed"}`
const kimiResponsesCallB = `{"id":"item|b","type":"function_call","call_id":"opaque|α/b","name":"lookup","arguments":"{\"q\":\"b\"}","status":"completed"}`

func kimiResponsesToolFrames() []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"type":"reasoning","id":"reason|α","summary":[],"encrypted_content":null}}`),
		json.RawMessage(`{"type":"response.reasoning_summary_text.delta","output_index":0,"content_index":0,"delta":"Check."}`),
		json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":` + kimiResponsesReasoning + `}`),
		json.RawMessage(`{"type":"response.output_item.added","output_index":1,"item":{"type":"function_call","id":"item|a","call_id":"opaque|α/a","name":"lookup","arguments":""}}`),
		json.RawMessage(`{"type":"response.function_call_arguments.delta","output_index":1,"delta":"{\"q\":"}`),
		json.RawMessage(`{"type":"response.function_call_arguments.delta","output_index":1,"delta":"\"a\"}"}`),
		json.RawMessage(`{"type":"response.output_item.done","output_index":1,"item":` + kimiResponsesCallA + `}`),
		json.RawMessage(`{"type":"response.output_item.done","output_index":2,"item":` + kimiResponsesCallB + `}`),
		json.RawMessage(`{"type":"response.completed","response":{"id":"kimi-tools","status":"completed","output":[` + kimiResponsesReasoning + `,` + kimiResponsesCallA + `,` + kimiResponsesCallB + `],"usage":` + kimiResponsesUsage + `}}`),
	}
}
func TestKimiResponsesHistory(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, mode := range []string{"native", "restored", "untrusted", "cross-account", "cross-market", "cross-protocol"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K06-K07-K08-")+mode)
				w := newKimiResponsesWorld(t, market, nil)
				frames := kimiResponsesToolFrames()
				enqueue(ev, w, kimiResponsesReply(t, frames...))
				enqueue(ev, w, kimiResponsesReply(t))
				req := kimiResponsesToolRequest()
				first, err := w.client.Complete(ctxFor(t), textScope("kimi-r-history-"+string(market)+mode), kimiResponsesTarget(), req, ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "auto"})})
				ev.Fixture("tool-frames.json", mustMarshal(t, frames))
				ev.Record("first", first)
				if !ev.Check("two distinct opaque calls", err == nil && len(first.Message.Content) == 3 && first.Message.StopReason == ai.StopReasonToolUse, "%v %+v", err, first.Message) {
					return
				}
				m := first.Message
				before := mustMarshal(t, m)
				ev.Check("versioned local pair encoding", strings.HasPrefix(m.Content[1].(ai.ToolCall).ID, "kmr1|") && m.Content[1].(ai.ToolCall).ID != m.Content[2].(ai.ToolCall).ID, "%+v", m.Content)
				if mode != "native" {
					stored := persist(t, m)
					ev.Record("stored", json.RawMessage(stored))
					var env *ai.NativeStateEnvelope
					m, env, _ = restore(t, stored)
					if mode != "untrusted" && env != nil {
						switch mode {
						case "cross-account":
							env.AccountScopeID = "another-account"
						case "cross-market":
							if market == ai.ProviderMoonshot {
								env.ProviderID = ai.ProviderMoonshotCN
								m.Provider = ai.ProviderMoonshotCN
							} else {
								env.ProviderID = ai.ProviderMoonshot
								m.Provider = ai.ProviderMoonshot
							}
						case "cross-protocol":
							env.API = ai.APIOpenAICompletions
							m.API = ai.APIOpenAICompletions
						}
						state, e := ai.TrustNativeState(textScope("kimi-r-restore"), *env)
						if e != nil {
							t.Fatal(e)
						}
						m.NativeState = state
					}
				}
				req.Messages = append(req.Messages, m)
				for _, block := range m.Content {
					if c, ok := block.(ai.ToolCall); ok {
						req.Messages = append(req.Messages, ai.ToolResultText(c, "found", false))
					}
				}
				second := drain(w.client.Stream(ctxFor(t), textScope("kimi-r-continue-"+string(market)+mode), kimiResponsesTarget(), req, nil))
				second.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("history continues", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
					return
				}
				var input []map[string]any
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
				reasoning := 0
				calls, results := []string{}, []string{}
				for _, item := range input {
					if item["type"] == "reasoning" {
						reasoning++
						ev.Check("summary projection exact", len(item) == 2 && jsonEqual(mustMarshal(t, item["summary"]), json.RawMessage(`[{"type":"summary_text","text":"Check."},{"type":"summary_text","text":"Then answer."}]`)), "%v", item)
					}
					if item["type"] == "function_call" {
						calls = append(calls, item["call_id"].(string))
						ev.Check("input function fields only", len(item) == 4, "%v", item)
					}
					if item["type"] == "function_call_output" {
						results = append(results, item["call_id"].(string))
					}
					if item["role"] == "assistant" {
						ev.Check("assistant message omits output fields", item["id"] == nil && item["status"] == nil && item["phase"] == nil, "%v", item)
					}
				}
				native := mode == "native" || mode == "restored"
				ev.Check("only trusted same-source reasoning replayed", (reasoning == 1) == native, "%v", input)
				ev.Check("all tool results pair uniquely", len(calls) == 2 && len(results) == 2 && calls[0] != calls[1] && calls[0] == results[0] && calls[1] == results[1], "%v %v", calls, results)
				if native {
					ev.Check("original opaque call ids restored", calls[0] == "opaque|α/a" && calls[1] == "opaque|α/b", "%v", calls)
					ev.Check("raw native state not rewritten", string(mustMarshal(t, first.Message)) == string(before), "native state mutated")
				}
			})
		}
	}
}
func TestKimiResponsesPlainAssistantProjection(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, kimiResponsesID(market, "K07-visible-assistant"))
			w := newKimiResponsesWorld(t, market, nil)
			enqueue(ev, w, kimiResponsesReply(t))
			req := kimiResponsesRequest()
			req.Messages = append(req.Messages, ai.AssistantMessage{Provider: market, API: ai.APIOpenAIResponses, Model: "kimi-k3", Content: []ai.AssistantContent{ai.Text{Text: "Earlier."}}})
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), req, nil))
			o.record(ev)
			if !ev.Check("visible history succeeds", o.err == nil, "%v", o.err) {
				return
			}
			ev.Check("no OpenAI synthesized output ids", !strings.Contains(string(w.provider.Requests()[0].Body), "msg_pi_"), "%s", w.provider.Requests()[0].Body)
		})
	}
}
