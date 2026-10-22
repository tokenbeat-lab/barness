package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiMessagesSequence(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, fault := range []string{"no-start", "duplicate-start", "duplicate-index", "unopened-delta", "open-terminal", "after-terminal", "duplicate-terminal", "mismatched-type", "missing-index", "missing-stop", "invalid-json", "tool-not-object", "incomplete-tool-json", "wrong-tool-stop"} {
			t.Run(string(market)+"/"+fault, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K12-"+fault)
				w := newKimiMessagesWorld(t, market, nil)
				script := kimiMessagesScript(`{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}`)
				frames := strings.Split(strings.TrimSpace(script), "\n\n")
				switch fault {
				case "no-start":
					frames = frames[1:]
				case "duplicate-start":
					frames = append([]string{frames[0]}, frames...)
				case "duplicate-index":
					frames = append(frames[:1], append([]string{frames[1]}, frames[1:]...)...)
				case "unopened-delta":
					frames = append(frames[:1], append([]string{`event: content_block_delta
data: {"type":"content_block_delta","index":99,"delta":{"type":"text_delta","text":"bad"}}`}, frames[1:]...)...)
				case "open-terminal":
					frames = append(frames[:2], frames[3:]...)
				case "after-terminal":
					frames = append(frames, frames[1])
				case "duplicate-terminal":
					frames = append(frames, frames[len(frames)-1])
				case "mismatched-type":
					frames[2] = strings.Replace(frames[2], "event: content_block_stop", "event: content_block_start", 1)
				case "missing-index":
					frames[2] = strings.ReplaceAll(frames[2], `"index":0,`, "")
				case "missing-stop":
					frames = frames[:len(frames)-1]
				case "invalid-json":
					frames[1] = "event: content_block_start\ndata: {invalid"
				case "tool-not-object", "incomplete-tool-json", "wrong-tool-stop":
					frames[1] = `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"opaque|/α","name":"inspect","input":{}}}`
					if fault == "tool-not-object" {
						frames[1] = strings.ReplaceAll(frames[1], `"input":{}`, `"input":[]`)
					}
					if fault == "incomplete-tool-json" {
						frames = append(frames[:2], append([]string{`event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"label\":"}}`}, frames[2:]...)...)
					}
				}
				enqueue(ev, w, provider.SSE([][]byte{[]byte(strings.Join(frames, "\n\n") + "\n\n")}))
				o := drain(w.client.Stream(ctxFor(t), textScope("kimi-sequence"), kimiMessagesTarget, kimiMessagesRequest(), nil))
				o.record(ev)
				ev.Check("invalid sequence produces diagnostic protocol failure", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				failed, done := 0, 0
				for _, e := range o.events {
					if e.Type() == ai.EventError {
						failed++
					}
					if e.Type() == ai.EventDone {
						done++
					}
				}
				ev.Check("one failure and no success terminal", failed == 1 && done == 0, "errors=%d done=%d", failed, done)
			})
		}
	}
}
