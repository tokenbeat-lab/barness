package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/examples/toolloop"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// TestToolLoopExample runs the tool round trip example (E10 工具往返, spec I6,
// User Stories 15–17, 49): the host validates the model's calls, runs them
// itself and starts the next turn as a new logical call with a new
// RequestID; calls of a turn that did not end with tool use, or that fail
// validation, are never run.
func TestToolLoopExample(t *testing.T) {
	f, raw := loadToolFixture(t)

	t.Run("responses", func(t *testing.T) {
		ev := run.Case(t, "P01-E10-example-tool-loop")
		ev.Fixture("tool-round-trip.json", raw)
		w := newWorld(t, tenantA)
		enqueue(ev, w, f.reply(t, f.Round1.Events, "", ""), f.reply(t, f.Round2.Events, "", ""))
		ran := &toolRuns{}
		turns, err := loopFor(w, f.target(), 4).Run(ctxFor(t), toolLoopRequest(f), ran.tools(f.Tools, f.ToolResults))
		recordTurns(ev, turns, err, ran)
		ev.Check("the loop ends with the model's answer", err == nil && len(turns) == 2 && last(turns).Message.StopReason == ai.StopReasonStop,
			"err=%v turns=%d", err, len(turns))
		checkTurnsAreCalls(ev, turns)
		ev.Check("each tool ran once with its validated arguments", ran.equal([]string{
			`get_weather {"city":"Zürich","unit":"celsius"}`, `get_time {"tz":"Europe/Zurich"}`}), "ran %q", ran.all())
		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs)) {
			ev.Check("turn 1 sends the question and tools", jsonEqual(reqs[0].Body, f.Round1.Expect.Body), "got %s", reqs[0].Body)
			ev.Check("turn 2 replays the calls with the host's results", jsonEqual(reqs[1].Body, f.Round2.Expect.Body), "got %s", reqs[1].Body)
		}
	})

	t.Run("anthropic", func(t *testing.T) {
		ev := run.Case(t, "P02-E10-example-tool-loop")
		tf, traw := loadFixture(t, anthropicProtocol, "text.json")
		hf, hraw := loadFixture(t, anthropicProtocol, "history.json")
		ev.Fixture("text.json", traw)
		ev.Fixture("history.json", hraw)
		first, answer := scenarioByID(t, tf, "interleaved"), scenarioByID(t, hf, "tool-round-trip")
		w := scenarioWorld(t, first)
		enqueue(ev, w, first.replies(t)...)
		enqueue(ev, w, answer.replies(t)...)
		ran := &toolRuns{}
		req := first.request(t)
		turns, err := loopFor(w, first.target(), 4).Run(ctxFor(t), req, ran.tools(req.Tools, map[string]string{"lookup": "42"}))
		recordTurns(ev, turns, err, ran)
		ev.Check("the loop ends with the model's answer", err == nil && len(turns) == 2 && last(turns).Message.StopReason == ai.StopReasonStop,
			"err=%v turns=%d", err, len(turns))
		checkTurnsAreCalls(ev, turns)
		ev.Check("the tool ran once with its validated arguments", ran.equal([]string{`lookup {"q":"x"}`}), "ran %q", ran.all())
		ev.Check("the signed turn replays natively", last(turns).Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
			"got %+v", last(turns).Metadata.NativeStateDowngrades)
		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if !ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs)) {
			return
		}
		var body struct {
			Messages []struct {
				Content []json.RawMessage `json:"content"`
			} `json:"messages"`
		}
		mustUnmarshal(t, reqs[1].Body, &body)
		ev.Check("turn 2 answers the call with the host's result", len(body.Messages) == 3 && jsonEqual(body.Messages[2].Content[0],
			[]byte(`{"type":"tool_result","tool_use_id":"toolu_01","content":"42","is_error":false,"cache_control":{"type":"ephemeral"}}`)),
			"got %s", reqs[1].Body)
	})

	for _, tr := range f.Truncated {
		t.Run("not-run/"+tr.ID, func(t *testing.T) {
			// A turn cut off by the output limit or a failed stream holds
			// calls that look complete; the host never runs them.
			ev := run.Case(t, "P01-E10-example-tool-loop-not-run-"+tr.ID)
			ev.Fixture("tool-round-trip.json", raw)
			w := newWorld(t, tenantA)
			enqueue(ev, w, f.reply(t, tr.Events, "", tr.End))
			ran := &toolRuns{}
			turns, err := loopFor(w, f.target(), 4).Run(ctxFor(t), toolLoopRequest(f), ran.tools(f.Tools, f.ToolResults))
			recordTurns(ev, turns, err, ran)
			ev.Check("the loop stops after the turn", len(turns) == 1 && last(turns).Message.StopReason == tr.Expect.StopReason,
				"turns=%d", len(turns))
			ev.Check("the turn's error is returned", (err != nil) == (tr.Expect.StopReason == ai.StopReasonError), "err=%v", err)
			ev.Check("no tool ran", len(ran.all()) == 0, "ran %q", ran.all())
			ev.Check("no next turn was sent", len(w.provider.Requests()) == 1, "got %d", len(w.provider.Requests()))
		})
	}

	t.Run("invalid-arguments-answered", func(t *testing.T) {
		// The host declares get_weather with a required "country" the
		// model's call lacks: the call is answered with pi's validation
		// message as an error result and never run; the valid call runs.
		ev := run.Case(t, "P01-E10-example-tool-loop-invalid-arguments")
		ev.Fixture("tool-round-trip.json", raw)
		w := newWorld(t, tenantA)
		enqueue(ev, w, f.reply(t, f.Round1.Events, "", ""), f.reply(t, f.Round2.Events, "", ""))
		tools := append([]ai.Tool(nil), f.Tools...)
		tools[0].Parameters = json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"},"unit":{"type":"string"},"country":{"type":"string"}},"required":["city","country"]}`)
		ran := &toolRuns{}
		turns, err := loopFor(w, f.target(), 4).Run(ctxFor(t), toolLoopRequest(f), ran.tools(tools, f.ToolResults))
		recordTurns(ev, turns, err, ran)
		ev.Check("the loop goes on to the answer", err == nil && len(turns) == 2, "err=%v turns=%d", err, len(turns))
		ev.Check("only the valid call ran", ran.equal([]string{`get_time {"tz":"Europe/Zurich"}`}), "ran %q", ran.all())
		if len(turns) > 0 && ev.Check("both calls are answered", len(turns[0].Results) == 2, "got %d", len(turns[0].Results)) {
			refused := turns[0].Results[0]
			text := refused.Content[0].(ai.Text).Text
			ev.Check("the invalid call is answered as an error with the validation message", refused.IsError &&
				refused.ToolCallID == "call_w1|fc_w1" && strings.HasPrefix(text, "Validation failed for tool \"get_weather\""), "got %+v", refused)
		}
	})

	t.Run("max-turns", func(t *testing.T) {
		ev := run.Case(t, "P01-E10-example-tool-loop-max-turns")
		ev.Fixture("tool-round-trip.json", raw)
		w := newWorld(t, tenantA)
		enqueue(ev, w, f.reply(t, f.Round1.Events, "", ""), f.reply(t, f.Round1.Events, "", ""))
		ran := &toolRuns{}
		turns, err := loopFor(w, f.target(), 2).Run(ctxFor(t), toolLoopRequest(f), ran.tools(f.Tools, f.ToolResults))
		recordTurns(ev, turns, err, ran)
		ev.Check("the loop stops at MaxTurns", errors.Is(err, toolloop.ErrTooManyTurns) && len(turns) == 2, "err=%v turns=%d", err, len(turns))
		ev.Check("no request beyond MaxTurns", len(w.provider.Requests()) == 2, "got %d", len(w.provider.Requests()))
	})
}

// loopFor is the example loop on w's Client for tenant A.
func loopFor(w *world, target ai.Target, maxTurns int) toolloop.Loop {
	return toolloop.Loop{Client: w.client, Target: target, MaxTurns: maxTurns, Scope: func() ai.CallScope {
		return ai.CallScope{TenantID: tenantA.tenant, ActorID: "actor-7", RequestID: requestid.New()}
	}}
}

// toolLoopRequest is the fixture's first-round input; the loop declares the
// tools.
func toolLoopRequest(f toolFixture) ai.Request {
	return ai.Request{SystemPrompt: f.SystemPrompt, Messages: []ai.Message{ai.UserText(f.User)}}
}

// toolRuns records the tools the host ran.
type toolRuns struct {
	mu  sync.Mutex
	ran []string
}

// tools offers declarations whose runs answer with results[name].
func (r *toolRuns) tools(declarations []ai.Tool, results map[string]string) []toolloop.Tool {
	out := make([]toolloop.Tool, len(declarations))
	for i, d := range declarations {
		out[i] = toolloop.Tool{Declaration: d, Run: func(_ context.Context, call ai.ValidToolCall) (string, error) {
			r.mu.Lock()
			r.ran = append(r.ran, call.Name()+" "+string(call.Arguments()))
			r.mu.Unlock()
			return results[call.Name()], nil
		}}
	}
	return out
}

func (r *toolRuns) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.ran...)
}

func (r *toolRuns) equal(want []string) bool { return slices.Equal(r.all(), want) }

func last(turns []toolloop.Turn) ai.Result {
	if len(turns) == 0 {
		return ai.Result{}
	}
	return turns[len(turns)-1].Result
}

func recordTurns(ev *evidence.Case, turns []toolloop.Turn, err error, ran *toolRuns) {
	ev.Record("turns", turns)
	ev.Record("error", errString(err))
	ev.Record("tools-run", ran.all())
}

// checkTurnsAreCalls asserts each turn was its own logical call of the same
// tenant, with a RequestID never used before.
func checkTurnsAreCalls(ev *evidence.Case, turns []toolloop.Turn) {
	seen := map[string]bool{}
	for i, turn := range turns {
		m := turn.Result.Metadata
		ev.Check("turn is tenant A's resolved call", m.TenantID == tenantA.tenant && m.Resolved, "turn %d: %+v", i, m.CallAttribution)
		ev.Check("turn has a new RequestID", m.RequestID != "" && !seen[m.RequestID], "turn %d: %q", i, m.RequestID)
		seen[m.RequestID] = true
	}
}
