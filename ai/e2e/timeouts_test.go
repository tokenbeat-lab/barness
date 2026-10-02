package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// short is the time limit a scenario expects to end it; long is a limit it
// must not reach. Their gap keeps the scenarios clear of scheduling noise.
const (
	short = 100 * time.Millisecond
	long  = 3 * time.Second
)

// TestTimeouts covers E08's time limits (spec I9, User Stories 29–30): the
// policy's connect, response header and provider read idle limits and its
// call time limit each end a call on their own; the host's deadline, the
// policy and the protocol's timeoutMs end it at whichever comes first, as
// deadline_exceeded; read idle counts only waits for upstream data; and
// every path releases what the call held.
func TestTimeouts(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-connect-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ConnectTimeout = short }, nil, func(c *ai.Config) {
					c.Transport = provider.TrackBodies(provider.StalledDialTransport())
				})
				o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-connect-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, "exceeded the resource policy's ConnectTimeout (100ms)")
				checkTimedOutAttempt(ev, o)
				ev.Check("nothing reached the provider", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("connect-ends-at-connection", func(t *testing.T) {
		// Once connected, a response slower than ConnectTimeout is not cut off.
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-connect-ends-at-connection-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ConnectTimeout = short }, nil)
				r := sseReply(t, aw.text, provider.FramingLF)
				r.OnReceive = func() { time.Sleep(3 * short) }
				enqueue(ev, aw.world, r)
				o := within(ev, "call ends", func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-connected-"+e.name) })
				checkTextSucceeded(ev, aw.text, o)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("response-header", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-response-header-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = short }, nil)
				enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent})
				o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-header-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, "exceeded the resource policy's ResponseHeaderTimeout (100ms)")
				checkTimedOutAttempt(ev, o)
				ev.Check("the provider received the request", len(aw.provider.Requests()) == 1, "got %d", len(aw.provider.Requests()))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	// The protocol's timeoutMs and the policy's ResponseHeaderTimeout bound
	// the same wait; the earlier one ends it. timeoutMs keeps the baseline's
	// "Request timed out." and error terminal.
	for _, c := range []struct {
		name      string
		policy    time.Duration
		timeoutMs int
		msg       string
	}{
		{"protocol-timeout-earlier", long, int(short / time.Millisecond), "Request timed out."},
		{"policy-earlier-than-protocol-timeout", short, int(long / time.Millisecond), "exceeded the resource policy's ResponseHeaderTimeout (100ms)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			for _, e := range timeoutEntries(c.timeoutMs) {
				t.Run(e.name, func(t *testing.T) {
					ev := run.Case(t, "E08-timeout-"+c.name+"-"+e.name)
					aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = c.policy }, nil)
					enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent})
					o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-"+c.name+"-"+e.name) })
					checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseRequest, c.msg)
					ev.Check("the whole message is the timeout's", o.result.Message.ErrorMessage == c.msg, "got %q", o.result.Message.ErrorMessage)
					checkAdmissionReleased(ev, aw)
				})
			}
		})
	}

	t.Run("invalid-protocol-timeout", func(t *testing.T) {
		ev := run.Case(t, "E08-timeout-invalid-protocol-timeout")
		aw := newAdmissionWorld(t, nil, nil)
		for _, e := range timeoutEntries(-1) {
			o := within(ev, "call ends", func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-negative-"+e.name) })
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
		}
		ev.Check("nothing reached the provider", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
	})

	t.Run("header-timeout-retried", func(t *testing.T) {
		// As the baseline retries a timed-out request, a response header
		// timeout is retried under the binding's retry policy.
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-header-retried-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ResponseHeaderTimeout = short }, nil, func(c *ai.Config) {
					c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0)
				})
				aw.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 1} })
				enqueue(ev, aw.world, provider.Reply{End: provider.EndSilent}, sseReply(t, aw.text, provider.FramingLF))
				o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-retried-"+e.name) })
				checkTextSucceeded(ev, aw.text, o)
				a := o.result.Metadata.Attempts
				ev.Check("the timed-out attempt and its retry are recorded", len(a) == 2 && a[0].Code == ai.CodeDeadlineExceeded && a[0].HTTPStatus == 0 && a[1].Code == "",
					"attempts=%+v", a)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("read-idle", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-read-idle-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ReadIdleTimeout = short }, nil)
				chunks := lfChunks(t, aw.text.Events)
				first := firstTextDelta(t, chunks)
				enqueue(ev, aw.world, heldReply(200, "text/event-stream", joinChunks(chunks[:first+1])))
				o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-idle-"+e.name) })
				checkTimedOut(ev, o, ai.StopReasonError, ai.PhaseStream, "exceeded the resource policy's ReadIdleTimeout (100ms)")
				got := contentSummary(o.result.Message.Content)
				ev.Check("text received before the stall is kept", len(got) == 1 && strings.HasPrefix(got[0], "text:"), "got %q", got)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("read-idle-counts-only-upstream-waits", func(t *testing.T) {
		// A slow response callback before the body is read and gaps between
		// chunks that each stay below the limit, together far above it,
		// never end the call: only a single wait for upstream data counts.
		const idle = 200 * time.Millisecond
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-read-idle-only-upstream-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { p.ReadIdleTimeout = idle }, nil)
				hooked := hookedEntries(aw.client.WithHooks(ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
					select {
					case <-time.After(3 * idle):
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				}}))
				i := entryIndex(t, e.name)
				r := sseReply(t, aw.text, provider.FramingLF)
				r.AfterChunk = func(int) { time.Sleep(idle / 4) }
				ev.Record("pacing", map[string]any{"chunks": len(r.Chunks), "gap": (idle / 4).String(), "idle": idle.String()})
				enqueue(ev, aw.world, r)
				start := time.Now()
				o := within(ev, "call ends", func() outcome { return aw.call(ctxFor(t), hooked[i], tenantA, "req-only-upstream-"+e.name) })
				elapsed := time.Since(start)
				checkTextSucceeded(ev, aw.text, o)
				ev.Check("the call outlasted the idle limit several times", elapsed > 3*idle, "took %s", elapsed)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("call-timeout", func(t *testing.T) {
		// Upstream keeps sending, so no per-attempt limit ends the call; the
		// policy's CallTimeout does, as the call's own deadline.
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-call-"+e.name)
				const callTimeout = 300 * time.Millisecond
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { fitCallTimeout(p, callTimeout) }, nil)
				events, _ := longOutput(t, 200)
				r := sseEvents(t, events, provider.FramingLF)
				r.AfterChunk = func(int) { time.Sleep(20 * time.Millisecond) }
				enqueue(ev, aw.world, r)
				start := time.Now()
				o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-call-timeout-"+e.name) })
				elapsed := time.Since(start)
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream, "OpenAI Responses stream ended before a terminal response event")
				ev.Check("the call ended at CallTimeout", elapsed >= callTimeout, "took %s", elapsed)
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("call-timeout-covers-setup", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-call-setup-"+e.name)
				aw := newAdmissionWorld(t, func(p *ai.ResourcePolicy) { fitCallTimeout(p, short) }, nil)
				hold := aw.host.HoldBindings()
				t.Cleanup(hold.Release)
				o := timedCall(ev, long, func() outcome { return aw.call(ctxFor(t), e, tenantA, "req-call-setup-"+e.name) })
				checkInterrupted(ev, o, ai.StopReasonError, ai.CodeDeadlineExceeded, ai.PhaseBinding, "Request timed out")
				ev.Check("nothing reached the provider", len(aw.provider.Requests()) == 0, "got %d", len(aw.provider.Requests()))
				checkAdmissionReleased(ev, aw)
			})
		}
	})

	t.Run("host-deadline-earlier", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-timeout-host-deadline-"+e.name)
				aw := newAdmissionWorld(t, nil, nil)
				enqueue(ev, aw.world, heldReply(200, "text/event-stream", string(lfChunks(t, aw.text.Events)[0])))
				ctx, cancel := context.WithTimeout(ctxFor(t), short)
				defer cancel()
				o := timedCall(ev, long, func() outcome { return aw.call(ctx, e, tenantA, "req-host-deadline-"+e.name) })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseStream, "OpenAI Responses stream ended before a terminal response event")
				checkAdmissionReleased(ev, aw)
			})
		}
	})
}

// timedCall runs fn and fails the check unless it returns before limit: the
// limit a scenario expects to end the call is far below it.
func timedCall(ev *evidence.Case, limit time.Duration, fn func() outcome) outcome {
	start := time.Now()
	o := within(ev, "call ends", fn)
	elapsed := time.Since(start)
	o.record(ev)
	ev.Record("elapsed", elapsed.String())
	ev.Check("the earlier limit ended the call", elapsed < limit, "took %s", elapsed)
	return o
}

// checkTimedOut asserts a call ended by a time limit: deadline_exceeded in
// phase, the given terminal, and a message naming the limit.
func checkTimedOut(ev *evidence.Case, o outcome, stop ai.StopReason, phase ai.Phase, msg string) {
	m := o.result.Message
	ev.Check("stop reason", m.StopReason == stop, "got %q want %q", m.StopReason, stop)
	ev.Check("errors.Is matches deadline_exceeded and phase", errors.Is(o.err, &ai.Error{Code: ai.CodeDeadlineExceeded, Phase: phase}),
		"want deadline_exceeded/%s, got %v", phase, o.err)
	ev.Check("errorMessage names the limit", strings.Contains(m.ErrorMessage, msg), "got %q want %q in it", m.ErrorMessage, msg)
	if o.streamed {
		got := eventSummary(o.events)
		ev.Check("stream ends with one terminal", len(got) > 0 && got[len(got)-1] == "error:"+string(stop), "got %q", got)
		ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
	}
}

// checkTimedOutAttempt asserts the timed-out request is recorded as one
// attempt without a response.
func checkTimedOutAttempt(ev *evidence.Case, o outcome) {
	a := o.result.Metadata.Attempts
	ev.Check("one attempt, timed out without a response", len(a) == 1 && a[0].Code == ai.CodeDeadlineExceeded && a[0].HTTPStatus == 0,
		"attempts=%+v", a)
}

// fitCallTimeout sets CallTimeout to d and lowers the limits that must fit
// within it.
func fitCallTimeout(p *ai.ResourcePolicy, d time.Duration) {
	p.CallTimeout, p.AdmissionWait = d, 0
	p.ConnectTimeout, p.ResponseHeaderTimeout, p.ReadIdleTimeout = d, d, d
}

// timeoutEntries are the four entry points with the protocol's timeoutMs.
func timeoutEntries(ms int) []outcomeEntry {
	full := ai.ResponsesOptions{TimeoutMs: ms}
	simple := ai.SimpleOptions{TimeoutMs: ms}
	return []outcomeEntry{
		{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.Stream(ctx, scope, target, req, full))
		}},
		{"stream-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.StreamSimple(ctx, scope, target, req, simple))
		}},
		{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.Complete(ctx, scope, target, req, full)
			return outcome{result: res, err: err}
		}},
		{"complete-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.CompleteSimple(ctx, scope, target, req, simple)
			return outcome{result: res, err: err}
		}},
	}
}

func entryIndex(t *testing.T, name string) int {
	t.Helper()
	for i, e := range outcomeEntries {
		if e.name == name {
			return i
		}
	}
	t.Fatalf("no entry %q", name)
	return 0
}

// firstTextDelta is the index of the first chunk carrying a text delta.
func firstTextDelta(t *testing.T, chunks [][]byte) int {
	t.Helper()
	for i, c := range chunks {
		if strings.Contains(string(c), "response.output_text.delta") {
			return i
		}
	}
	t.Fatal("no text delta in the fixture")
	return 0
}

func joinChunks(chunks [][]byte) string {
	var b strings.Builder
	for _, c := range chunks {
		b.Write(c)
	}
	return b.String()
}
