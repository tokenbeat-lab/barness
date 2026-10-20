package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMResponsesSimpleToolChoice(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, choice := range []ai.ToolChoice{ai.ToolChoiceAuto, ai.ToolChoiceNone} {
			for _, entry := range []string{"stream-simple", "complete-simple"} {
				t.Run(model+"/"+string(choice)+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "GLM05-simple-tools-"+t.Name())
					w := newGLMResponsesWorld(t, nil)
					scenarios, raw := glmResponsesFixtures(t)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
					opts := ai.SimpleOptions{ToolChoice: ai.Value(choice)}
					var o outcome
					if entry == "stream-simple" {
						o = drain(w.client.StreamSimple(ctxFor(t), textScope("glm-simple-tools"), glmResponsesTarget(model), glmResponsesToolRequest(), opts))
					} else {
						o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope("glm-simple-tools"), glmResponsesTarget(model), glmResponsesToolRequest(), opts)
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("simple documented tool choice accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
					if len(w.provider.Requests()) == 1 {
						ev.Check("simple choice retained", string(bodyFields(t, w.provider.Requests()[0].Body)["tool_choice"]) == `"`+string(choice)+`"`, "%s", w.provider.Requests()[0].Body)
					}
				})
			}
		}
	}
}

func TestGLMResponsesIncompleteTools(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, mode := range []string{"eof", "length", "invalid-schema"} {
			t.Run(model+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "GLM05-incomplete-tools-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				events := []json.RawMessage{json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"type":"function_call","id":"fc_partial","call_id":"call_partial","name":"lookup_probe_code","arguments":""}}`), json.RawMessage(`{"type":"response.function_call_arguments.delta","output_index":0,"delta":"{\"label\":"}`)}
				if mode == "length" {
					events = append(events, json.RawMessage(`{"type":"response.incomplete","response":{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[]}}`))
				}
				if mode == "invalid-schema" {
					events = []json.RawMessage{json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":{"type":"function_call","id":"fc_invalid","call_id":"call_invalid","name":"lookup_probe_code","arguments":"{\"missing\":42}"}}`), json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}
				}
				enqueue(ev, w, glmResponsesReply(t, events))
				req := glmResponsesToolRequest()
				s := w.client.Stream(ctxFor(t), textScope("glm-incomplete-tools"), glmResponsesTarget(model), req, nil)
				var seen []ai.Event
				for s.Next() {
					e := s.Event()
					seen = append(seen, e)
					if delta, ok := e.(ai.ToolCallDeltaEvent); ok {
						snapshot := delta.Partial.Snapshot()
						_, err := ai.ValidateToolCall(snapshot, 0, req.Tools)
						ev.Check("pending parameters cannot execute", err != nil, "%v %+v", err, snapshot)
					}
				}
				res, err := s.Result()
				o := outcome{result: res, err: err, events: seen}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				_, invalid := ai.ValidateToolCall(res.Message, 0, req.Tools)
				ev.Check("incomplete or invalid call never executable", invalid != nil, "%+v", res.Message)
				if mode == "eof" {
					ev.Check("EOF fails", err != nil && res.Message.StopReason == ai.StopReasonError, "%v", err)
				}
				if mode == "length" {
					ev.Check("length retains valid terminal", err == nil && res.Message.StopReason == ai.StopReasonLength, "%v", err)
				}
				ev.Check("Host owns tool loop", len(w.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "%+v", res.Metadata)
			})
		}
	}
}

func TestGLMResponsesBasicNativeHistory(t *testing.T) {
	scenarios, raw := glmResponsesFixtures(t)
	for _, sc := range scenarios {
		if sc.Model == "" || !strings.HasPrefix(sc.Name, "observed-array-") {
			continue
		}
		t.Run(sc.Model, func(t *testing.T) {
			ev := run.Case(t, "GLM05-own-basic-native-"+sc.Model)
			ev.Fixture("text.json", raw)
			w := newGLMResponsesWorld(t, nil)
			enqueue(ev, w, glmResponsesReply(t, sc.Events))
			enqueue(ev, w, glmResponsesReply(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}))
			req := glmResponsesRequest()
			first, err := w.client.Complete(ctxFor(t), textScope("glm-basic-first"), glmResponsesTarget(sc.Model), req, nil)
			ev.Record("first-result", first)
			ev.Check("own observed basic reasoning retained", err == nil && len(first.Message.Content) >= 2, "%v %+v", err, first.Message)
			req.Messages = append(req.Messages, first.Message, ai.UserText("Continue."))
			second := drain(w.client.Stream(ctxFor(t), textScope("glm-basic-second"), glmResponsesTarget(sc.Model), req, nil))
			second.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("own basic native-history conversion works offline", second.err == nil && second.result.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{}, "%v", second.err)
			if len(w.provider.Requests()) == 2 {
				var input []map[string]any
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
				count := 0
				for _, item := range input {
					if item["type"] == "reasoning" {
						count++
						content, ok := item["content"].(map[string]any)
						ev.Check("own reasoning uses documented object", ok && len(item) == 2 && content["text"] == sc.Thinking, "%+v", item)
					}
				}
				ev.Check("one own-model reasoning source replayed", count == 1, "%+v", input)
			}
		})
	}
}

func TestGLMResponsesMalformedTrustedNative(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, signature := range []string{`not-json`, `{"type":"reasoning","content":null}`, `{"type":"reasoning","content":{"type":"reasoning_text","text":1}}`, `{"type":"message","content":[]}`} {
			t.Run(model+"/"+signature, func(t *testing.T) {
				ev := run.Case(t, "GLM05-malformed-native-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				env := ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: "acct-tenant-a-glm-responses", ProviderID: ai.ProviderZhipu, API: ai.APIOpenAIResponses, ModelID: model}
				state, err := ai.TrustNativeState(textScope("glm-malformed-trust"), env)
				ev.Check("Host can establish envelope", err == nil, "%v", err)
				req := glmResponsesRequest()
				req.Messages = append(req.Messages, ai.AssistantMessage{Provider: ai.ProviderZhipu, API: ai.APIOpenAIResponses, Model: model, NativeState: state, Content: []ai.AssistantContent{ai.Thinking{Thinking: "Visible.", Signature: signature}}})
				o := drain(w.client.Stream(ctxFor(t), textScope("glm-malformed-native"), glmResponsesTarget(model), req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("no silent native loss or network Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}
