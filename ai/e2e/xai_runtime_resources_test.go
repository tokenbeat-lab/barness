package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAIRuntimeBounds(t *testing.T) {
	for _, mode := range []string{"request", "frame", "tool-json", "ciphertext", "repeated-terminal", "queue-count", "queue-bytes", "read-timeout", "header-timeout"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-E15-bound-"+mode)
			frames := xaiFrames()
			if mode == "tool-json" {
				frames = []json.RawMessage{xaiTerminal(xaiCallA, xaiUsage)}
			}
			if mode == "ciphertext" {
				frames = []json.RawMessage{xaiTerminal(`{"type":"reasoning","id":"r","summary":[],"encrypted_content":"`+strings.Repeat("cipher", 200)+`"}`, xaiUsage)}
			}
			if mode == "repeated-terminal" {
				final := xaiTerminal(xaiMessage, xaiUsage)
				frames = []json.RawMessage{final, final, final}
			}
			w := newXAIRuntimeWorld(t, nil, func(c *ai.Config) {
				switch mode {
				case "request":
					setRequestBytes(c.Policy, 300)
				case "frame":
					setFrameBytes(c.Policy, 256)
				case "tool-json":
					setToolJSONBytes(c.Policy, 4)
				case "ciphertext":
					setOutputBytes(c.Policy, 1000)
				case "repeated-terminal":
					setOutputBytes(c.Policy, int64(len(lfChunks(t, frames[:1])[0])+10))
				case "queue-count":
					c.Policy.MaxQueuedEvents = 2
				case "queue-bytes":
					c.Policy.MaxQueuedEventBytes = 4
				case "read-timeout":
					c.Policy.ReadIdleTimeout = short
				case "header-timeout":
					c.Policy.ResponseHeaderTimeout = short
				}
			})
			reply := xaiReply(t, frames...)
			if mode == "read-timeout" {
				reply = xaiReply(t, frames[:2]...)
				reply.End = provider.EndHold
			}
			if mode == "header-timeout" {
				reply = provider.Reply{End: provider.EndSilent}
			}
			// Request rejection has no HTTP script to consume. A later explicit
			// call must reach the recovery script, rather than a leftover failure.
			if mode != "request" {
				enqueue(ev, w.world, reply)
			} else {
				ev.Record("rejected-script", reply)
			}
			ev.Record("frames", frames)
			s := w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiToolRequest(), nil)
			defer s.Close()
			settled := resultWithin(ev, "bounded producer settles without consumer", s)
			o := drain(s)
			o.record(ev)
			code, phase := ai.CodeResourceLimit, ai.PhaseStream
			if mode == "request" {
				phase = ai.PhaseCapability
			}
			if strings.HasPrefix(mode, "queue-") {
				phase = ai.PhaseEventQueue
			}
			if strings.HasSuffix(mode, "timeout") {
				code = ai.CodeDeadlineExceeded
			}
			if mode == "header-timeout" {
				phase = ai.PhaseRequest
			}
			ev.Check("bound has precise classification and immutable settlement", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}) && settled.err == o.err, "%v", o.err)
			glmSharedOneTerminal(ev, o)
			w.released(ev, "after-failure")
			// This finite, empty response is valid below every selected byte
			// limit. Complete proves admission/body recovery even when the
			// Stream-specific queue has deliberately tiny capacity.
			enqueue(ev, w.world, xaiReply(t, xaiTerminal("", "")))
			res, err := w.client.Complete(ctxFor(t), textScope("recover-"+mode), xaiTarget(), ai.Request{Messages: []ai.Message{ai.UserText("ok")}}, nil)
			ev.Record("recovery-result", res)
			ev.Record("recovery-error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("next explicit call recovers using same Client", err == nil && len(res.Metadata.Attempts) == 1, "%v", err)
			w.released(ev, "after-recovery")
		})
	}
}

func TestXAIRuntimeCompleteNoEventQueue(t *testing.T) {
	ev := run.Case(t, "XAI-E15-complete-no-event-queue")
	w := newXAIRuntimeWorld(t, nil, func(c *ai.Config) { c.Policy.MaxQueuedEvents = 1; c.Policy.MaxQueuedEventBytes = 1 })
	enqueue(ev, w.world, xaiReply(t))
	res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil)
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	ev.Record("peak-queued-events", w.gauges.PeakQueuedEvents())
	ev.Check("Complete aggregates without an event queue", err == nil && textOf(res.Message) == "Hello Grok." && w.gauges.PeakQueuedEvents() == 0, "%v peak=%d", err, w.gauges.PeakQueuedEvents())
	w.released(ev)
}

func TestXAIRuntimeCancellation(t *testing.T) {
	for _, mode := range []string{"before-send", "before-headers", "after-content"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-E15-cancel-"+mode)
			w := newXAIRuntimeWorld(t, nil)
			b := xaiBinding()
			b.Retry.MaxRetries = 1
			w.host.PutBinding(b)
			ready := make(chan struct{})
			r := xaiReply(t, xaiFrames()[:2]...)
			r.End = provider.EndHold
			if mode == "before-headers" {
				r = provider.Reply{End: provider.EndSilent, OnReceive: func() { close(ready) }}
			}
			if mode != "before-send" {
				enqueue(ev, w.world, r)
			}
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			if mode == "before-send" {
				cancel()
			}
			s := w.client.Stream(ctx, textScope(t.Name()), xaiTarget(), xaiRequest(), nil)
			defer s.Close()
			var o outcome
			o.streamed = true
			if mode == "before-headers" {
				waitFor(ev, "request accepted before headers", ready)
				cancel()
			}
			if mode == "after-content" {
				for s.Next() {
					o.events = append(o.events, s.Event())
					if _, ok := s.Event().(ai.TextDeltaEvent); ok {
						cancel()
						break
					}
				}
			}
			settled := resultWithin(ev, "canceled Result needs no consumer", s)
			for s.Next() {
				o.events = append(o.events, s.Event())
			}
			o.result, o.err = s.Result()
			o.streamErr = s.Err()
			o.record(ev)
			ev.Check("cancellation settles once", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}) && o.err == settled.err, "%v", o.err)
			want := 1
			if mode == "before-send" {
				want = 0
			}
			ev.Check("canceled stream is never replayed", len(w.provider.Requests()) == want && len(o.result.Metadata.Attempts) == want, "%+v", o.result.Metadata)
			if mode == "after-content" {
				ev.Check("known prefix survives cancellation", textOf(o.result.Message) == "Hello Grok." && o.result.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "%+v", o.result)
			}
			glmSharedOneTerminal(ev, o)
			w.released(ev, "after-cancellation")
			enqueue(ev, w.world, xaiReply(t))
			res, err := w.client.Complete(ctxFor(t), textScope("recover-cancel-"+mode), xaiTarget(), xaiRequest(), nil)
			ev.Record("recovery", res)
			ev.Record("requests", w.provider.Requests())
			ev.Check("explicit recovery succeeds", err == nil, "%v", err)
			w.released(ev, "after-recovery")
		})
	}
}
