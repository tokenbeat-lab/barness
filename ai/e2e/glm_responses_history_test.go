package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type glmResponsesToolScript struct {
	Name   string            `json:"name"`
	Model  string            `json:"model"`
	Tools  []json.RawMessage `json:"tools"`
	Answer []json.RawMessage `json:"answer"`
}

func glmResponsesToolFixtures(t *testing.T) ([]glmResponsesToolScript, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-responses/tools.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Scenarios []glmResponsesToolScript `json:"scenarios"`
	}
	mustUnmarshal(t, raw, &f)
	return f.Scenarios, raw
}

func glmResponsesToolRequest() ai.Request {
	return ai.Request{Messages: []ai.Message{ai.UserText("Look up alpha and beta.")}, Tools: []ai.Tool{
		{Name: "lookup_probe_code", Description: "Look up a synthetic code.", Parameters: json.RawMessage(`{"type":"object","properties":{"label":{"type":"string"}},"required":["label"],"additionalProperties":false}`)},
	}}
}

func TestGLMResponsesToolRoundTrip(t *testing.T) {
	scenarios, raw := glmResponsesToolFixtures(t)
	for _, sc := range scenarios {
		t.Run(sc.Model+"/"+sc.Name, func(t *testing.T) {
			ev := run.Case(t, "GLM05-tools-"+sc.Model+"-"+sc.Name)
			ev.Fixture("tools.json", raw)
			w := newGLMResponsesWorld(t, nil)
			enqueue(ev, w, glmResponsesReply(t, sc.Tools))
			enqueue(ev, w, glmResponsesReply(t, sc.Answer))
			req, target := glmResponsesToolRequest(), glmResponsesTarget(sc.Model)
			first := drain(w.client.Stream(ctxFor(t), textScope("glm-responses-tools-first"), target, req, nil))
			first.record(ev)
			if !ev.Check("one generation exposes executable tools", first.err == nil && first.result.Message.StopReason == ai.StopReasonToolUse, "%v %+v", first.err, first.result.Message) {
				return
			}
			before := persist(t, first.result.Message)
			req.Messages = append(req.Messages, first.result.Message)
			var calls []ai.ToolCall
			for i, b := range first.result.Message.Content {
				if call, ok := b.(ai.ToolCall); ok {
					validated, err := ai.ValidateToolCall(first.result.Message, i, req.Tools)
					ev.Check("Host validates final tool arguments", err == nil && validated.ID() == call.ID, "%v %+v", err, call)
					if strings.HasPrefix(sc.Name, "schema-") && err == nil {
						var args struct {
							Label string `json:"label"`
						}
						want := "alpha"
						if len(calls) > 0 {
							want = "beta"
						}
						ev.Check("authoritative argument done replaces streamed value", validated.Decode(&args) == nil && args.Label == want && strings.Contains(call.ID, "|fc_"), "%+v %+v", args, call)
					}
					req.Messages = append(req.Messages, ai.ToolResultText(call, "BRN42", false))
					calls = append(calls, call)
				}
			}
			ev.Check("at least one function returned", len(calls) > 0, "%+v", calls)
			progress := false
			for _, event := range first.events {
				if _, ok := event.(ai.ToolCallDeltaEvent); ok {
					progress = true
				}
			}
			ev.Check("function argument progress is observable", progress, "%v", eventSummary(first.events))
			second := drain(w.client.Stream(ctxFor(t), textScope("glm-responses-tools-second"), target, req, nil))
			ev.Record("continuation-events", eventsJSON(second.events))
			ev.Record("continuation-result", second.result)
			ev.Record("continuation-error", errString(second.err))
			ev.Record("requests", w.provider.Requests())
			ev.Record("candidate-catalog", w.client.Catalog())
			ev.Check("Host continuation completes", second.err == nil && second.result.Message.StopReason == ai.StopReasonStop, "%v", second.err)
			ev.Check("trusted native origin retained", second.result.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{}, "%+v", second.result.Metadata)
			ev.Check("history native state never mutated", string(before) == string(persist(t, first.result.Message)), "history mutated")
			reqs := w.provider.Requests()
			if !ev.Check("two explicit requests", len(reqs) == 2, "%d", len(reqs)) {
				return
			}
			var input []map[string]any
			mustUnmarshal(t, bodyFields(t, reqs[1].Body)["input"], &input)
			var reasoning []map[string]any
			var callIDs, resultIDs []string
			for _, item := range input {
				switch item["type"] {
				case "message":
					_, id := item["id"]
					_, status := item["status"]
					_, phase := item["phase"]
					ev.Check("InputMessage is independent of output item identity", !id && !status && !phase && len(item) == 3, "%+v", item)
				case "reasoning":
					reasoning = append(reasoning, item)
				case "function_call":
					callIDs = append(callIDs, item["call_id"].(string))
					_, has := item["id"]
					ev.Check("GLM input does not reuse function item id", !has, "%+v", item)
				case "function_call_output":
					resultIDs = append(resultIDs, item["call_id"].(string))
				}
			}
			for _, b := range first.result.Message.Content {
				if thought, ok := b.(ai.Thinking); ok && thought.Signature != "" {
					if !ev.Check("native thought appears once", len(reasoning) > 0, "%+v", input) {
						continue
					}
					item := reasoning[0]
					reasoning = reasoning[1:]
					content, ok := item["content"].(map[string]any)
					ev.Check("documented InputReasoning object keeps visible text", len(item) == 2 && ok && len(content) == 2 && content["type"] == "reasoning_text" && content["text"] == thought.Thinking, "%+v", item)
				}
			}
			for i, call := range calls {
				id, _, _ := strings.Cut(call.ID, "|")
				ev.Check("call_id pairs result independently of item id", i < len(callIDs) && i < len(resultIDs) && callIDs[i] == id && resultIDs[i] == id, "calls=%v results=%v", callIDs, resultIDs)
			}
			fields := bodyFields(t, reqs[1].Body)
			ev.Check("all user and visible assistant messages use documented message discriminator", !strings.Contains(string(fields["input"]), `"msg_pi_`) && !strings.Contains(string(fields["input"]), `"msg_native"`) && input[0]["type"] == "message", "%s", fields["input"])
			ev.Check("every turn stateless with complete user history", string(fields["store"]) == "false" && strings.Contains(string(fields["input"]), "Look up alpha"), "%s", reqs[1].Body)
			for _, field := range []string{"previous_response_id", "conversation", "background"} {
				_, has := fields[field]
				ev.Check("no saved state "+field, !has, "%s", reqs[1].Body)
			}
		})
	}
}
