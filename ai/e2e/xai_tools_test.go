package e2e

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const xaiCallA = `{"id":"item|α","type":"function_call","call_id":"call|α","name":"lookup","arguments":"{\"q\":\"A\"}"}`
const xaiCallB = `{"id":"item|β","type":"function_call","call_id":"call|β","name":"lookup","arguments":"{\"q\":\"B\"}"}`

func xaiToolRequest() ai.Request {
	req := xaiRequest()
	req.Tools = []ai.Tool{{Name: "lookup", Description: "Look up a value", Parameters: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}}}`)}}
	return req
}
func TestXAIStreamTools(t *testing.T) {
	for _, sc := range []struct {
		name, output, choice   string
		parallelFalse, invalid bool
		frames                 []json.RawMessage
	}{
		{name: "final-only", output: xaiCallA, choice: "required"},
		{name: "parallel-same-name", output: xaiCallA + `,` + xaiCallB},
		{name: "done-only", output: xaiCallA, frames: []json.RawMessage{json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":` + xaiCallA + `}`)}},
		{name: "arguments-done", output: xaiCallA, frames: []json.RawMessage{
			json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"item|α","type":"function_call","call_id":"call|α","name":"lookup","arguments":""}}`),
			json.RawMessage(`{"type":"response.function_call_arguments.delta","output_index":0,"item_id":"item|α","delta":"{\"q\":"}`),
			json.RawMessage(`{"type":"response.function_call_arguments.done","output_index":0,"item_id":"item|α","arguments":"{\"q\":\"A\"}"}`)}},
		{name: "required-missing", output: xaiMessage, choice: "required", invalid: true},
		{name: "none-returned", output: xaiCallA, choice: "none", invalid: true},
		{name: "parallel-disabled", output: xaiCallA + `,` + xaiCallB, parallelFalse: true, invalid: true},
		{name: "partial-arguments", output: `{"id":"f","type":"function_call","call_id":"c","name":"lookup","arguments":"{\"q\":"}`, invalid: true},
		{name: "array-arguments", output: `{"id":"f","type":"function_call","call_id":"c","name":"lookup","arguments":"[]"}`, invalid: true},
		{name: "empty-arguments", output: `{"id":"f","type":"function_call","call_id":"c","name":"lookup","arguments":""}`, invalid: true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E06-output-"+sc.name)
			w := newXAIWorld(t, nil)
			frames := append(sc.frames, xaiTerminal(sc.output, xaiUsage))
			enqueue(ev, w, xaiReply(t, frames...))
			ev.Record("frames", frames)
			opts := ai.ResponsesOptions{}
			if sc.choice != "" {
				opts.ToolChoice = ai.Value(ai.ResponsesToolChoice{Mode: sc.choice})
			}
			if sc.parallelFalse {
				opts.SamplingParams = map[string]json.RawMessage{"parallel_tool_calls": json.RawMessage(`false`)}
			}
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiToolRequest(), opts))
			o.record(ev)
			if sc.invalid {
				ev.Check("completed output enforces tool contract", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				return
			}
			ev.Check("tools accepted as complete calls", o.err == nil && o.result.Message.StopReason == ai.StopReasonToolUse, "%v %+v", o.err, o.result.Message)
			want := 1
			if sc.name == "parallel-same-name" {
				want = 2
			}
			if !ev.Check("one block per call", len(o.result.Message.Content) == want, "%+v", o.result.Message.Content) {
				return
			}
			for i, block := range o.result.Message.Content {
				call, ok := block.(ai.ToolCall)
				pairRaw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(call.ID, "xar1|"))
				var pair []string
				_ = json.Unmarshal(pairRaw, &pair)
				expectedCall, expectedItem, args := "call|α", "item|α", `{"q":"A"}`
				if i == 1 {
					expectedCall, expectedItem, args = "call|β", "item|β", `{"q":"B"}`
				}
				ev.Check("opaque pair and complete raw arguments retained", ok && err == nil && len(pair) == 2 && pair[0] == expectedCall && pair[1] == expectedItem && call.Name == "lookup" && call.RawArguments == args, "%+v pair=%q", call, pair)
			}
		})
	}
}
