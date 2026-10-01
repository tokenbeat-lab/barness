package e2e

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestStreamLifecycle covers E01: every legal way of consuming a call — event
// by event, Result only, both in parallel, or Close at any time — with stable
// event order and block indexes, a live PartialView that is safe to read
// concurrently, Complete without an event backlog, and inputs and outputs that
// share no mutable storage with the caller (spec I1, I2, I8).
func TestStreamLifecycle(t *testing.T) {
	t.Run("close-while-resolver-blocked", func(t *testing.T) {
		for _, e := range streamEntries {
			t.Run(e.name, func(t *testing.T) {
				ev, w, f := startLifecycleCase(t, "E01-close-while-resolver-blocked-"+e.name, provider.FramingLF)
				hold := w.host.HoldBindings()
				defer hold.Release()
				s := within(ev, "Stream returns while the resolver is blocked", func() *ai.Stream {
					return e.start(ctxFor(t), w, textScope("req-close-blocked-"+e.name), lifecycleTarget(f), lifecycleRequest(f))
				})
				if s == nil {
					return
				}
				waitFor(ev, "call is blocked in the binding resolver", hold.Entered())

				// Close from several goroutines at once, then once more: each
				// returns only after the background process has ended.
				var wg sync.WaitGroup
				for range 4 {
					wg.Add(1)
					go func() { defer wg.Done(); _ = s.Close() }()
				}
				waitFor(ev, "concurrent Close calls return", doneWhen(wg.Wait))
				within(ev, "repeated Close returns", func() error { return s.Close() })

				o := within(ev, "stream drains after Close", func() outcome { return drain(s) })
				o.record(ev)
				checkFailed(ev, o.result, o.err)
				ev.Check("Close during setup ends the call as canceled",
					errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled, Phase: ai.PhaseBinding}), "got %v", o.err)
				ev.Check("only an error terminal was published", reflect.DeepEqual(eventSummary(o.events), []string{"error:error"}),
					"got %q", eventSummary(o.events))
				ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
			})
		}
	})

	t.Run("interleaved-blocks", func(t *testing.T) {
		for _, e := range lifecycleEntries {
			for _, framing := range []provider.Framing{provider.FramingLF, provider.FramingBytewise} {
				name := e.name + "-" + string(framing)
				t.Run(name, func(t *testing.T) {
					ev, w, f := startLifecycleCase(t, "E01-interleaved-blocks-"+name, framing)
					res := within(ev, "call finishes", func() ai.Result { return e.invoke(ev, w, f, textScope("req-interleaved-"+name)) })
					checkLifecycleRequest(ev, w, f)
					checkLifecycleMessage(ev, f, res.Message)
				})
			}
		}
	})

	t.Run("live-partial-concurrent-reads", func(t *testing.T) {
		ev, w, f := startLifecycleCase(t, "E01-live-partial-concurrent-reads", provider.FramingBytewise)
		s := w.client.Stream(ctxFor(t), textScope("req-live-partial"), lifecycleTarget(f), lifecycleRequest(f), nil)
		defer s.Close()
		if !ev.Check("first event", within(ev, "first Next returns", s.Next), "stream ended before start") {
			return
		}
		view := partialOf(s.Event())
		if !ev.Check("start carries the live PartialView", view != nil, "event=%#v", s.Event()) {
			return
		}

		// A reader polls the view while the consumer below drains events. The
		// race detector proves the view is synchronized; each snapshot must be a
		// state the call actually passed through.
		stop := make(chan struct{})
		readerDone := make(chan []string, 1)
		go func() {
			var problems []string
			for {
				snap := view.Snapshot()
				if p := checkGrowingSnapshot(f, snap); p != "" {
					problems = append(problems, p)
				}
				select {
				case <-stop:
					readerDone <- problems
					return
				default:
				}
			}
		}()

		seen := map[int]string{} // content index → deltas so far
		var lagging []string
		within(ev, "stream drains", func() bool {
			for s.Next() {
				e := s.Event()
				if p := partialOf(e); p != nil && p != view {
					ev.Check("every event refers to the same view", false, "%s has another view", e.Type())
				}
				var idx int
				switch e := e.(type) {
				case ai.TextDeltaEvent:
					idx, seen[e.ContentIndex] = e.ContentIndex, seen[e.ContentIndex]+e.Delta
				case ai.ThinkingDeltaEvent:
					idx, seen[e.ContentIndex] = e.ContentIndex, seen[e.ContentIndex]+e.Delta
				default:
					continue
				}
				// Live, not event-time: the view already holds at least everything
				// this delta's event reported, possibly more.
				snap := view.Snapshot()
				if idx >= len(snap.Content) || !strings.HasPrefix(blockText(snap.Content[idx]), strings.TrimSuffix(seen[idx], "\n\n")) {
					lagging = append(lagging, fmt.Sprintf("block %d: view %q lacks %q", idx, snap.Content, seen[idx]))
				}
			}
			return true
		})
		close(stop)
		problems := waitValue(ev, "snapshot reader stops", readerDone)
		ev.Record("reader_problems", problems)
		ev.Check("concurrent snapshots are states the call passed through", len(problems) == 0, "%s", strings.Join(problems, "\n"))
		ev.Check("view is never behind the event being read", len(lagging) == 0, "%s", strings.Join(lagging, "\n"))

		r := resultWithin(ev, "Result returns", s)
		res := r.res
		ev.Check("call succeeds", r.err == nil, "err=%v", r.err)
		final := view.Snapshot()
		ev.Check("view equals the final message after the terminal", reflect.DeepEqual(final, res.Message),
			"view=%+v\nresult=%+v", final, res.Message)
		final.Content[0] = ai.Text{Text: "caller scribble"}
		final.Content = append(final.Content, ai.Text{Text: "appended"})
		again := view.Snapshot()
		ev.Check("final message is stable and snapshots share no storage", reflect.DeepEqual(again, res.Message),
			"view=%+v", again)
	})

	t.Run("result-without-next", func(t *testing.T) {
		for _, e := range streamEntries {
			t.Run(e.name, func(t *testing.T) {
				ev, w, f := startLifecycleCase(t, "E01-result-without-next-"+e.name, provider.FramingLF)
				s := e.start(ctxFor(t), w, textScope("req-result-only-"+e.name), lifecycleTarget(f), lifecycleRequest(f))
				defer s.Close()
				r := resultWithin(ev, "Result returns without any Next", s)
				ev.Record("result", r.res)
				ev.Check("Result error is nil", r.err == nil, "got %v", r.err)
				ev.Check("Err stays nil until Next returns false", s.Err() == nil, "got %v", s.Err())
				checkLifecycleMessage(ev, f, r.res.Message)
				checkLifecycleRequest(ev, w, f)
			})
		}
	})

	t.Run("result-parallel-with-consumption", func(t *testing.T) {
		ev, w, f := startLifecycleCase(t, "E01-result-parallel-with-consumption", provider.FramingBytewise)
		s := w.client.Stream(ctxFor(t), textScope("req-parallel"), lifecycleTarget(f), lifecycleRequest(f), nil)
		defer s.Close()
		waiter := make(chan callResult, 1)
		go func() {
			res, err := s.Result()
			waiter <- callResult{res, err}
		}()
		o := within(ev, "events drain while Result is awaited", func() outcome { return drain(s) })
		r := waitValue(ev, "parallel Result returns", waiter)
		o.record(ev)
		checkLifecycleEvents(ev, f, o.events, o.result)
		ev.Check("parallel Result error is nil", r.err == nil, "got %v", r.err)
		ev.Check("Err is nil after a successful terminal", o.streamErr == nil, "got %v", o.streamErr)
		ev.Check("parallel Result equals the drained Result", reflect.DeepEqual(r.res, o.result), "parallel=%+v\ndrained=%+v", r.res, o.result)
	})

	t.Run("err-scanner-style", func(t *testing.T) {
		// A refused call: Err stays nil before Next and while the error
		// terminal is current, and equals the Result error afterwards.
		ev, w, f := startLifecycleCase(t, "E01-err-scanner-style", provider.FramingLF)
		s := w.client.Stream(ctxFor(t), textScope("req-err-scanner"), ai.Target{BindingID: "missing", ModelID: f.Model}, lifecycleRequest(f), nil)
		defer s.Close()
		before := s.Err()
		var during []error
		within(ev, "stream drains", func() bool {
			for s.Next() {
				during = append(during, s.Err())
			}
			return true
		})
		after := s.Err()
		r := resultWithin(ev, "Result returns", s)
		err := r.err
		ev.Record("result", r.res)
		ev.Check("Err is nil before Next", before == nil, "got %v", before)
		ev.Check("Err is nil while events are current", len(during) == 1 && during[0] == nil, "got %v", during)
		ev.Check("Err equals the Result error after Next returns false", after != nil && after == err, "Err=%v Result=%v", after, err)
	})

	t.Run("close-after-terminal-keeps-result", func(t *testing.T) {
		ev, w, f := startLifecycleCase(t, "E01-close-after-terminal-keeps-result", provider.FramingLF)
		s := w.client.Stream(ctxFor(t), textScope("req-close-after"), lifecycleTarget(f), lifecycleRequest(f), nil)
		o := within(ev, "stream drains", func() outcome { return drain(s) })
		var wg sync.WaitGroup
		for range 4 {
			wg.Add(1)
			go func() { defer wg.Done(); _ = s.Close() }()
		}
		waitFor(ev, "concurrent Close after the terminal returns", doneWhen(wg.Wait))
		r := resultWithin(ev, "Result returns", s)
		ev.Record("result", r.res)
		ev.Check("Close after the terminal changes nothing", r.err == nil && reflect.DeepEqual(r.res, o.result),
			"err=%v\nbefore=%+v\nafter=%+v", r.err, o.result, r.res)
		ev.Check("Next stays false", !within(ev, "Next returns", s.Next), "Next returned true after the terminal")
	})

	t.Run("complete-queues-no-events", func(t *testing.T) {
		for _, e := range lifecycleEntries {
			if !strings.HasPrefix(e.name, "complete-") {
				continue
			}
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E01-complete-queues-no-events-"+e.name)
				f, raw := loadLifecycleFixture(t)
				ev.Fixture("fixture.json", raw)
				p := probe.New()
				w := newWorldWith(t, func(c *ai.Config) { c.Probe = p }, tenantA)
				enqueue(ev, w, sseEvents(t, f.Events, provider.FramingLF), sseEvents(t, f.Events, provider.FramingLF))

				res := within(ev, "Complete returns", func() ai.Result { return e.invoke(ev, w, f, textScope("req-complete-probe-"+e.name)) })
				checkLifecycleMessage(ev, f, res.Message)
				completePeak := p.PeakQueuedEvents()

				// The same output on a Stream whose events nobody reads is
				// queued in full, so the probe demonstrably observes queueing.
				s := w.client.Stream(ctxFor(t), textScope("req-stream-probe-"+e.name), lifecycleTarget(f), lifecycleRequest(f), nil)
				resultWithin(ev, "unread Stream's Result returns", s)
				streamPeak := p.PeakQueuedEvents()
				_ = s.Close()
				ev.Record("probe", map[string]int64{"complete_peak_queued_events": completePeak, "stream_peak_queued_events": streamPeak})
				ev.Check("Complete queued no events", completePeak == 0, "peak %d", completePeak)
				ev.Check("an unread Stream queues every event", streamPeak == int64(len(f.Expect.Events)),
					"peak %d, want %d", streamPeak, len(f.Expect.Events))
			})
		}
	})

	t.Run("inputs-copied-on-receipt", func(t *testing.T) {
		for _, e := range lifecycleEntries {
			t.Run(e.name, func(t *testing.T) {
				ev, w, f := startLifecycleCase(t, "E01-inputs-copied-on-receipt-"+e.name, provider.FramingLF)
				hold := w.host.HoldBindings()
				defer hold.Release()
				req := lifecycleRequest(f)
				content := req.Messages[0].(ai.UserMessage).Content
				got := make(chan ai.Result, 1)
				go func() { got <- e.invokeWith(ev, w, f, textScope("req-copy-"+e.name), req) }()
				waitFor(ev, "call is blocked in the binding resolver", hold.Entered())
				// Mutate everything the caller still holds while the call is in
				// flight; none of it may reach the call.
				content[0] = ai.Text{Text: "mutated after hand-over"}
				req.Messages[0] = ai.UserText("replaced after hand-over")
				req.Messages = append(req.Messages, ai.UserText("appended after hand-over"))
				hold.Release()
				res := waitValue(ev, "call finishes", got)
				checkLifecycleRequest(ev, w, f)
				checkLifecycleMessage(ev, f, res.Message)
			})
		}
	})

	t.Run("outputs-share-no-storage", func(t *testing.T) {
		ev, w, f := startLifecycleCase(t, "E01-outputs-share-no-storage", provider.FramingLF)
		s := w.client.Stream(ctxFor(t), textScope("req-outputs"), lifecycleTarget(f), lifecycleRequest(f), nil)
		o := within(ev, "stream drains", func() outcome { return drain(s) })
		done, ok := o.events[len(o.events)-1].(ai.DoneEvent)
		if !ev.Check("stream ends with done", ok, "got %q", eventSummary(o.events)) {
			return
		}
		first := resultWithin(ev, "Result returns", s).res
		first.Message.Content[0] = ai.Text{Text: "scribbled on a Result"}
		done.Message.Content[1] = ai.Text{Text: "scribbled on the done event"}
		view := partialOf(o.events[0])
		snap := view.Snapshot()
		snap.Content[2] = ai.Text{Text: "scribbled on a snapshot"}

		second := resultWithin(ev, "Result returns again", s).res
		ev.Record("result", second)
		checkLifecycleMessage(ev, f, second.Message)
		ev.Check("view unaffected by scribbles", reflect.DeepEqual(view.Snapshot(), second.Message), "view=%+v", view.Snapshot())
	})
}
