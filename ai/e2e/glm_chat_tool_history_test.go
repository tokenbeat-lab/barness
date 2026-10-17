package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Characterization through the new route: existing native-state degradation
// remains the public contract, including visible reasoning and tool pairing.
func TestGLMChatToolHistory(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, mode := range []string{"trusted-restored", "untrusted", "cross-model", "cross-provider", "cross-api", "cross-account", "cross-model-account"} {
			t.Run(model+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "GLM03-tool-history-"+model+"-"+mode)
				w := newGLMChatWorld(t, nil)
				sc, raw := glmToolFixture(t, model)
				ev.Fixture("tools.json", raw)
				enqueue(ev, w, chatEvents(t, sc.Tools, ""))
				enqueue(ev, w, chatEvents(t, sc.Answer, ""))
				req, target := glmToolRequest(model), ai.Target{BindingID: "glm-chat", ModelID: model}
				first, err := w.client.Complete(ctxFor(t), textScope("glm-history-first"), target, req, nil)
				ev.Record("first-result", first)
				if !ev.Check("first call completed tools", err == nil && first.Message.StopReason == ai.StopReasonToolUse, "%v", err) {
					return
				}
				stored := persist(t, first.Message)
				ev.Record("stored-record", json.RawMessage(stored))
				message, env, _ := restore(t, stored)
				_, trusted := message.NativeState.Envelope()
				ev.Check("deserialization never restores trust", !trusted && env != nil, "trusted=%t env=%+v", trusted, env)
				if env == nil {
					return
				}
				want := ai.NativeStateDowngrades{}
				switch mode {
				case "untrusted":
					want.NoEnvelope = 1
				case "cross-model", "cross-model-account":
					want.CrossModel = 1
					target.ModelID = map[string]string{"glm-5.3": "glm-5.2", "glm-5.2": "glm-5.3"}[model]
				case "cross-provider":
					want.CrossModel = 1
					message.Provider, env.ProviderID = ai.ProviderOpenAI, ai.ProviderOpenAI
				case "cross-api":
					want.CrossModel = 1
					message.API, env.API = ai.APIOpenAIResponses, ai.APIOpenAIResponses
				case "cross-account":
					want.AccountMismatch = 1
				}
				if mode == "cross-account" || mode == "cross-model-account" {
					env.AccountScopeID = "different-authorized-history-account"
				}
				if mode != "untrusted" {
					state, err := ai.TrustNativeState(textScope("glm-history-restore"), *env)
					if !ev.Check("Host can vouch for its restored record", err == nil, "%v", err) {
						return
					}
					message.NativeState = state
				}
				req.Messages = append(req.Messages, message, ai.ToolResultText(message.Content[1].(ai.ToolCall), "42", false), ai.ToolResultText(message.Content[2].(ai.ToolCall), "UTC", false))
				second := drain(w.client.Stream(ctxFor(t), textScope("glm-history-second"), target, req, nil))
				second.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("history continues with expected downgrade", second.err == nil && second.result.Metadata.NativeStateDowngrades == want, "%v %+v", second.err, second.result.Metadata)
				reqs := w.provider.Requests()
				if !ev.Check("two explicit calls", len(reqs) == 2, "%d", len(reqs)) {
					return
				}
				var messages []map[string]any
				mustUnmarshal(t, bodyFields(t, reqs[1].Body)["messages"], &messages)
				checkGLMToolReplay(ev, messages, mode == "trusted-restored")
			})
		}
	}
}

func TestGLMChatPendingTools(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM03-pending-tools-"+model)
			w := newGLMChatWorld(t, nil)
			sc, raw := glmToolFixture(t, model)
			ev.Fixture("tools.json", raw)
			reply := chatEvents(t, sc.Tools[:2], "")
			reply.End = provider.EndHold
			enqueue(ev, w, reply)
			req := glmToolRequest(model)
			s := w.client.Stream(ctxFor(t), textScope("glm-pending-tools"), ai.Target{BindingID: "glm-chat", ModelID: model}, req, nil)
			defer s.Close()
			var events []json.RawMessage
			observed := false
			for s.Next() {
				e := s.Event()
				events = append(events, mustMarshal(t, eventsJSON([]ai.Event{e})[0]))
				if delta, ok := e.(ai.ToolCallDeltaEvent); ok {
					partial := delta.Partial.Snapshot()
					_, err := ai.ValidateToolCall(partial, 1, req.Tools)
					ev.Record("pending-snapshot", partial)
					ev.Check("Host cannot execute a streaming partial", partial.StopReason == ai.StopReasonPending && err != nil, "%v %+v", err, partial)
					observed = true
					break
				}
			}
			ev.Check("argument progress observable before completion", observed, "no delta")
			_ = s.Close()
			for s.Next() {
				events = append(events, mustMarshal(t, eventsJSON([]ai.Event{s.Event()})[0]))
			}
			res, err := s.Result()
			ev.Record("events", events)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("canceled incomplete turn never becomes executable", err != nil && res.Message.StopReason == ai.StopReasonAborted && len(res.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%v %+v", err, res)
		})
	}
}

func TestGLMChatIncompleteToolCalls(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, end := range []string{"eof", "length", "invalid-schema"} {
			t.Run(model+"/"+end, func(t *testing.T) {
				ev := run.Case(t, "GLM03-incomplete-tools-"+model+"-"+end)
				w := newGLMChatWorld(t, nil)
				sc, raw := glmToolFixture(t, model)
				ev.Fixture("tools.json", raw)
				events := sc.Tools[:2]
				if end == "length" {
					events = append(append([]json.RawMessage{}, events...), json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"length"}]}`), json.RawMessage(`"[DONE]"`))
				}
				if end == "invalid-schema" {
					events = []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_invalid","function":{"name":"lookup","arguments":"{\"q\":\"x\",\"n\":\"not-an-integer\"}"}}]},"finish_reason":"tool_calls"}]}`), json.RawMessage(`"[DONE]"`)}
				}
				enqueue(ev, w, chatEvents(t, events, ""))
				req := glmToolRequest(model)
				o := drain(w.client.Stream(ctxFor(t), textScope("glm-incomplete"), ai.Target{BindingID: "glm-chat", ModelID: model}, req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				index := 1
				if end == "invalid-schema" {
					index = 0
				}
				_, err := ai.ValidateToolCall(o.result.Message, index, req.Tools)
				ev.Check("Host cannot execute incomplete or invalid arguments", err != nil, "message=%+v", o.result.Message)
				if end == "eof" {
					ev.Check("missing terminal is a protocol error", o.err != nil && o.result.Message.StopReason == ai.StopReasonError, "%v", o.err)
				}
				if end == "length" {
					ev.Check("truncation retains length terminal", o.err == nil && o.result.Message.StopReason == ai.StopReasonLength, "%v", o.err)
				}
				ev.Check("library executes no Host tool loop", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata)
			})
		}
	}
}
