package e2e

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestToolRoundTrip is the local tool round trip (spec I6 tools, User Stories
// 15–17, P01 "工具调用→宿主结果→下一轮", E03 tool part). The model calls two tools
// with argument JSON split mid-token across deltas; a test host validates
// each call with the public helper, executes it itself and sends the results
// in a second logical call with a new RequestID; the model answers in text.
// barness-ai only converts protocol data: it never executes a tool.
func TestToolRoundTrip(t *testing.T) {
	f, raw := loadToolFixture(t)
	for _, e := range toolEntries {
		t.Run(e.name, func(t *testing.T) {
			ev := run.Case(t, "P01-E03-tool-round-trip-"+e.name)
			ev.Fixture("tool-round-trip.json", raw)
			w := newWorld(t, tenantA)
			enqueue(ev, w, f.reply(t, f.Round1.Events, provider.FramingLF, ""), f.reply(t, f.Round2.Events, provider.FramingLF, ""))

			// Round 1: the model asks for tools.
			scope1 := textScope("req-tools-1-" + e.name)
			req1 := f.request()
			o1 := within(ev, "round 1 ends", func() outcome { return e.invoke(ctxFor(t), w, f, scope1, req1, true) })
			o1.record(ev)
			checkToolRound(ev, "round 1", f.Round1, o1)
			reqs := w.provider.Requests()
			if ev.Check("round 1 sent one request", len(reqs) == 1, "got %d", len(reqs)) {
				ev.Check("round 1 body: tools and tool_choice", jsonEqual(reqs[0].Body, withToolChoice(t, f.Round1.Expect.Body, e.toolChoice(f))),
					"got %s", reqs[0].Body)
			}

			// The host validates each call before it would execute it.
			assistant := o1.result.Message
			for i, block := range assistant.Content {
				call, ok := block.(ai.ToolCall)
				if !ev.Check("round 1 content is tool calls", ok, "block %d is %T", i, block) {
					return
				}
				valid, err := ai.ValidateToolCall(assistant, i, f.Tools)
				if !ev.Check("call "+call.Name+" is executable", err == nil, "err=%v", err) {
					return
				}
				var args map[string]string
				ev.Check("validated arguments decode", valid.Decode(&args) == nil && len(args) > 0, "got %s", valid.Arguments())
				ev.Check("validated call keeps its identity", valid.ID() == call.ID && valid.Name() == call.Name, "got %q %q", valid.ID(), valid.Name())
			}

			// Round 2: a new logical call carries a result for each call.
			scope2 := textScope("req-tools-2-" + e.name)
			req2 := f.round2Request(assistant)
			o2 := within(ev, "round 2 ends", func() outcome { return e.invoke(ctxFor(t), w, f, scope2, req2, false) })
			o2.record(ev)
			checkToolRound(ev, "round 2", f.Round2, o2)
			reqs = w.provider.Requests()
			if ev.Check("two requests in all", len(reqs) == 2, "got %d", len(reqs)) {
				ev.Check("round 2 body: replayed calls and results", jsonEqual(reqs[1].Body, f.Round2.Expect.Body), "got %s", reqs[1].Body)
				ev.Check("both rounds use the tenant's key", reqs[0].KeyAlias == tenantA.alias && reqs[1].KeyAlias == tenantA.alias,
					"got %q %q", reqs[0].KeyAlias, reqs[1].KeyAlias)
			}
			ev.Check("each round is its own logical call", o1.result.Metadata.RequestID == scope1.RequestID &&
				o2.result.Metadata.RequestID == scope2.RequestID && scope1.RequestID != scope2.RequestID,
				"got %q %q", o1.result.Metadata.RequestID, o2.result.Metadata.RequestID)
			ev.Check("round 1 message unchanged by the replay", reflect.DeepEqual(req2.Messages[1], assistant), "")
		})
	}

	// Argument fragments split across SSE frames (down to single bytes,
	// through the multi-byte ü) join into the same calls.
	for _, framing := range provider.Framings {
		t.Run("framing-"+string(framing), func(t *testing.T) {
			ev := run.Case(t, "P01-E03-tool-call-framing-"+string(framing))
			ev.Fixture("tool-round-trip.json", raw)
			w := newWorld(t, tenantA)
			enqueue(ev, w, f.reply(t, f.Round1.Events, framing, ""))
			o := within(ev, "call ends", func() outcome {
				return toolEntries[0].invoke(ctxFor(t), w, f, textScope("req-tools-framing-"+string(framing)), f.request(), true)
			})
			o.record(ev)
			checkToolRound(ev, "round 1", f.Round1, o)
		})
	}

	// The full entry forces one function with the object form of tool_choice.
	t.Run("forced-function", func(t *testing.T) {
		ev := run.Case(t, "P01-E03-tool-choice-function")
		ev.Fixture("tool-round-trip.json", raw)
		w := newWorld(t, tenantA)
		enqueue(ev, w, f.reply(t, f.Round1.Events, provider.FramingLF, ""))
		opts := ai.ResponsesOptions{ToolChoice: &ai.ResponsesToolChoice{Function: "get_weather"}}
		o := within(ev, "call ends", func() outcome {
			return drain(w.client.Stream(ctxFor(t), textScope("req-tools-function"), f.target(), f.request(), opts))
		})
		o.record(ev)
		checkToolRound(ev, "round 1", f.Round1, o)
		var body map[string]any
		reqs := w.provider.Requests()
		if ev.Check("one request", len(reqs) == 1, "got %d", len(reqs)) && ev.Check("body decodes", json.Unmarshal(reqs[0].Body, &body) == nil, "") {
			ev.Check("tool_choice names the function", reflect.DeepEqual(body["tool_choice"], map[string]any{"type": "function", "name": "get_weather"}),
				"got %v", body["tool_choice"])
		}
	})

	// A live view is a running message: never executable.
	t.Run("live-view-not-executable", func(t *testing.T) {
		ev := run.Case(t, "P01-E03-tool-call-live-view")
		ev.Fixture("tool-round-trip.json", raw)
		w := newWorld(t, tenantA)
		enqueue(ev, w, f.reply(t, f.Round1.Events, provider.FramingLF, ""))
		s := w.client.Stream(ctxFor(t), textScope("req-tools-live"), f.target(), f.request(), nil)
		defer s.Close()
		checked := false
		for s.Next() {
			if d, ok := s.Event().(ai.ToolCallDeltaEvent); ok && !checked {
				checked = true
				_, err := ai.ValidateToolCall(d.Partial.Snapshot(), d.ContentIndex, f.Tools)
				ev.Check("live snapshot refused", errors.Is(err, &ai.ToolCallError{Kind: ai.ToolCallNotExecutable}), "err=%v", err)
			}
		}
		ev.Check("a delta was seen", checked, "")
	})
}

// TestToolCallTruncation: a call in a turn cut off by the output limit or a
// failed stream keeps what was received for display, and the helper never
// calls it executable (spec I8 "正常 length 仍是明确截断终态，不等于完整工具参数",
// E02 "截断工具不被执行").
func TestToolCallTruncation(t *testing.T) {
	f, raw := loadToolFixture(t)
	for _, sc := range f.Truncated {
		for _, e := range toolEntries {
			t.Run(sc.ID+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E02-tool-"+sc.ID+"-"+e.name)
				ev.Fixture("tool-round-trip.json", raw)
				w, bodies := newTrackedWorld(t)
				enqueue(ev, w, f.reply(t, sc.Events, provider.FramingLF, sc.End))
				o := within(ev, "call ends", func() outcome {
					return e.invoke(ctxFor(t), w, f, textScope("req-trunc-"+sc.ID+"-"+e.name), f.request(), false)
				})
				o.record(ev)
				m := o.result.Message
				ev.Check("stop reason", m.StopReason == sc.Expect.StopReason, "got %q want %q", m.StopReason, sc.Expect.StopReason)
				ev.Check("received call kept for display", reflect.DeepEqual(contentSummary(m.Content), sc.Expect.Content),
					"got %q\nwant %q", contentSummary(m.Content), sc.Expect.Content)
				if sc.Expect.Code != "" {
					ev.Check("classified failure", errors.Is(o.err, &ai.Error{Code: sc.Expect.Code}), "got %v", o.err)
				} else {
					ev.Check("truncation is not a failure", o.err == nil, "got %v", o.err)
				}
				if o.streamed {
					ev.Check("event sequence", reflect.DeepEqual(eventSummary(o.events), sc.Expect.Events), "got %q\nwant %q", eventSummary(o.events), sc.Expect.Events)
				}
				_, err := ai.ValidateToolCall(m, 0, f.Tools)
				ev.Check("truncated call is not executable", errors.Is(err, &ai.ToolCallError{Kind: ai.ToolCallNotExecutable}), "err=%v", err)
				ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
			})
		}
	}
}

// checkToolRound asserts one round's outcome against the fixture.
func checkToolRound(ev *evidence.Case, round string, r toolRound, o outcome) {
	x := r.Expect
	m := o.result.Message
	ev.Check(round+": success", o.err == nil, "err=%v", o.err)
	ev.Check(round+": stop reason", m.StopReason == x.StopReason, "got %q want %q", m.StopReason, x.StopReason)
	ev.Check(round+": content", reflect.DeepEqual(contentSummary(m.Content), x.Content), "got %q\nwant %q", contentSummary(m.Content), x.Content)
	ev.Check(round+": response id", m.ResponseID == x.ResponseID, "got %q", m.ResponseID)
	ev.Check(round+": usage", m.Usage == x.Usage, "got %+v want %+v", m.Usage, x.Usage)
	if o.streamed {
		ev.Check(round+": event sequence", reflect.DeepEqual(eventSummary(o.events), x.Events), "got %q\nwant %q", eventSummary(o.events), x.Events)
		ev.Check(round+": Err is nil", o.streamErr == nil, "got %v", o.streamErr)
		for _, e := range o.events {
			if end, ok := e.(ai.ToolCallEndEvent); ok {
				ev.Check(round+": toolcall_end carries the final call", reflect.DeepEqual(m.Content[end.ContentIndex], end.ToolCall),
					"event=%+v final=%+v", end.ToolCall, m.Content[end.ContentIndex])
			}
		}
	}
}
