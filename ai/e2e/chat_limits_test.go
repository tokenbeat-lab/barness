package e2e

// The suites in this file parallel their Anthropic counterparts case for
// case on the shared helpers, with Chat's data and its differences
// (ADR-0013); sharing one protocol-parametric suite is issue 30.

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// chatEntries are the four public entry points on the "chat" binding, the
// full ones with Chat options.
var chatEntries = scenarioEntries(ai.ChatOptions{})

// chatTarget is the Chat model the generated scripts below stream as.
var chatTarget = ai.Target{BindingID: "chat", ModelID: "gpt-4.1-mini"}

// TestChatByteLimits is TestByteLimits on OpenAI Chat Completions (P04/E08,
// spec I9, ADR-0002, ADR-0007): the request body, SSE frame (LF and CRLF),
// streamed output, tool call argument JSON and error body at their boundary
// and at limit+1 on every entry point. The adapter decodes the stream
// itself through barness's body wrapper (ADR-0013), so the frame and output
// bounds are checked there, below the SDK's own 32 MiB line limit.
func TestChatByteLimits(t *testing.T) {
	a := loadScenarioText(t, chatProtocol)
	chunks := a.chunks(t)
	largestFrame, streamBytes := largestAndTotal(chunks)
	script := func(framed [][]byte) func(*testing.T, *evidence.Case, limitWorld, bool) {
		return func(_ *testing.T, ev *evidence.Case, lw limitWorld, _ bool) {
			ev.Fixture("text.json", a.raw)
			enqueue(ev, lw.world, provider.SSE(framed))
		}
	}
	within := func(ev *evidence.Case, _ limitWorld, o outcome) { checkScenarioTextSucceeded(ev, a, o) }
	keepsReceived := func(ev *evidence.Case, _ limitWorld, o outcome) {
		got := textOf(o.result.Message)
		ev.Check("message keeps the text received before the limit", strings.HasPrefix(a.text(), got), "got %q", got)
	}
	nothingSent := func(ev *evidence.Case, lw limitWorld, o outcome) {
		ev.Check("no inference request sent", len(lw.provider.Requests()) == 0, "got %d", len(lw.provider.Requests()))
		ev.Check("no attempt recorded", len(o.result.Metadata.Attempts) == 0, "got %+v", o.result.Metadata.Attempts)
	}
	boundaryOn := func(bc boundary) boundary {
		bc.entries, bc.casePrefix = chatEntries, "P04-E08"
		return bc
	}

	t.Run("request-body", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "request-body", field: "MaxRequestBytes", phase: ai.PhaseRequest, set: setRequestBytes,
			size: func(t *testing.T, e outcomeEntry) int64 {
				w := newWorld(t, tenantA)
				w.provider.Enqueue(a.reply(t))
				if o := e.invoke(ctxFor(t), w, textScope("req-p04-measure-"+e.name), a.target(), a.request(t)); o.err != nil {
					t.Fatalf("measuring call failed: %v", o.err)
				}
				return int64(len(w.provider.Requests()[0].Body))
			},
			script: script(chunks), target: a.target(), request: a.request(t), within: within, over: nothingSent,
		}))
	})

	t.Run("frame", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "frame", field: "MaxFrameBytes", phase: ai.PhaseStream, set: setFrameBytes, size: fixedSize(largestFrame),
			script: script(chunks), target: a.target(), request: a.request(t), within: within, over: keepsReceived,
		}))
	})

	crlf, err := provider.EncodeChatSSE(a.sc.Events, provider.FramingCRLF)
	if err != nil {
		t.Fatal(err)
	}
	largestCRLF, _ := largestAndTotal(crlf)
	t.Run("frame-crlf", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "frame-crlf", field: "MaxFrameBytes", phase: ai.PhaseStream, set: setFrameBytes, size: fixedSize(largestCRLF),
			script: script(crlf), target: a.target(), request: a.request(t), within: within, over: keepsReceived,
		}))
	})

	t.Run("output", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "output", field: "MaxOutputBytes", phase: ai.PhaseStream, set: setOutputBytes, size: fixedSize(streamBytes),
			script: script(chunks), target: a.target(), request: a.request(t), within: within, over: keepsReceived,
		}))
	})

	tf, traw := loadFixture(t, chatProtocol, "text.json")
	tool := scenarioByID(t, tf, "interleaved")
	const toolArgs = `{"q":"x","n":2}` // the call's argument JSON, from its streamed fragments
	t.Run("tool-json", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "tool-json", field: "MaxToolJSONBytes", phase: ai.PhaseStream, set: setToolJSONBytes, size: fixedSize(len(toolArgs)),
			script: func(t *testing.T, ev *evidence.Case, lw limitWorld, _ bool) {
				ev.Fixture("text.json", traw)
				enqueue(ev, lw.world, tool.replies(t)...)
			},
			target: tool.target(), request: tool.request(t),
			within: func(ev *evidence.Case, _ limitWorld, o outcome) {
				m := o.result.Message
				ev.Check("call succeeds with the tool call", o.err == nil && m.StopReason == ai.StopReasonToolUse, "err=%v stop=%q", o.err, m.StopReason)
				call, ok := m.Content[len(m.Content)-1].(ai.ToolCall)
				ev.Check("the call holds its arguments", ok && call.RawArguments == toolArgs, "got %+v", m.Content)
			},
			over: func(ev *evidence.Case, _ limitWorld, o outcome) {
				for _, c := range o.result.Message.Content {
					if call, ok := c.(ai.ToolCall); ok {
						ev.Check("no call holds more argument JSON than the limit", len(call.RawArguments) < len(toolArgs), "call holds %q", call.RawArguments)
					}
				}
			},
		}))
	})

	// Chat's failures have no JSON 429; its plain-text one is the limited
	// body here.
	ff, fraw := loadFixture(t, chatProtocol, "failures.json")
	limited := scenarioByID(t, ff, "http-429-text")
	limitedBody := limited.Replies[0].Body
	t.Run("error-body", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "error-body", field: "MaxErrorBodyBytes", phase: ai.PhaseRequest, set: setErrorBodyBytes, size: fixedSize(len(limitedBody)),
			script: func(t *testing.T, ev *evidence.Case, lw limitWorld, over bool) {
				ev.Fixture("failures.json", fraw)
				if over {
					// A limited error body ends the call; it is never retried.
					lw.updateBindingOf(tenantA, "chat", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 2} })
				}
				r := limited.replies(t)[0]
				enqueue(ev, lw.world, r, r, r)
			},
			target: limited.target(), request: limited.request(t),
			within: func(ev *evidence.Case, lw limitWorld, o outcome) {
				ev.Check("rate_limited with pi's message", errorsIsCode(o.err, ai.CodeRateLimited) &&
					o.result.Message.ErrorMessage == limited.Expect.ErrorMessage, "err=%v message=%q", o.err, o.result.Message.ErrorMessage)
				ev.Check("one request", len(lw.provider.Requests()) == 1, "got %d", len(lw.provider.Requests()))
			},
			over: func(ev *evidence.Case, lw limitWorld, o outcome) {
				ev.Check("the limited body is not retried", len(lw.provider.Requests()) == 1, "got %d", len(lw.provider.Requests()))
				ev.Check("the attempt is recorded as limited", len(o.result.Metadata.Attempts) == 1 &&
					o.result.Metadata.Attempts[0].Code == ai.CodeResourceLimit && o.result.Metadata.Attempts[0].HTTPStatus == 429,
					"got %+v", o.result.Metadata.Attempts)
			},
		}))
	})

	t.Run("held", func(t *testing.T) { testChatHeldOversize(t) })
}

// chatChunk is one streamed Chat Completions chunk of the generated scripts
// below with the given choice delta and finish reason.
func chatChunk(t *testing.T, delta map[string]any, finish any) json.RawMessage {
	t.Helper()
	return mustMarshal(t, map[string]any{"id": "chatcmpl-001", "object": "chat.completion.chunk", "created": 1760000000,
		"model": "gpt-4.1-mini-2025-04-14", "usage": nil,
		"choices": []any{map[string]any{"index": 0, "delta": delta, "logprobs": nil, "finish_reason": finish}}})
}

// testChatHeldOversize: a Chat provider exceeds a limit and then holds the
// connection without finishing; the call ends in time because the limit is
// checked while reading.
func testChatHeldOversize(t *testing.T) {
	start := "data: " + string(chatChunk(t, map[string]any{"content": "Hel"}, nil)) + "\n\n"
	heartbeats := strings.Repeat(": keep-alive\n\n", 200)
	cases := []struct {
		id, field string
		phase     ai.Phase
		set       func(*ai.ResourcePolicy)
		reply     provider.Reply
	}{
		{"frame", "MaxFrameBytes", ai.PhaseStream, func(p *ai.ResourcePolicy) { setFrameBytes(p, 1024) },
			heldReply(200, "text/event-stream", start, "data: "+strings.Repeat("x", 1100))},
		{"output", "MaxOutputBytes", ai.PhaseStream, func(p *ai.ResourcePolicy) { setOutputBytes(p, 2048) },
			heldReply(200, "text/event-stream", start, heartbeats)},
		{"error-body", "MaxErrorBodyBytes", ai.PhaseRequest, func(p *ai.ResourcePolicy) { setErrorBodyBytes(p, 1024) },
			heldReply(500, "application/json", `{"error":{"message":"`+strings.Repeat("e", 1100))},
	}
	for _, c := range cases {
		for _, e := range chatEntries {
			t.Run(c.id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P04-E08-held-"+c.id+"-"+e.name)
				lw := newLimitWorld(t, c.set)
				enqueue(ev, lw.world, c.reply)
				o := within(ev, "call ends while the provider still holds the connection", func() outcome {
					return e.invoke(ctxFor(t), lw.world, textScope("req-p04-held-"+c.id+"-"+e.name), chatTarget, ai.Request{Messages: []ai.Message{ai.UserText("hi")}})
				})
				o.record(ev)
				checkLimited(ev, o, c.phase, c.field)
				checkReleased(ev, lw)
			})
		}
	}
}

// chatLongOutput is a generated Chat stream of one text block in n content
// deltas of 9 bytes each, then the finish chunk, the usage chunk and
// [DONE], and the deltas.
func chatLongOutput(t *testing.T, n int) ([]json.RawMessage, []string) {
	t.Helper()
	var deltas []string
	var events []json.RawMessage
	for i := range n {
		d := fmt.Sprintf("part-%03d ", i)
		deltas = append(deltas, d)
		events = append(events, chatChunk(t, map[string]any{"content": d}, nil))
	}
	events = append(events,
		chatChunk(t, map[string]any{}, "stop"),
		json.RawMessage(`{"id":"chatcmpl-001","object":"chat.completion.chunk","created":1760000000,"model":"gpt-4.1-mini-2025-04-14","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":200,"total_tokens":210}}`),
		json.RawMessage(provider.ChatDone),
	)
	return events, deltas
}

// TestChatEventQueueLimits is TestEventQueueLimits on OpenAI Chat
// Completions (P04/E08): a Stream nobody reads is bounded by
// MaxQueuedEvents and MaxQueuedEventBytes, ends as resource_limit in the
// event_queue phase keeping its queued events in order, and Complete under
// the same policy finishes.
func TestChatEventQueueLimits(t *testing.T) {
	events, deltas := chatLongOutput(t, longOutputDeltas)
	full := longOutputEvents(deltas)
	queued := int64(len(full) - 1)
	eventBytes := int64(2 * 9 * len(deltas))
	req := ai.Request{Messages: []ai.Message{ai.UserText("Write a long answer.")}}
	starts := []streamEntry{
		{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) *ai.Stream {
			return w.client.Stream(ctx, scope, target, req, ai.ChatOptions{})
		}},
		streamEntries[1],
	}
	cases := []struct {
		id, field string
		set       func(*ai.ResourcePolicy)
		kept      int // events queued before the one that did not fit; 0 completes
	}{
		{"events-at-limit", "MaxQueuedEvents", func(p *ai.ResourcePolicy) { p.MaxQueuedEvents = int(queued) }, 0},
		{"events-over-limit", "MaxQueuedEvents", func(p *ai.ResourcePolicy) { p.MaxQueuedEvents = int(queued) - 1 }, int(queued) - 1},
		{"events-overflow-mid-deltas", "MaxQueuedEvents", func(p *ai.ResourcePolicy) { p.MaxQueuedEvents = 50 }, 50},
		{"bytes-at-limit", "MaxQueuedEventBytes", func(p *ai.ResourcePolicy) { p.MaxQueuedEventBytes = eventBytes }, 0},
		{"bytes-over-limit", "MaxQueuedEventBytes", func(p *ai.ResourcePolicy) { p.MaxQueuedEventBytes = eventBytes - 1 }, int(queued) - 1},
	}
	for _, c := range cases {
		for _, e := range starts {
			t.Run(c.id+"/result-only-"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P04-E08-queue-"+c.id+"-result-only-"+e.name)
				lw := newLimitWorld(t, c.set)
				enqueue(ev, lw.world, chatEvents(t, events, provider.FramingLF))
				s := e.start(ctxFor(t), lw.world, textScope("req-p04-queue-"+c.id+"-"+e.name), chatTarget, req)
				r := resultWithin(ev, "Result returns while no event is read", s)
				o := within(ev, "queued events drain after the call ended", func() outcome { return drain(s) })
				o.record(ev)
				got := eventSummary(o.events)
				ev.Check("Result and the drained stream agree", r.err == o.err && reflect.DeepEqual(r.res, o.result), "Result=%v drained=%v", r.err, o.err)
				ev.Check("the queue never held more than its limit plus the terminal", lw.probe.PeakQueuedEvents() <= int64(lw.policyEvents(c.set))+1,
					"peak %d", lw.probe.PeakQueuedEvents())
				if c.kept == 0 {
					ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
					ev.Check("every event is delivered", reflect.DeepEqual(got, full), "got %d events", len(got))
					checkReleased(ev, lw)
					return
				}
				checkLimited(ev, o, ai.PhaseEventQueue, c.field)
				ev.Check("the queued events are the call's first events, none dropped", len(got) == c.kept+1 &&
					reflect.DeepEqual(got[:c.kept], full[:c.kept]), "got %q", got)
				var delivered strings.Builder
				for _, e := range o.events {
					if d, ok := e.(ai.TextDeltaEvent); ok {
						delivered.WriteString(d.Delta)
					}
				}
				ev.Check("delivered deltas are a prefix of the final text", strings.HasPrefix(textOf(o.result.Message), delivered.String()),
					"text %q", textOf(o.result.Message))
				checkReleased(ev, lw)
			})
		}
		if c.kept == 0 {
			continue
		}
		t.Run(c.id+"/complete", func(t *testing.T) {
			ev := run.Case(t, "P04-E08-queue-"+c.id+"-complete")
			lw := newLimitWorld(t, c.set)
			enqueue(ev, lw.world, chatEvents(t, events, provider.FramingLF))
			r := within(ev, "Complete returns", func() callResult {
				res, err := lw.client.Complete(ctxFor(t), textScope("req-p04-queue-complete-"+c.id), chatTarget, req, ai.ChatOptions{})
				return callResult{res, err}
			})
			ev.Record("result", r.res)
			ev.Check("Complete succeeds under the same policy", r.err == nil, "err=%v", r.err)
			ev.Check("final text", textOf(r.res.Message) == strings.Join(deltas, ""), "got %q", textOf(r.res.Message))
			ev.Check("Complete queued no events", lw.probe.PeakQueuedEvents() == 0, "peak %d", lw.probe.PeakQueuedEvents())
			checkReleased(ev, lw)
		})
	}
}
