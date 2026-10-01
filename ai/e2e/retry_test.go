package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestExplicitRetry covers E05: the initial request is retried only as the
// operator configured on the binding, by the frozen pi rules (x-should-retry,
// then the network-error shape and 408/409/429/5xx; retry-after-ms, then
// retry-after seconds or date, then capped exponential backoff with jitter;
// maxRetryDelay). Retries are never replayed once the stream started, every
// attempt is recorded separately under the same RequestID and snapshot, and
// onPayload/onResponse run once per logical call (spec I8, User Stories
// 31–32). The clock is fake, so no case sleeps for real.
func TestExplicitRetry(t *testing.T) {
	f, raw := loadRetryFixture(t)
	text, textRaw := loadTextFixture(t, "text-basic.json")
	for _, sc := range f.Scenarios {
		for i, base := range outcomeEntries {
			t.Run(sc.ID+"/"+base.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E05-"+sc.ID+"-"+base.name)
				ev.Fixture("retry.json", raw)
				ev.Fixture("text-basic.json", textRaw)
				clk := f.clock()
				w, bodies := newRetryWorld(t, clk, sc.Retry.policy())
				enqueue(ev, w, sc.replies(t, text)...)
				var hooks hookCounts
				e := hookedEntries(w.client.WithHooks(hooks.hooks()))[i]
				scope := textScope("req-e05-" + sc.ID + "-" + base.name)
				o := within(ev, "call ends", func() outcome { return e.invoke(ctxFor(t), w, scope, textTarget(text), textRequest(text)) })
				o.record(ev)

				x := sc.Expect
				checkAttemptRequests(ev, w, slices.Repeat([]tenantKey{tenantA}, x.Requests)...)
				checkSleeps(ev, clk.Sleeps(), x.SleepsMs)
				ev.Check("onPayload runs once per logical call", hooks.payloads.Load() == 1, "got %d", hooks.payloads.Load())
				ev.Check("onResponse runs only after an initial response", hooks.responses.Load() == int32(x.Responses), "got %d want %d", hooks.responses.Load(), x.Responses)
				checkRetryOutcome(ev, text, sc, o)
				checkAttempts(ev, scope, o.result.Metadata.Attempts, x.Attempts)
				ev.Check("retries never re-resolve the snapshot", w.host.ReadsFor(scope.RequestID) == host.Reads{Bindings: 1, Credentials: 1}, "got %+v", w.host.ReadsFor(scope.RequestID))
				ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
				checkNoSecrets(ev, o)
			})
		}
	}

	t.Run("repeatable", func(t *testing.T) {
		// The same scenario on a fresh world yields the same attempts and delays.
		ev := run.Case(t, "P01-E05-repeatable")
		sc := retryScenarioByID(t, f, "retryable-statuses")
		var seen []string
		for range 2 {
			clk := f.clock()
			w, _ := newRetryWorld(t, clk, sc.Retry.policy())
			enqueue(ev, w, sc.replies(t, text)...)
			res, err := w.client.Complete(ctxFor(t), textScope("req-e05-repeatable"), textTarget(text), textRequest(text), nil)
			ev.Check("call succeeds", err == nil, "err=%v", err)
			seen = append(seen, fmt.Sprintf("requests=%d sleeps=%v attempts=%+v", len(w.provider.Requests()), clk.Sleeps(), res.Metadata.Attempts))
		}
		ev.Record("runs", seen)
		ev.Check("both runs agree", seen[0] == seen[1], "first %s\nsecond %s", seen[0], seen[1])
	})

	t.Run("cancel-during-backoff", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E05-cancel-during-backoff-"+e.name)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				o := interruptBackoff(t, ev, f, text, e, ctx, cancel)
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeCanceled, ai.PhaseRequest, "Request aborted")
			})
		}
	})

	t.Run("deadline-bounds-uncapped-backoff", func(t *testing.T) {
		// maxRetryDelay 0 lifts only the per-delay cap: a 120s wait still
		// ends at the call's deadline.
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E05-deadline-bounds-uncapped-backoff-"+e.name)
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()
				o := interruptBackoff(t, ev, f, text, e, ctx, func() { <-ctx.Done() })
				checkInterrupted(ev, o, ai.StopReasonAborted, ai.CodeDeadlineExceeded, ai.PhaseRequest, "Request timed out.")
			})
		}
	})

	t.Run("attempts-keep-the-snapshot", func(t *testing.T) {
		// While the first attempt is at the Provider the tenant rotates K1→K2
		// with a new binding version. The remaining attempt still uses the
		// call's snapshot; only a new logical call sees K2.
		ev := run.Case(t, "P01-E05-attempts-keep-the-snapshot")
		w, _ := newRetryWorld(t, f.clock(), ai.RetryPolicy{MaxRetries: 1})
		failing := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte("{}")}, Script: "{}"}
		failing.OnReceive = func() {
			w.updateBinding(tenantA, func(b *ai.Binding) { b.Version = "b2" })
			w.host.PutCredential(primaryCredential(tenantA2, "v2"))
		}
		enqueue(ev, w, failing, sseReply(t, text, provider.FramingLF), sseReply(t, text, provider.FramingLF))
		pinned, err := w.client.Complete(ctxFor(t), textScope("req-e05-pinned"), textTarget(text), textRequest(text), nil)
		ev.Record("pinned", pinned)
		ev.Check("retried call succeeds", err == nil, "err=%v", err)
		ev.Check("two attempts", len(pinned.Metadata.Attempts) == 2, "attempts=%+v", pinned.Metadata.Attempts)
		next, err := w.client.Complete(ctxFor(t), textScope("req-e05-after-rotation"), textTarget(text), textRequest(text), nil)
		ev.Record("next", next)
		ev.Check("new call succeeds", err == nil, "err=%v", err)
		checkAttemptRequests(ev, w, tenantA, tenantA, tenantA2)
	})

	t.Run("invalid-retry-policy", func(t *testing.T) {
		negative := -time.Second
		for name, p := range map[string]ai.RetryPolicy{
			"negative-max-retries":     {MaxRetries: -1},
			"negative-max-retry-delay": {MaxRetries: 1, MaxRetryDelay: &negative},
		} {
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P01-E05-invalid-retry-policy-"+name)
				w, _ := newRetryWorld(t, f.clock(), p)
				res, err := w.client.Complete(ctxFor(t), textScope("req-e05-invalid-"+name), textTarget(text), textRequest(text), nil)
				ev.Record("result", res)
				checkRejected(ev, outcome{result: res, err: err}, ai.CodeInvalidRequest, ai.PhaseBinding)
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				ev.Check("no attempt recorded", len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata.Attempts)
			})
		}
	})
}

// interruptBackoff starts a call on e that fails its first attempt with a
// 429 asking for 120s, with maxRetryDelay 0 so the wait is not refused, runs
// interrupt once that wait has begun, and returns the outcome.
func interruptBackoff(t *testing.T, ev *evidence.Case, f retryFixture, text textFixture, e outcomeEntry, ctx context.Context, interrupt func()) outcome {
	t.Helper()
	clk := f.clock()
	held := clk.HoldSleeps()
	uncapped := time.Duration(0)
	w, bodies := newRetryWorld(t, clk, ai.RetryPolicy{MaxRetries: 2, MaxRetryDelay: &uncapped})
	body := `{"error":{"message":"Rate limit reached"}}`
	enqueue(ev, w, provider.Reply{Status: 429, Header: map[string]string{"Content-Type": "application/json", "retry-after": "120"},
		Chunks: [][]byte{[]byte(body)}, Script: body})
	done := make(chan outcome, 1)
	go func() {
		done <- e.invoke(ctx, w, textScope("req-e05-backoff-"+e.name), textTarget(text), textRequest(text))
	}()
	d := waitValue(ev, "backoff wait began", held)
	ev.Check("waits the requested delay", d == 120*time.Second, "got %s", d)
	interrupt()
	o := waitValue(ev, "call ends", done)
	o.record(ev)
	ev.Check("no further attempt", len(w.provider.Requests()) == 1, "got %d", len(w.provider.Requests()))
	attempts := o.result.Metadata.Attempts
	ev.Check("the failed attempt is recorded with its planned delay", len(attempts) == 1 && attempts[0].HTTPStatus == 429 &&
		attempts[0].Code == ai.CodeRateLimited && attempts[0].RetryDelay == 120*time.Second, "attempts=%+v", attempts)
	ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
	return o
}

// checkRetryOutcome asserts sc's terminal: the text fixture's message on
// success, the expected classification and kept output otherwise.
func checkRetryOutcome(ev *evidence.Case, text textFixture, sc retryScenario, o outcome) {
	x := sc.Expect
	m := o.result.Message
	ev.Check("stop reason", m.StopReason == x.StopReason, "got %q want %q", m.StopReason, x.StopReason)
	if x.StopReason == ai.StopReasonStop {
		ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
		checkFinalMessage(ev, text, m)
		return
	}
	ev.Check("errorMessage", m.ErrorMessage == x.ErrorMessage, "got %q\nwant %q", m.ErrorMessage, x.ErrorMessage)
	ev.Check("errors.Is matches code and phase", errors.Is(o.err, &ai.Error{Code: x.Code, Phase: x.Phase}), "want %s/%s, got %v", x.Code, x.Phase, o.err)
	var ae *ai.Error
	if ev.Check("error is a classified *ai.Error", errors.As(o.err, &ae), "err=%v", o.err) {
		ev.Check("HTTP status", ae.HTTPStatus == x.HTTPStatus, "got %d want %d", ae.HTTPStatus, x.HTTPStatus)
		ev.Check("provider request id", ae.ProviderRequestID == x.ProviderRequestID, "got %q want %q", ae.ProviderRequestID, x.ProviderRequestID)
		ev.Check("retry-after", ae.RetryAfter == msDuration(x.RetryAfterMs), "got %s want %s", ae.RetryAfter, msDuration(x.RetryAfterMs))
	}
	want := x.Content
	if want == nil {
		want = []string{}
	}
	ev.Check("received content is kept, once", reflect.DeepEqual(contentSummary(m.Content), want), "got %q want %q", contentSummary(m.Content), want)
	ev.Check("response id", m.ResponseID == x.ResponseID, "got %q want %q", m.ResponseID, x.ResponseID)
	if o.streamed {
		starts := 0
		for _, e := range o.events {
			if e.Type() == ai.EventStart {
				starts++
			}
		}
		ev.Check("at most one start event", starts <= 1, "events=%q", eventSummary(o.events))
		ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
	}
}

// checkAttempts asserts the call's attempt records: one per HTTP attempt,
// each with its own AttemptID under the call's RequestID.
func checkAttempts(ev *evidence.Case, scope ai.CallScope, got []ai.Attempt, want []retryAttempt) {
	if !ev.Check("one record per attempt", len(got) == len(want), "got %+v want %+v", got, want) {
		return
	}
	ids := map[string]bool{}
	for i, a := range got {
		w := want[i]
		ids[a.AttemptID] = true
		ev.Check("attempt id belongs to the call", strings.HasPrefix(a.AttemptID, scope.RequestID), "attempt %d id %q", i+1, a.AttemptID)
		ev.Check("attempt record", a.HTTPStatus == w.HTTPStatus && a.Code == w.Code && a.ProviderRequestID == w.ProviderRequestID &&
			a.RetryDelay == msDuration(w.RetryDelayMs), "attempt %d got %+v want %+v", i+1, a, w)
	}
	ev.Check("attempt ids are distinct", len(ids) == len(got), "got %+v", got)
}

// checkAttemptRequests asserts the Provider received one request per key, in
// order, and that the attempts of one call sent the same body.
func checkAttemptRequests(ev *evidence.Case, w *world, keys ...tenantKey) {
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	if !ev.Check("inference requests", len(reqs) == len(keys), "got %d want %d", len(reqs), len(keys)) {
		return
	}
	for i, r := range reqs {
		ev.Check("pinned key used", r.KeyAlias == keys[i].alias, "request %d got %q want %q", i+1, r.KeyAlias, keys[i].alias)
		ev.Check("same endpoint", r.Path == "/v1/responses", "request %d path %q", i+1, r.Path)
		if i > 0 && keys[i] == keys[0] {
			ev.Check("attempts resend the same body", bytes.Equal(r.Body, reqs[0].Body), "request %d differs", i+1)
		}
	}
}

func checkSleeps(ev *evidence.Case, got []time.Duration, wantMs []float64) {
	want := make([]time.Duration, len(wantMs))
	for i, ms := range wantMs {
		want[i] = msDuration(ms)
	}
	ev.Record("sleeps", got)
	ev.Check("backoff waits", reflect.DeepEqual(got, want) || len(got) == 0 && len(want) == 0, "got %v want %v", got, want)
}

func retryScenarioByID(t *testing.T, f retryFixture, id string) retryScenario {
	t.Helper()
	for _, sc := range f.Scenarios {
		if sc.ID == id {
			return sc
		}
	}
	t.Fatalf("retry.json has no scenario %q", id)
	return retryScenario{}
}
