package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestFailureTerminals covers E02 on P01: setup, HTTP, in-stream, protocol and
// transport failures, normal truncation, and cancellation and deadlines in
// every phase. Each ends in one complete final message whose StopReason and
// classified error tell normal stop, truncation, failure and cancellation
// apart, keeps what was received, and releases every response body (spec I8,
// I9, E02, P01).
func TestFailureTerminals(t *testing.T) {
	f, raw := loadFailureFixture(t)
	for _, sc := range f.Scenarios {
		for _, e := range outcomeEntries {
			streamed := strings.HasPrefix(e.name, "stream-")
			if sc.CancelAfterEvents > 0 && !streamed {
				continue
			}
			t.Run(sc.ID+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E02-"+sc.ID+"-"+e.name)
				ev.Fixture("failures.json", raw)
				w, bodies := newTrackedWorld(t)
				if sc.Unreachable {
					w.updateBinding(tenantA, func(b *ai.Binding) { b.Endpoint = deadEndpoint() })
				} else {
					enqueue(ev, w, f.reply(t, sc))
				}
				scope := textScope("req-e02-" + sc.ID + "-" + e.name)
				var o outcome
				if sc.CancelAfterEvents > 0 {
					s := w.client.Stream(ctxFor(t), scope, f.target(), f.request(), nil)
					o = within(ev, "stream ends after Close", func() outcome { return drainClosingAfter(s, sc.CancelAfterEvents) })
				} else {
					o = within(ev, "call ends", func() outcome { return e.invoke(ctxFor(t), w, scope, f.target(), f.request()) })
				}
				o.record(ev)
				checkFailureOutcome(ev, f, sc, o)
				checkProviderSaw(ev, w, sc)
				ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
			})
		}
	}

	t.Run("cancel-during-request", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E02-cancel-during-request-"+e.name)
				ctx, cancel := context.WithCancel(ctxFor(t))
				o := callWhileProviderHolds(t, ev, e, ctx, cancel)
				// pi's provider-retry wrapper replaces the SDK's abort error
				// with its own, whether or not retries are configured.
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeCanceled, ai.PhaseRequest, "Request aborted")
			})
		}
	})

	t.Run("deadline-during-request", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E02-deadline-during-request-"+e.name)
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()
				o := callWhileProviderHolds(t, ev, e, ctx, func() { <-ctx.Done() })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseRequest, "Request timed out.")
			})
		}
	})

	t.Run("deadline-during-setup", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				// Baseline stage semantics: an interrupted setup is an error, not
				// an abort; the classification still says it was the deadline.
				ev := run.Case(t, "P01-E02-deadline-during-setup-"+e.name)
				w, bodies := newTrackedWorld(t)
				hold := w.host.HoldBindings()
				defer hold.Release()
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()
				o := within(ev, "call ends at the deadline", func() outcome {
					return e.invoke(ctx, w, textScope("req-setup-deadline-"+e.name), f.target(), f.request())
				})
				o.record(ev)
				checkInterrupted(ev, o, ai.StopReasonError, ai.CodeDeadlineExceeded, ai.PhaseBinding, "Request timed out")
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
			})
		}
	})

	t.Run("deadline-during-stream", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E02-deadline-during-stream-"+e.name)
				ev.Fixture("failures.json", raw)
				w, bodies := newTrackedWorld(t)
				// The shared partial events, then the provider keeps the stream open.
				held := failureScenario{Reply: failureReply{Status: 200, End: provider.EndHold}, Expect: failureExpect{Partial: true}}
				reply := f.reply(t, held)
				enqueue(ev, w, reply)
				ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
				defer cancel()
				o := within(ev, "call ends at the deadline", func() outcome {
					return e.invoke(ctx, w, textScope("req-stream-deadline-"+e.name), f.target(), f.request())
				})
				o.record(ev)
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream,
					"OpenAI Responses stream ended before a terminal response event")
				// The deadline is not synchronized with the client's reading, so
				// the partial text is only known to be a prefix of what was sent.
				text := strings.Join(contentSummary(o.result.Message.Content), "")
				ev.Check("partial content is kept", strings.HasPrefix("text:Hello, ", text) && text != "", "got %q", text)
				ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
			})
		}
	})
}
