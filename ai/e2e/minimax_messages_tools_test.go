package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func miniMaxToolRequest() ai.Request {
	req := miniMaxMessagesRequest()
	req.Tools = []ai.Tool{{Name: "inspect", Description: "Inspect a label.", Parameters: json.RawMessage(`{"type":"object","properties":{"label":{"type":"string"}},"required":["label"]}`)}}
	return req
}

func TestMiniMaxMessagesTools(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			t.Run(model+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "MM04-tools-"+model+"-"+entry)
				w := newMiniMaxMessagesWorld(t, nil)
				req := miniMaxToolRequest()
				script := miniMaxMessagesFixture(t, "tools.sse", model)
				ev.Fixture("synthetic-tools.sse", script)
				enqueue(ev, w, provider.SSE([][]byte{script}), miniMaxMessagesReply(t, model))
				invoke := func(id string) outcome {
					var o outcome
					switch entry {
					case "stream":
						o = drain(w.client.Stream(ctxFor(t), textScope(id), miniMaxMessagesTarget(model), req, nil))
					case "complete":
						o.result, o.err = w.client.Complete(ctxFor(t), textScope(id), miniMaxMessagesTarget(model), req, nil)
					case "stream-simple":
						o = drain(w.client.StreamSimple(ctxFor(t), textScope(id), miniMaxMessagesTarget(model), req, ai.SimpleOptions{}))
					case "complete-simple":
						o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope(id), miniMaxMessagesTarget(model), req, ai.SimpleOptions{})
					}
					return o
				}
				first := invoke("mm04-tools-first")
				first.record(ev)
				if !ev.Check("mixed blocks preserve wire order", first.err == nil && len(first.result.Message.Content) == 4 && first.result.Message.StopReason == ai.StopReasonToolUse, "%v %+v", first.err, first.result.Message.Content) {
					return
				}
				thought, ok := first.result.Message.Content[0].(ai.Thinking)
				ev.Check("thinking and signature fragments retained", ok && thought.Thinking == "Synthetic reasoning." && thought.Signature == "sig-part1-part2", "%+v", thought)
				for i, label := range []string{"alpha", "beta"} {
					call, err := ai.ValidateToolCall(first.result.Message, i+2, req.Tools)
					var args struct{ Label string }
					if !ev.Check("Host validates independently assembled tool arguments", err == nil && call.Decode(&args) == nil && args.Label == label, "%v %+v", err, args) {
						return
					}
				}
				req.Messages = append(req.Messages, first.result.Message)
				for _, b := range first.result.Message.Content {
					if call, ok := b.(ai.ToolCall); ok {
						req.Messages = append(req.Messages, ai.ToolResultText(call, "Synthetic tool result.", false))
					}
				}
				second := invoke("mm04-tools-final")
				ev.Record("final-result", second.result)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("Host tool round trip completes", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
					return
				}
				var body struct {
					Messages []struct {
						Role    string
						Content []json.RawMessage
					}
				}
				mustUnmarshal(t, w.provider.Requests()[1].Body, &body)
				ev.Check("signed native thinking replayed exactly", len(body.Messages) > 1 && jsonEqual(body.Messages[1].Content[0], []byte(`{"type":"thinking","thinking":"Synthetic reasoning.","signature":"sig-part1-part2"}`)), "%s", w.provider.Requests()[1].Body)
				ev.Check("same origin replay has no downgrade", second.result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "%+v", second.result.Metadata)
			})
		}
	}
}

func TestMiniMaxMessagesStartInput(t *testing.T) {
	for _, shape := range []string{"complete", "placeholder", "mixed"} {
		t.Run(shape, func(t *testing.T) {
			ev := run.Case(t, "MM04-start-input-"+shape)
			w := newMiniMaxMessagesWorld(t, nil)
			script := string(miniMaxMessagesFixture(t, "tools.sse", "MiniMax-M3"))
			frames := strings.Split(strings.TrimSpace(script), "\n\n")
			var keep []string
			for _, f := range frames {
				if strings.Contains(f, `"index":3`) {
					continue
				}
				if strings.Contains(f, `"index":2`) && strings.Contains(f, "content_block_start") && shape != "placeholder" {
					f = strings.ReplaceAll(f, `"input":{}`, `"input":{"label":"alpha"}`)
				}
				if shape == "complete" && strings.Contains(f, `"index":2`) && strings.Contains(f, "content_block_delta") {
					continue
				}
				keep = append(keep, f)
			}
			enqueue(ev, w, provider.SSE([][]byte{[]byte(strings.Join(keep, "\n\n") + "\n\n")}))
			req := miniMaxToolRequest()
			o := drain(w.client.Stream(ctxFor(t), textScope("mm04-input-"+shape), miniMaxMessagesTarget("MiniMax-M3"), req, nil))
			o.record(ev)
			if shape == "mixed" {
				ev.Check("unverified mixed complete input and deltas refused", o.err != nil, "%v", o.err)
				return
			}
			if !ev.Check("start input shape accepted", o.err == nil, "%v", o.err) {
				return
			}
			call, err := ai.ValidateToolCall(o.result.Message, 2, req.Tools)
			var args struct{ Label string }
			ev.Check("complete input or fragmented placeholder retained", err == nil && call.Decode(&args) == nil && args.Label == "alpha", "%v %+v", err, args)
		})
	}
}
