package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestAnthropicHistory is P02/E03: history conversion, native state replay,
// tool results and images (testdata/anthropic/history.json), and a live tool
// round trip in which the host validates the model's call, executes it and
// passes the first call's message straight into the second call.
func TestAnthropicHistory(t *testing.T) {
	f, raw := loadAnthropicFixture(t, "history.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P02-E03-"+sc.ID)
			runAnthropic(t, ev, sc, raw, modeStream)
		})
	}

	t.Run("live-tool-round-trip", func(t *testing.T) {
		ev := run.Case(t, "P02-E03-live-tool-round-trip")
		tf, traw := loadAnthropicFixture(t, "text.json")
		first := anthropicScenarioByID(t, tf, "interleaved")
		answer := anthropicScenarioByID(t, f, "tool-round-trip")
		ev.Fixture("text.json", traw)
		ev.Fixture("history.json", raw)
		w := anthropicWorld(t, first)
		enqueue(ev, w, first.replies(t)...)
		enqueue(ev, w, answer.replies(t)...)

		req := first.request(t)
		res, err := w.client.Complete(ctxFor(t), textScope("req-round-1"), first.target(), req, ai.AnthropicOptions{})
		ev.Record("round-1", res)
		if !ev.Check("first round asks for a tool", err == nil && res.Message.StopReason == ai.StopReasonToolUse, "err=%v stop=%q", err, res.Message.StopReason) {
			return
		}
		call, err := ai.ValidateToolCall(res.Message, 2, req.Tools)
		if !ev.Check("the host validates the call", err == nil && call.ID() == "toolu_01", "err=%v", err) {
			return
		}
		var args struct{ Q string }
		ev.Check("arguments decode", call.Decode(&args) == nil && args.Q == "x", "got %+v", args)

		req.Messages = append(req.Messages, res.Message,
			ai.ToolResultText(res.Message.Content[2].(ai.ToolCall), "42", false))
		res2, err := w.client.Complete(ctxFor(t), textScope("req-round-2"), first.target(), req, ai.AnthropicOptions{})
		ev.Record("round-2", res2)
		ev.Check("second round completes", err == nil && res2.Message.StopReason == ai.StopReasonStop, "err=%v", err)
		ev.Check("native state replayed, nothing downgraded", res2.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
			"got %+v", res2.Metadata.NativeStateDowngrades)

		reqs := w.provider.Requests()
		if !ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs)) {
			return
		}
		var body struct {
			Messages []struct {
				Role    string            `json:"role"`
				Content []json.RawMessage `json:"content"`
			} `json:"messages"`
		}
		mustUnmarshal(t, reqs[1].Body, &body)
		if !ev.Check("user, assistant, tool result", len(body.Messages) == 3, "got %s", reqs[1].Body) {
			return
		}
		want := []string{
			`{"type":"thinking","thinking":"Let me think","signature":"sig-abc"}`,
			`{"type":"text","text":"Hi there"}`,
			`{"type":"tool_use","id":"toolu_01","name":"lookup","input":{"q":"x"}}`,
		}
		got := body.Messages[1].Content
		ok := len(got) == len(want)
		for i := 0; ok && i < len(want); i++ {
			ok = jsonEqual(got[i], []byte(want[i]))
		}
		ev.Check("the turn is replayed with its signed thinking and call", ok, "got %s", mustMarshal(t, got))
		ev.Check("the result answers the call", jsonEqual(body.Messages[2].Content[0],
			[]byte(`{"type":"tool_result","tool_use_id":"toolu_01","content":"42","is_error":false,"cache_control":{"type":"ephemeral"}}`)),
			"got %s", body.Messages[2].Content[0])
	})
}
