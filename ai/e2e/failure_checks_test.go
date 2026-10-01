package e2e

import (
	"context"
	"errors"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// newTrackedWorld is tenant A's world whose transport counts unclosed
// response bodies.
func newTrackedWorld(t *testing.T) (*world, *provider.BodyTracker) {
	t.Helper()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	w := newWorldWith(t, func(c *ai.Config) { c.Transport = bodies }, tenantA)
	return w, bodies
}

// deadEndpoint is a loopback endpoint nothing listens on any more.
func deadEndpoint() string {
	srv := httptest.NewServer(nil)
	url := srv.URL
	srv.Close()
	return url + "/v1"
}

// drainClosingAfter reads s and, once n events were read, has another
// goroutine Close it while the consumer goes on (and blocks) in Next. Close
// must release the blocked Next without any further call.
func drainClosingAfter(s *ai.Stream, n int) outcome {
	var once sync.Once
	read, closed := make(chan struct{}), make(chan struct{})
	signal := func() { once.Do(func() { close(read) }) }
	go func() {
		<-read
		_ = s.Close()
		close(closed)
	}()
	o := outcome{streamed: true}
	for s.Next() {
		o.events = append(o.events, s.Event())
		if len(o.events) == n {
			signal()
		}
	}
	signal()
	<-closed
	o.streamErr = s.Err()
	o.result, o.err = s.Result()
	return o
}

// callWhileProviderHolds starts a call on e whose request the provider
// receives but does not answer, runs interrupt once the request is in
// flight, and returns the outcome.
func callWhileProviderHolds(t *testing.T, ev *evidence.Case, e outcomeEntry, ctx context.Context, interrupt func()) outcome {
	t.Helper()
	w, bodies := newTrackedWorld(t)
	f, raw := loadFailureFixture(t)
	ev.Fixture("failures.json", raw)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	enqueue(ev, w, provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "text/event-stream"},
		OnReceive: func() { close(entered); <-release }})
	// Registered after the world, so it runs before the provider shuts down.
	t.Cleanup(func() { once.Do(func() { close(release) }) })

	done := make(chan outcome, 1)
	go func() { done <- e.invoke(ctx, w, textScope("req-held-"+e.name), f.target(), f.request()) }()
	waitFor(ev, "provider received the request", entered)
	interrupt()
	o := waitValue(ev, "call ends", done)
	once.Do(func() { close(release) })
	o.record(ev)
	ev.Check("exactly one inference request", len(w.provider.Requests()) == 1, "got %d", len(w.provider.Requests()))
	ev.Check("every response body is closed", bodies.Open() == 0, "%d open", bodies.Open())
	return o
}

// checkInterrupted asserts a canceled or timed-out call: its terminal, its
// classification and the Scanner-style error contract.
func checkInterrupted(ev *evidence.Case, o outcome, stop ai.StopReason, code ai.Code, phase ai.Phase, msg string) {
	m := o.result.Message
	ev.Check("stop reason", m.StopReason == stop, "got %q want %q", m.StopReason, stop)
	ev.Check("errors.Is matches code and phase", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}), "want %s/%s, got %v", code, phase, o.err)
	ev.Check("errorMessage", m.ErrorMessage == msg, "got %q want %q", m.ErrorMessage, msg)
	ev.Check("message still carries its timestamp", m.Timestamp > 0, "got %d", m.Timestamp)
	if o.streamed {
		got := eventSummary(o.events)
		ev.Check("stream ends with one terminal", len(got) > 0 && got[len(got)-1] == "error:"+string(stop), "got %q", got)
		ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
	}
}

// checkFailureOutcome asserts one E02 scenario's outcome against the fixture.
func checkFailureOutcome(ev *evidence.Case, f failureFixture, sc failureScenario, o outcome) {
	x := sc.Expect
	m := o.result.Message
	ev.Check("stop reason", m.StopReason == x.StopReason, "got %q want %q", m.StopReason, x.StopReason)
	ev.Check("raw stop reason", m.RawStopReason == x.RawStopReason, "got %q want %q", m.RawStopReason, x.RawStopReason)
	ev.Check("errorMessage", m.ErrorMessage == x.ErrorMessage, "got %q\nwant %q", m.ErrorMessage, x.ErrorMessage)
	wantContent := x.Content
	if wantContent == nil {
		wantContent = []string{}
	}
	ev.Check("received content is kept", reflect.DeepEqual(contentSummary(m.Content), wantContent), "got %q want %q", contentSummary(m.Content), wantContent)
	ev.Check("response id", m.ResponseID == x.ResponseID, "got %q want %q", m.ResponseID, x.ResponseID)
	ev.Check("usage is kept", m.Usage == x.Usage, "got %+v want %+v", m.Usage, x.Usage)
	ev.Check("message identity", m.API == ai.APIOpenAIResponses && m.Provider == ai.ProviderOpenAI && m.Model == f.Model,
		"api=%q provider=%q model=%q", m.API, m.Provider, m.Model)
	ev.Check("message carries its timestamp", m.Timestamp > 0, "got %d", m.Timestamp)
	ev.Check("call identity resolved", o.result.Metadata.Resolved, "metadata=%+v", o.result.Metadata)

	failed := x.StopReason == ai.StopReasonError || x.StopReason == ai.StopReasonAborted
	if failed {
		ev.Check("errors.Is matches code and phase", errors.Is(o.err, &ai.Error{Code: x.Code, Phase: x.Phase}), "want %s/%s, got %v", x.Code, x.Phase, o.err)
		var ae *ai.Error
		if ev.Check("error is a classified *ai.Error", errors.As(o.err, &ae), "err=%v", o.err) {
			ev.Check("HTTP status", ae.HTTPStatus == x.HTTPStatus, "got %d want %d", ae.HTTPStatus, x.HTTPStatus)
			ev.Check("provider request id", ae.ProviderRequestID == x.ProviderRequestID, "got %q want %q", ae.ProviderRequestID, x.ProviderRequestID)
			want := time.Duration(x.RetryAfterMs) * time.Millisecond
			ev.Check("retry-after", ae.RetryAfter == want, "got %s want %s", ae.RetryAfter, want)
			ev.Check("error message equals errorMessage", ae.Message == m.ErrorMessage, "error=%q message=%q", ae.Message, m.ErrorMessage)
		}
	} else {
		ev.Check("a successful terminal returns a nil error", o.err == nil, "got %v", o.err)
	}
	if o.streamed {
		want := f.wantEvents(sc)
		got := eventSummary(o.events)
		ev.Check("event sequence", reflect.DeepEqual(got, want), "got %q\nwant %q", got, want)
		ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
		if n := len(o.events); n > 0 {
			switch e := o.events[n-1].(type) {
			case ai.ErrorEvent:
				ev.Check("error event carries the Result's message and error", failed && reflect.DeepEqual(e.Message, m) && error(e.Err) == o.err, "event=%+v", e)
			case ai.DoneEvent:
				ev.Check("done event carries the Result's message", !failed && reflect.DeepEqual(e.Message, m), "event=%+v", e)
			}
		}
	}
	checkNoSecrets(ev, o)
}

// checkProviderSaw asserts the provider received the expected number of
// requests, each with the tenant's own key: a failure never switches keys or
// falls back to another identity.
func checkProviderSaw(ev *evidence.Case, w *world, sc failureScenario) {
	want := 1
	if sc.Expect.Requests != nil {
		want = *sc.Expect.Requests
	}
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	ev.Check("inference requests", len(reqs) == want, "got %d want %d", len(reqs), want)
	for _, r := range reqs {
		ev.Check("tenant key used", r.KeyAlias == tenantA.alias, "got %q", r.KeyAlias)
	}
}
