package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxMessagesSequence(t *testing.T) {
	for _, fault := range []string{"no-start", "duplicate-start", "duplicate-index", "unopened-delta", "wrong-delta", "open-terminal", "stop-before-delta", "after-terminal", "duplicate-terminal", "mismatched-type", "missing-index", "unknown-block", "missing-stop", "missing-stop-reason"} {
		t.Run(fault, func(t *testing.T) {
			ev := run.Case(t, "MM04-sequence-"+fault)
			w := newMiniMaxMessagesWorld(t, nil)
			frames := strings.Split(strings.TrimSpace(string(miniMaxMessagesFixture(t, "text.sse", "MiniMax-M3"))), "\n\n")
			find := func(kind string) int {
				for i, f := range frames {
					if strings.HasPrefix(f, "event: "+kind+"\n") {
						return i
					}
				}
				panic(kind)
			}
			switch fault {
			case "no-start":
				frames = frames[1:]
			case "duplicate-start":
				frames = append([]string{frames[0]}, frames...)
			case "duplicate-index":
				i := find("content_block_start")
				frames = append(frames[:i], append([]string{frames[i]}, frames[i:]...)...)
			case "unopened-delta":
				i := find("content_block_delta")
				frames[i] = strings.ReplaceAll(frames[i], `"index":0`, `"index":99`)
			case "wrong-delta":
				i := find("content_block_delta")
				frames[i] = strings.ReplaceAll(frames[i], "text_delta", "thinking_delta")
			case "open-terminal":
				i := find("content_block_stop")
				frames = append(frames[:i], frames[i+1:]...)
			case "stop-before-delta":
				i, j := find("message_delta"), find("message_stop")
				frames[i], frames[j] = frames[j], frames[i]
			case "after-terminal":
				frames = append(frames, frames[find("content_block_start")])
			case "duplicate-terminal":
				frames = append(frames, frames[find("message_stop")])
			case "mismatched-type":
				i := find("content_block_stop")
				frames[i] = strings.Replace(frames[i], "event: content_block_stop", "event: content_block_start", 1)
			case "missing-index":
				i := find("content_block_delta")
				frames[i] = strings.ReplaceAll(frames[i], `"index":0,`, "")
			case "unknown-block":
				i := find("content_block_start")
				frames[i] = strings.ReplaceAll(frames[i], `"type":"text"`, `"type":"opaque"`)
			case "missing-stop":
				frames = frames[:len(frames)-1]
			case "missing-stop-reason":
				i := find("message_delta")
				frames[i] = strings.ReplaceAll(frames[i], `"stop_reason":"end_turn"`, `"stop_reason":null`)
			}
			enqueue(ev, w, provider.SSE([][]byte{[]byte(strings.Join(frames, "\n\n") + "\n\n")}))
			o := drain(w.client.Stream(ctxFor(t), textScope("mm04-sequence-"+fault), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("malformed sequence has a diagnostic protocol failure", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
			failed, done := 0, 0
			for _, e := range o.events {
				if e.Type() == ai.EventError {
					failed++
				}
				if e.Type() == ai.EventDone {
					done++
				}
			}
			ev.Check("one failure terminal with no replay", failed == 1 && done == 0 && len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "errors=%d done=%d", failed, done)
		})
	}
}
