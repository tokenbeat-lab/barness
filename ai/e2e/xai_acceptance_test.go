package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIAcceptanceToolChoices(t *testing.T) {
	for _, choice := range []string{"auto", "none"} {
		t.Run(choice, func(t *testing.T) {
			ev := run.Case(t, "XAI-E06-explicit-success-"+choice)
			w := newXAIWorld(t, nil)
			output := xaiMessage
			stop := ai.StopReasonStop
			if choice == "auto" {
				output, stop = xaiCallA, ai.StopReasonToolUse
			}
			enqueue(ev, w, xaiReply(t, xaiTerminal(output, xaiUsage)))
			res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiToolRequest(), ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: choice})})
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			ev.Check("explicit tool policy accepts matching output", err == nil && res.Message.StopReason == stop, "%v %+v", err, res.Message)
			if len(w.provider.Requests()) == 1 {
				ev.Check("wire carries explicit choice", string(bodyFields(t, w.provider.Requests()[0].Body)["tool_choice"]) == `"`+choice+`"`, "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}

func TestXAIAcceptanceReturnedNativeToolRoundTrip(t *testing.T) {
	for _, stored := range []bool{false, true} {
		name := "direct"
		if stored {
			name = "restored"
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E07-E08-returned-native-tool-"+name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t, xaiTerminal(xaiReasoning+`,`+xaiCallA, xaiUsage)), xaiReply(t))
			first, err := w.client.Complete(ctxFor(t), textScope("native-first-"+name), xaiTarget(), xaiToolRequest(), nil)
			ev.Record("first-result", first)
			if !ev.Check("first call returns actual native tool state", err == nil && first.Message.StopReason == ai.StopReasonToolUse && len(first.Message.Content) == 2, "%v %+v", err, first.Message) {
				return
			}
			msg := first.Message
			if stored {
				data := persist(t, msg)
				ev.Fixture("saved-turn.json", data)
				var env *ai.NativeStateEnvelope
				msg, env, _ = restore(t, data)
				if env == nil {
					t.Fatal("returned native envelope missing")
				}
				msg.NativeState, err = ai.TrustNativeState(textScope("native-next-"+name), *env)
				if err != nil {
					t.Fatal(err)
				}
			}
			call, ok := msg.Content[1].(ai.ToolCall)
			if !ok {
				t.Fatal("returned block is not a tool call")
			}
			req := xaiToolRequest()
			req.Messages = append(req.Messages, msg, ai.ToolResultText(call, "lookup returned A", false))
			last, lastErr := w.client.Complete(ctxFor(t), textScope("native-next-"+name), xaiTarget(), req, nil)
			ev.Record("last-result", last)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("public continuation succeeds", lastErr == nil && len(w.provider.Requests()) == 2, "%v", lastErr) {
				return
			}
			var input []map[string]json.RawMessage
			mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
			var native, callID, itemID, outputID string
			for _, item := range input {
				switch string(item["type"]) {
				case `"reasoning"`:
					native = string(mustMarshal(t, item))
				case `"function_call"`:
					mustUnmarshal(t, item["call_id"], &callID)
					mustUnmarshal(t, item["id"], &itemID)
				case `"function_call_output"`:
					mustUnmarshal(t, item["call_id"], &outputID)
				}
			}
			ev.Check("returned ciphertext and opaque tool pair replay exactly", jsonEqual([]byte(native), []byte(xaiReasoning)) && callID == "call|α" && itemID == "item|α" && outputID == callID, "native=%s call=%q item=%q output=%q", native, callID, itemID, outputID)
		})
	}
}

func TestXAIAcceptanceCrossAPIToolAssociations(t *testing.T) {
	ev := run.Case(t, "XAI-E07-cross-api-exact-tool-associations")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	a := ai.ToolCall{ID: "foreign|first", Name: "lookup", Arguments: ai.UncheckedArguments(`{"q":"A"}`)}
	b := ai.ToolCall{ID: "foreign|second", Name: "lookup", Arguments: ai.UncheckedArguments(`{"q":"B"}`)}
	msg := xaiTrustedMessage(t, a, b)
	msg.API = ai.APIOpenAICompletions
	req := xaiToolRequest()
	req.Messages = append(req.Messages, msg, ai.ToolResultText(b, "result B", false), ai.ToolResultText(a, "result A", false))
	res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil)
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("cross-API tools project successfully", err == nil && len(w.provider.Requests()) == 1, "%v", err) {
		return
	}
	var input []map[string]any
	mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
	calls, outputs := map[string]string{}, map[string]string{}
	for _, item := range input {
		if item["type"] == "function_call" {
			args, _ := item["arguments"].(string)
			id, _ := item["call_id"].(string)
			calls[args] = id
		}
		if item["type"] == "function_call_output" {
			output, _ := item["output"].(string)
			id, _ := item["call_id"].(string)
			outputs[output] = id
		}
	}
	ev.Check("renamed calls remain distinct and match their own reversed results", len(calls) == 2 && len(outputs) == 2 && calls[`{"q":"A"}`] != "" && calls[`{"q":"B"}`] != "" && calls[`{"q":"A"}`] != calls[`{"q":"B"}`] && calls[`{"q":"A"}`] != "foreign|first" && calls[`{"q":"B"}`] != "foreign|second" && outputs["result A"] == calls[`{"q":"A"}`] && outputs["result B"] == calls[`{"q":"B"}`], "calls=%v outputs=%v", calls, outputs)
}

func TestXAIAcceptanceUserImageOrder(t *testing.T) {
	ev := run.Case(t, "XAI-E09-user-text-png-text-jpeg-order")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	png, jpeg := xaiTestImage(t, "image/png"), xaiTestImage(t, "image/jpeg")
	req := ai.Request{Messages: []ai.Message{ai.UserMessage{Content: []ai.UserContent{ai.Text{Text: "before"}, png, ai.Text{Text: "between"}, jpeg}}}}
	res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil)
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("mixed user input accepted", err == nil && len(w.provider.Requests()) == 1, "%v", err) {
		return
	}
	var input []struct {
		Content []struct {
			Type, Text string
			URL        string `json:"image_url"`
		}
	}
	mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
	if !ev.Check("one user message retains four distinct parts", len(input) == 1 && len(input[0].Content) == 4, "%+v", input) {
		return
	}
	p := input[0].Content
	ev.Check("wire order and image payloads are exact", p[0].Type == "input_text" && p[0].Text == "before" && p[1].Type == "input_image" && p[1].URL == "data:image/png;base64,"+png.Data && p[2].Type == "input_text" && p[2].Text == "between" && p[3].Type == "input_image" && p[3].URL == "data:image/jpeg;base64,"+jpeg.Data && !strings.Contains(p[0].Text, "between"), "%+v", p)
}

func TestXAIAcceptanceCollidingOpaqueTuples(t *testing.T) {
	ev := run.Case(t, "XAI-E07-delimiter-colliding-returned-tuples")
	w := newXAIWorld(t, nil)
	output := `{"type":"function_call","id":"c","call_id":"a|b","name":"lookup","arguments":"{\"q\":\"A\"}"},{"type":"function_call","id":"b|c","call_id":"a","name":"lookup","arguments":"{\"q\":\"B\"}"}`
	enqueue(ev, w, xaiReply(t, xaiTerminal(output, xaiUsage)), xaiReply(t))
	first, err := w.client.Complete(ctxFor(t), textScope("tuple-first"), xaiTarget(), xaiToolRequest(), nil)
	ev.Record("first-result", first)
	if !ev.Check("two returned tuple identities survive", err == nil && len(first.Message.Content) == 2, "%v %+v", err, first.Message) {
		return
	}
	a, okA := first.Message.Content[0].(ai.ToolCall)
	b, okB := first.Message.Content[1].(ai.ToolCall)
	if !ev.Check("naive delimiter collision has distinct local IDs", okA && okB && a.ID != "" && a.ID != b.ID, "first=%q second=%q", a.ID, b.ID) {
		return
	}
	req := xaiToolRequest()
	req.Messages = append(req.Messages, first.Message, ai.ToolResultText(b, "second", false), ai.ToolResultText(a, "first", false))
	res, err := w.client.Complete(ctxFor(t), textScope("tuple-next"), xaiTarget(), req, nil)
	ev.Record("next-result", res)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("collision-free continuation succeeds", err == nil && len(w.provider.Requests()) == 2, "%v", err) {
		return
	}
	var input []map[string]any
	mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
	calls, outputs := map[string]string{}, map[string]string{}
	for _, item := range input {
		if item["type"] == "function_call" {
			id, _ := item["id"].(string)
			call, _ := item["call_id"].(string)
			calls[id] = call
		}
		if item["type"] == "function_call_output" {
			output, _ := item["output"].(string)
			call, _ := item["call_id"].(string)
			outputs[output] = call
		}
	}
	ev.Check("each original call and result retain exact paired IDs", len(calls) == 2 && calls["c"] == "a|b" && calls["b|c"] == "a" && outputs["first"] == "a|b" && outputs["second"] == "a", "calls=%v outputs=%v", calls, outputs)
}

func TestXAIAcceptanceExplicitEffortStateless(t *testing.T) {
	ev := run.Case(t, "XAI-E02-explicit-effort-retains-stateless-encryption")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), ai.ResponsesOptions{ReasoningEffort: ai.ThinkingHigh})
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("explicit effort succeeds", err == nil && len(w.provider.Requests()) == 1, "%v", err) {
		return
	}
	body := bodyFields(t, w.provider.Requests()[0].Body)
	ev.Check("explicit effort still disables store and requests encrypted reasoning", string(body["store"]) == "false" && string(body["include"]) == `["reasoning.encrypted_content"]` && jsonEqual(body["reasoning"], []byte(`{"effort":"high"}`)), "%s", w.provider.Requests()[0].Body)
}
