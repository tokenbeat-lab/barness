package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestChatHistory is P04/E03: history conversion, reasoning replay, tool
// results and images (testdata/chat/history.json).
func TestChatHistory(t *testing.T) {
	f, raw := loadFixture(t, chatProtocol, "history.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P04-E03-"+sc.ID)
			runScenario(t, ev, sc, raw, modeStream)
		})
	}
}

// TestChatToolRoundTrip is P04's live tool round trip: the host validates
// the model's tool call, executes it and passes the first call's message
// straight into the second call, which replays the reasoning under the
// field it streamed in, the text and the call.
func TestChatToolRoundTrip(t *testing.T) {
	ev := run.Case(t, "P04-E03-live-tool-round-trip")
	tf, traw := loadFixture(t, chatProtocol, "text.json")
	hf, hraw := loadFixture(t, chatProtocol, "history.json")
	first := scenarioByID(t, tf, "interleaved")
	answer := scenarioByID(t, hf, "tool-round-trip")
	ev.Fixture("text.json", traw)
	ev.Fixture("history.json", hraw)
	w := scenarioWorld(t, first)
	enqueue(ev, w, first.replies(t)...)
	enqueue(ev, w, answer.replies(t)...)

	req := first.request(t)
	req.SystemPrompt = answer.Context.SystemPrompt
	res, err := w.client.Complete(ctxFor(t), textScope("req-round-1"), first.target(), req, ai.ChatOptions{})
	ev.Record("round-1", res)
	if !ev.Check("first round asks for a tool", err == nil && res.Message.StopReason == ai.StopReasonToolUse, "err=%v stop=%q", err, res.Message.StopReason) {
		return
	}
	call, err := ai.ValidateToolCall(res.Message, 2, req.Tools)
	if !ev.Check("the host validates the call", err == nil && call.ID() == "call_1", "err=%v", err) {
		return
	}
	var args struct {
		Q string
		N int
	}
	ev.Check("arguments decode", call.Decode(&args) == nil && args.Q == "x" && args.N == 2, "got %+v", args)

	req.Messages = append(req.Messages, res.Message, ai.ToolResultText(res.Message.Content[2].(ai.ToolCall), "42", false))
	res2, err := w.client.Complete(ctxFor(t), textScope("req-round-2"), first.target(), req, ai.ChatOptions{})
	ev.Record("round-2", res2)
	ev.Check("second round completes", err == nil && res2.Message.StopReason == ai.StopReasonStop, "err=%v", err)
	ev.Check("native state replayed, nothing downgraded", res2.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
		"got %+v", res2.Metadata.NativeStateDowngrades)

	reqs := w.provider.Requests()
	if !ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs)) {
		return
	}
	ev.Check("the second request is the stored round trip's", jsonEqual(reqs[1].Body, answer.Expect.Request.Body),
		"got %s\nwant %s", reqs[1].Body, answer.Expect.Request.Body)
}
