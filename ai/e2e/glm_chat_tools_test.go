package e2e

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

type glmToolScript struct {
	Model  string            `json:"model"`
	Tools  []json.RawMessage `json:"tools"`
	Answer []json.RawMessage `json:"answer"`
}

func glmToolFixture(t *testing.T, model string) (glmToolScript, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-chat/tools.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct{ Scenarios []glmToolScript }
	mustUnmarshal(t, raw, &f)
	for _, sc := range f.Scenarios {
		if sc.Model == model {
			return sc, raw
		}
	}
	t.Fatalf("missing tool fixture for %s", model)
	return glmToolScript{}, nil
}

func glmToolRequest(model string) ai.Request {
	return ai.Request{Messages: []ai.Message{
		ai.AssistantMessage{API: ai.APIOpenAICompletions, Provider: ai.ProviderZhipu, Model: model, Content: []ai.AssistantContent{ai.Text{Text: "Earlier."}}},
		ai.UserText("Look up x twice and the UTC clock."),
	}, Tools: []ai.Tool{
		{Name: "lookup", Description: "Lookup a synthetic value.", Parameters: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"},"n":{"type":"integer"}},"required":["q","n"],"additionalProperties":false}`)},
		{Name: "clock", Description: "Read a synthetic clock.", Parameters: json.RawMessage(`{"type":"object","properties":{"zone":{"type":"string"}},"required":["zone"],"additionalProperties":false}`)},
	}}
}

func TestGLMChatToolRoundTrip(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, choice := range []string{"default", "auto"} {
			t.Run(model+"/"+choice, func(t *testing.T) {
				ev := run.Case(t, "GLM03-tool-round-trip-"+model+"-"+choice)
				w := newGLMChatWorld(t, nil)
				sc, raw := glmToolFixture(t, model)
				ev.Fixture("tools.json", raw)
				enqueue(ev, w, chatEvents(t, sc.Tools, ""))
				enqueue(ev, w, chatEvents(t, sc.Answer, ""))
				ev.Record("candidate-catalog", w.client.Catalog())
				req, target := glmToolRequest(model), ai.Target{BindingID: "glm-chat", ModelID: model}
				opts := ai.ChatOptions{ReasoningEffort: ai.ThinkingHigh}
				if choice == "auto" {
					opts.ToolChoice = ai.Value(ai.ChatToolChoice{Mode: "auto"})
				}
				first := drain(w.client.Stream(ctxFor(t), textScope("glm-tools-"+model+choice), target, req, opts))
				first.record(ev)
				if !ev.Check("one turn returns complete tool calls", first.err == nil && first.result.Message.StopReason == ai.StopReasonToolUse && len(first.result.Message.Content) == 3, "%v %+v", first.err, first.result.Message) {
					return
				}
				ev.Check("partial deltas precede complete end events", slices.Equal(eventSummary(first.events), []string{
					"start", "thinking_start#0", "thinking_delta#0:Check ",
					"toolcall_start#1", "toolcall_delta#1:{\"q\":\"", "toolcall_start#2", "toolcall_delta#2:{\"zone\":\"",
					"thinking_delta#0:both.", "toolcall_delta#1:x\",\"n\":", "toolcall_delta#2:UTC\"}", "toolcall_delta#1:2}",
					"thinking_end#0:Check both.", `toolcall_end#1:call_a lookup {"q":"x","n":2}`, `toolcall_end#2:call_b clock {"zone":"UTC"}`, "done:toolUse",
				}), "%v", eventSummary(first.events))
				callA, errA := ai.ValidateToolCall(first.result.Message, 1, req.Tools)
				callB, errB := ai.ValidateToolCall(first.result.Message, 2, req.Tools)
				if !ev.Check("Host validates both complete calls", errA == nil && errB == nil, "%v / %v", errA, errB) {
					return
				}
				var a struct {
					Q string
					N int
				}
				var b struct{ Zone string }
				ev.Check("index and ID fragments assemble independently", callA.ID() == "call_a" && callB.ID() == "call_b" && callA.Decode(&a) == nil && a.Q == "x" && a.N == 2 && callB.Decode(&b) == nil && b.Zone == "UTC", "%+v %+v", a, b)
				message := first.result.Message
				req.Messages = append(req.Messages, message, ai.ToolResultText(message.Content[1].(ai.ToolCall), "42", false), ai.ToolResultText(message.Content[2].(ai.ToolCall), "UTC", false))
				second := drain(w.client.Stream(ctxFor(t), textScope("glm-answer-"+model+choice), target, req, opts))
				ev.Record("continuation-events", eventsJSON(second.events))
				ev.Record("continuation-result", second.result)
				ev.Record("continuation-error", errString(second.err))
				ev.Record("requests", w.provider.Requests())
				ev.Check("Host's separate generation answers", second.err == nil && second.result.Message.StopReason == ai.StopReasonStop && len(second.result.Message.Content) == 1 && second.result.Message.Content[0].(ai.Text).Text == "The answer is 42 at UTC.", "%v %+v", second.err, second.result.Message)
				ev.Check("trusted tool turn keeps native state", second.result.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{}, "%+v", second.result.Metadata)
				for _, r := range []ai.Result{first.result, second.result} {
					ev.Check("each call has one correctly attributed Attempt", len(r.Metadata.Attempts) == 1 && r.Metadata.ProviderID == ai.ProviderZhipu && r.Metadata.API == ai.APIOpenAICompletions && r.Metadata.ModelID == model && r.Metadata.Attempts[0].Usage == r.Message.Usage, "%+v", r.Metadata)
				}
				ev.Check("trailing usage retained in each turn", first.result.Message.Usage.Input == 12 && first.result.Message.Usage.Output == 8 && second.result.Message.Usage.Input == 30 && second.result.Message.Usage.Output == 12, "%+v %+v", first.result.Message.Usage, second.result.Message.Usage)
				reqs := w.provider.Requests()
				if !ev.Check("two explicit Host inference requests", len(reqs) == 2, "%d", len(reqs)) {
					return
				}
				for _, r := range reqs {
					fields := bodyFields(t, r.Body)
					ev.Check("function streams request tool_stream", string(fields["tool_stream"]) == "true", "%s", r.Body)
					ev.Check("cross-turn preserved thinking never enabled", string(fields["thinking"]) == `{"type":"enabled"}`, "%s", r.Body)
				}
				var messages []map[string]any
				mustUnmarshal(t, bodyFields(t, reqs[1].Body)["messages"], &messages)
				checkGLMToolReplay(ev, messages, true)
			})
		}
	}
}

func checkGLMToolReplay(ev *evidence.Case, messages []map[string]any, native bool) {
	if !ev.Check("visible history and valid tool pairs retained", len(messages) == 5, "%+v", messages) {
		return
	}
	ev.Check("plain assistant omits DeepSeek-only empty reasoning", messages[0]["content"] == "Earlier." && messages[0]["reasoning_content"] == nil, "%+v", messages[0])
	ev.Check("same-turn reasoning obeys trust", (native && messages[2]["reasoning_content"] == "Check both.") || (!native && messages[2]["reasoning_content"] == nil), "%+v", messages[2])
	if !native {
		ev.Check("downgraded reasoning stays visible", messages[2]["content"] == "Check both.", "%+v", messages[2])
	}
	calls, _ := messages[2]["tool_calls"].([]any)
	if !ev.Check("exactly two calls replay", len(calls) == 2, "%+v", messages[2]) {
		return
	}
	a, b := calls[0].(map[string]any), calls[1].(map[string]any)
	ev.Check("call IDs and result order pair correctly", a["id"] == "call_a" && b["id"] == "call_b" && messages[3]["role"] == "tool" && messages[3]["tool_call_id"] == "call_a" && messages[3]["content"] == "42" && messages[4]["role"] == "tool" && messages[4]["tool_call_id"] == "call_b" && messages[4]["content"] == "UTC", "%+v", messages)
}
