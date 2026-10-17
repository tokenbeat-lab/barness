package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// These controlled faults mutate an independent real tool stream one variable
// at a time. They do not claim that GLM emitted any malformed sequence.
func TestGLMMessagesSequence(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		c := glmMessagesCaptureByName(t, model+"-tool")
		for _, fault := range []string{"no-message-start", "duplicate-message-start", "duplicate-index", "unopened-delta", "wrong-delta-kind", "open-block-at-terminal", "premature-message-stop", "after-terminal", "event-type-mismatch"} {
			t.Run(model+"/"+fault, func(t *testing.T) {
				ev := run.Case(t, "GLM07-sequence-"+model+"-"+fault)
				w := newGLMMessagesWorld(t, nil)
				r := c.reply(t)
				frames := strings.Split(strings.TrimSpace(r.Script), "\n\n")
				find := func(kind string) int {
					for i, f := range frames {
						if strings.HasPrefix(f, "event: "+kind+"\n") {
							return i
						}
					}
					panic("missing source event " + kind)
				}
				switch fault {
				case "no-message-start":
					frames = frames[1:]
				case "duplicate-message-start":
					frames = append([]string{frames[0]}, frames...)
				case "duplicate-index":
					i := find("content_block_start")
					frames = append(frames[:i], append([]string{frames[i]}, frames[i:]...)...)
				case "unopened-delta":
					frames = append(frames[:1], append([]string{"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":99,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{}\"}}"}, frames[1:]...)...)
				case "wrong-delta-kind":
					i := find("content_block_delta")
					frames[i] = strings.ReplaceAll(strings.ReplaceAll(frames[i], "input_json_delta", "text_delta"), "thinking_delta", "text_delta")
				case "open-block-at-terminal":
					i := find("content_block_stop")
					frames = append(frames[:i], frames[i+1:]...)
				case "premature-message-stop":
					i, j := find("message_delta"), find("message_stop")
					frames[i], frames[j] = frames[j], frames[i]
				case "after-terminal":
					frames = append(frames, frames[find("content_block_start")])
				case "event-type-mismatch":
					i := find("content_block_stop")
					frames[i] = strings.Replace(frames[i], "event: content_block_stop", "event: content_block_start", 1)
				}
				script := strings.Join(frames, "\n\n") + "\n\n"
				r = provider.SSE([][]byte{[]byte(script)})
				enqueue(ev, w, r)
				o := drain(w.client.Stream(ctxFor(t), textScope("glm07-sequence"), ai.Target{BindingID: "glm-messages", ModelID: model}, glmMessagesCapturedRequest(t, c), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("illegal sequence fails as a protocol error", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				failed, done := 0, 0
				for _, e := range o.events {
					if e.Type() == ai.EventError {
						failed++
					}
					if e.Type() == ai.EventDone {
						done++
					}
				}
				ev.Check("one error terminal and no retry", failed == 1 && done == 0 && len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "error=%d done=%d", failed, done)
			})
		}
	}
}

// Interleaving is a decoder stress script assembled from Messages event types,
// not evidence that the vendor emits parallel tool streams.
func TestGLMMessagesMultiIndex(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM07-multi-index-"+model)
			w := newGLMMessagesWorld(t, nil)
			frames := []string{
				`{"type":"message_start","message":{"id":"msg-multi","model":"` + model + `"}}`,
				`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"call-alpha","name":"lookup_probe_code","input":{}}}`,
				`{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"call-beta","name":"lookup_probe_code","input":{}}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"label\":"}}`,
				`{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"label\":\"alpha\"}"}}`,
				`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"alpha\"}"}}`,
				`{"type":"content_block_stop","index":1}`,
				`{"type":"content_block_stop","index":0}`,
				`{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"input_tokens":10,"output_tokens":5}}`,
				`{"type":"message_stop"}`,
			}
			var script strings.Builder
			for _, f := range frames {
				fields := bodyFields(t, []byte(f))
				script.WriteString("event: " + strings.Trim(string(fields["type"]), `"`) + "\ndata: " + f + "\n\n")
			}
			enqueue(ev, w, provider.SSE([][]byte{[]byte(script.String())}))
			req := glmMessagesCapturedRequest(t, glmMessagesCaptureByName(t, model+"-tool"))
			o := drain(w.client.Stream(ctxFor(t), textScope("glm07-multi-index"), ai.Target{BindingID: "glm-messages", ModelID: model}, req, nil))
			o.record(ev)
			ev.Check("both tool indexes survive fragmented interleaving", o.err == nil && len(o.result.Message.Content) == 2 && o.result.Message.StopReason == ai.StopReasonToolUse, "%v %+v", o.err, o.result.Message.Content)
			for i, id := range []string{"call-alpha", "call-beta"} {
				if len(o.result.Message.Content) <= i {
					break
				}
				call, err := ai.ValidateToolCall(o.result.Message, i, req.Tools)
				var args struct{ Label string }
				ev.Check("Host validates independent call with stable order and ID", err == nil && call.ID() == id && call.Decode(&args) == nil && args.Label == "alpha", "%v %+v", err, args)
			}
		})
	}
}
