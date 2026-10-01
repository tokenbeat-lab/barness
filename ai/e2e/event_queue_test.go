package e2e

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// longOutputDeltas is how many text deltas the long-output script streams.
const longOutputDeltas = 200

// longOutput is a generated Responses stream of one text block in n deltas
// of 9 bytes each ("part-000 " …), and the deltas.
func longOutput(t *testing.T, n int) ([]json.RawMessage, []string) {
	t.Helper()
	var deltas []string
	events := []json.RawMessage{
		json.RawMessage(`{"type":"response.created","sequence_number":0,"response":{"id":"resp_long","object":"response","created_at":1790000000,"status":"in_progress","model":"gpt-4.1-mini-2025-04-14","output":[],"usage":null}}`),
		json.RawMessage(`{"type":"response.output_item.added","sequence_number":1,"output_index":0,"item":{"id":"msg_long","type":"message","status":"in_progress","role":"assistant","content":[]}}`),
	}
	for i := range n {
		d := fmt.Sprintf("part-%03d ", i)
		deltas = append(deltas, d)
		events = append(events, mustMarshal(t, map[string]any{"type": "response.output_text.delta", "sequence_number": 2 + i,
			"item_id": "msg_long", "output_index": 0, "content_index": 0, "delta": d}))
	}
	full := strings.Join(deltas, "")
	events = append(events,
		mustMarshal(t, map[string]any{"type": "response.output_item.done", "sequence_number": 2 + n, "output_index": 0,
			"item": map[string]any{"id": "msg_long", "type": "message", "status": "completed", "role": "assistant",
				"content": []any{map[string]any{"type": "output_text", "text": full, "annotations": []any{}}}}}),
		json.RawMessage(`{"type":"response.completed","sequence_number":999,"response":{"id":"resp_long","object":"response","created_at":1790000000,"status":"completed","model":"gpt-4.1-mini-2025-04-14","output":[],"usage":{"input_tokens":10,"input_tokens_details":{"cached_tokens":0},"output_tokens":200,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":210}}}`),
	)
	return events, deltas
}

// longOutputEvents is the full event summary barness publishes for the long
// output: n+3 queued events, then the terminal.
func longOutputEvents(deltas []string) []string {
	out := []string{"start", "text_start#0"}
	for _, d := range deltas {
		out = append(out, "text_delta#0:"+d)
	}
	return append(out, "text_end#0:"+strings.Join(deltas, ""), "done:stop")
}

// TestEventQueueLimits covers E08's event queue (spec I9, User Stories 33–34):
// a Stream whose events nobody reads is bounded by MaxQueuedEvents and
// MaxQueuedEventBytes; at the limit it completes, above it the call ends as
// resource_limit in the event_queue phase without blocking the producer or
// Result, the queued events are kept in order with nothing dropped between
// them, and the terminal still fits. Complete with the same output and
// policy finishes normally, as it queues nothing.
func TestEventQueueLimits(t *testing.T) {
	events, deltas := longOutput(t, longOutputDeltas)
	full := longOutputEvents(deltas)
	queued := int64(len(full) - 1)           // every event but the terminal
	eventBytes := int64(2 * 9 * len(deltas)) // the deltas, then text_end's content
	target := ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}
	req := ai.Request{Messages: []ai.Message{ai.UserText("Write a long answer.")}}

	cases := []struct {
		id    string
		field string
		set   func(*ai.ResourcePolicy)
		// kept is how many events are queued before the one that did not fit;
		// 0 means the call completes.
		kept int
	}{
		{"events-at-limit", "MaxQueuedEvents", func(p *ai.ResourcePolicy) { p.MaxQueuedEvents = int(queued) }, 0},
		{"events-over-limit", "MaxQueuedEvents", func(p *ai.ResourcePolicy) { p.MaxQueuedEvents = int(queued) - 1 }, int(queued) - 1},
		{"events-overflow-mid-deltas", "MaxQueuedEvents", func(p *ai.ResourcePolicy) { p.MaxQueuedEvents = 50 }, 50},
		{"bytes-at-limit", "MaxQueuedEventBytes", func(p *ai.ResourcePolicy) { p.MaxQueuedEventBytes = eventBytes }, 0},
		{"bytes-over-limit", "MaxQueuedEventBytes", func(p *ai.ResourcePolicy) { p.MaxQueuedEventBytes = eventBytes - 1 }, int(queued) - 1},
	}
	for _, c := range cases {
		for _, e := range streamEntries {
			t.Run(c.id+"/result-only-"+e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-queue-"+c.id+"-result-only-"+e.name)
				lw := newLimitWorld(t, c.set)
				enqueue(ev, lw.world, sseEvents(t, events, provider.FramingLF))
				s := e.start(ctxFor(t), lw.world, textScope("req-e08-queue-"+c.id+"-"+e.name), target, req)
				// Nobody reads: Result must not wait for a consumer.
				r := resultWithin(ev, "Result returns while no event is read", s)
				o := within(ev, "queued events drain after the call ended", func() outcome { return drain(s) })
				o.record(ev)
				got := eventSummary(o.events)
				peak := lw.probe.PeakQueuedEvents()
				ev.Record("probe_peak_queued_events", peak)
				ev.Check("Result and the drained stream agree", r.err == o.err && reflect.DeepEqual(r.res, o.result), "Result=%v drained=%v", r.err, o.err)
				ev.Check("the queue never held more than its limit plus the terminal", peak <= int64(lw.policyEvents(c.set))+1, "peak %d", peak)

				if c.kept == 0 {
					ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
					ev.Check("every event is delivered", reflect.DeepEqual(got, full), "got %d events", len(got))
					ev.Check("final text", textOf(o.result.Message) == strings.Join(deltas, ""), "got %q", textOf(o.result.Message))
					checkReleased(ev, lw)
					return
				}
				checkLimited(ev, o, ai.PhaseEventQueue, c.field)
				ev.Check("the queued events are the call's first events, none dropped", len(got) == c.kept+1 &&
					reflect.DeepEqual(got[:c.kept], full[:c.kept]), "got %q", got)
				// The terminal message holds everything received, including
				// the event that did not fit.
				var delivered strings.Builder
				for _, e := range o.events {
					if d, ok := e.(ai.TextDeltaEvent); ok {
						delivered.WriteString(d.Delta)
					}
				}
				text := textOf(o.result.Message)
				ev.Check("delivered deltas are a prefix of the final text", strings.HasPrefix(text, delivered.String()), "text %q", text)
				if strings.HasPrefix(full[c.kept], "text_delta") {
					ev.Check("the final text holds the delta that did not fit", len(text) > delivered.Len(), "text %q delivered %q", text, delivered.String())
				}
				checkReleased(ev, lw)
			})
		}
		if c.kept == 0 {
			continue
		}
		t.Run(c.id+"/complete", func(t *testing.T) {
			ev := run.Case(t, "E08-queue-"+c.id+"-complete")
			lw := newLimitWorld(t, c.set)
			enqueue(ev, lw.world, sseEvents(t, events, provider.FramingLF))
			r := within(ev, "Complete returns", func() callResult {
				res, err := lw.client.Complete(ctxFor(t), textScope("req-e08-queue-complete-"+c.id), target, req, nil)
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

// TestEventQueueReleasesAsRead covers the other side of the queue bounds:
// what a consumer has read no longer counts. The provider pauses once every
// delta is sent until the consumer has read them all; the text_end that
// follows carries as many bytes again and fits only because the read deltas
// left the queue.
func TestEventQueueReleasesAsRead(t *testing.T) {
	for _, e := range streamEntries {
		t.Run(e.name, func(t *testing.T) {
			ev := run.Case(t, "E08-queue-bytes-released-as-read-"+e.name)
			events, deltas := longOutput(t, longOutputDeltas)
			lw := newLimitWorld(t, func(p *ai.ResourcePolicy) { p.MaxQueuedEventBytes = int64(9 * len(deltas)) })
			read := make(chan struct{})
			reply := sseEvents(t, events, provider.FramingLF)
			lastDelta := 1 + len(deltas) // chunk index: created, added, then the deltas
			reply.AfterChunk = func(i int) {
				if i == lastDelta {
					select {
					case <-read:
					case <-time.After(waitDeadline):
					}
				}
			}
			enqueue(ev, lw.world, reply)
			s := e.start(ctxFor(t), lw.world, textScope("req-e08-queue-release-"+e.name), ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"},
				ai.Request{Messages: []ai.Message{ai.UserText("Write a long answer.")}})
			o := within(ev, "stream drains", func() outcome {
				defer s.Close()
				o := outcome{streamed: true}
				seen := 0
				for s.Next() {
					o.events = append(o.events, s.Event())
					if _, ok := s.Event().(ai.TextDeltaEvent); ok {
						if seen++; seen == len(deltas) {
							close(read)
						}
					}
				}
				o.streamErr = s.Err()
				o.result, o.err = s.Result()
				return o
			})
			o.record(ev)
			ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
			ev.Check("every event is delivered", reflect.DeepEqual(eventSummary(o.events), longOutputEvents(deltas)), "got %d events", len(o.events))
			checkReleased(ev, lw)
		})
	}
}

// policyEvents is MaxQueuedEvents of the default test policy changed by set.
func (limitWorld) policyEvents(set func(*ai.ResourcePolicy)) int {
	p := validPolicy()
	set(p)
	return p.MaxQueuedEvents
}
