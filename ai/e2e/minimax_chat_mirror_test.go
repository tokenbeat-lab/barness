package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxChatObservedMirror(t *testing.T) {
	for _, framing := range []string{"frames", "bytewise"} {
		t.Run(framing, func(t *testing.T) {
			ev := run.Case(t, "MM10-actual-chat-mirror-"+framing)
			first, err := os.ReadFile("testdata/minimax-observed/chat-mirror-text-first-frame.sse")
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("actual-first-frame.sse", first)
			tail := []json.RawMessage{
				json.RawMessage(`{"choices":[{"delta":{"reasoning_content":" answer","reasoning_details":[{"type":"reasoning.text","id":"reasoning-text-1","format":"MiniMax-response-v1","index":0,"text":" answer"}]}}]}`),
				json.RawMessage(`{"choices":[{"delta":{"content":"Ready."},"finish_reason":"stop"}]}`),
				json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2},"completion_tokens_details":{"reasoning_tokens":3}}}`),
			}
			ev.Fixture("synthetic-completion-frames.json", mustMarshal(t, tail))
			reply := chatEvents(t, tail, "")
			chunks := append([][]byte{first}, reply.Chunks...)
			if framing == "bytewise" {
				raw := append(append([]byte{}, first...), []byte(reply.Script)...)
				chunks = nil
				for _, b := range raw {
					chunks = append(chunks, []byte{b})
				}
			}
			w := newMiniMaxChatWorld(t, nil)
			enqueue(ev, w, provider.SSE(chunks))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			visible, deltas := "", ""
			for _, b := range o.result.Message.Content {
				if v, ok := b.(ai.Thinking); ok {
					visible += v.Thinking
				}
			}
			for _, e := range o.events {
				if v, ok := e.(ai.ThinkingDeltaEvent); ok {
					deltas += v.Delta
				}
			}
			ev.Check("observed equal mirror streams once", o.err == nil && visible == "The answer" && deltas == "The answer", "%v visible=%q deltas=%q", o.err, visible, deltas)
			ev.Check("one completed stream and late known usage", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1 && o.result.Metadata.Attempts[0].UsageReporting == ai.UsageComplete && o.result.Message.Usage.Input == 10 && o.result.Message.Usage.Output == 8, "%+v", o.result)
			ev.Check("answer remains independent", strings.Contains(string(mustMarshal(t, o.result.Message.Content)), "Ready."), "%+v", o.result.Message)
		})
	}
}
