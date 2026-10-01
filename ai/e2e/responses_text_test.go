package e2e

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestResponsesText is the first tracer bullet: OpenAI × Responses plain text
// through every public entry point (P01 text path, E01 success terminal).
func TestResponsesText(t *testing.T) {
	// The four public entry points must agree on request, events and result.
	for _, e := range entries {
		t.Run(e.name, func(t *testing.T) {
			ev, w, f := startTextCase(t, "P01-E01-text-"+e.name, provider.FramingLF)
			scope := textScope("req-" + e.name)
			res := e.invoke(ev, w, f, scope)
			checkMetadata(ev, res.Metadata, scope, f)
			checkTextRequest(ev, w, f, tenantA)
		})
	}

	// Every SSE framing decodes to the same events and final message.
	for _, framing := range provider.Framings {
		t.Run("framing-"+string(framing), func(t *testing.T) {
			ev, w, f := startTextCase(t, "P01-E01-text-framing-"+string(framing), framing)
			scope := textScope("req-framing-" + string(framing))
			s := w.client.Stream(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
			res := checkStreamedText(ev, f, s)
			checkMetadata(ev, res.Metadata, scope, f)
			checkTextRequest(ev, w, f, tenantA)
		})
	}

	t.Run("stream-returns-before-resolution", func(t *testing.T) {
		ev, w, f := startTextCase(t, "P01-E01-stream-returns-before-resolution", provider.FramingLF)
		release := w.host.HoldBindings()
		defer release()
		scope := textScope("req-stream-early")

		returned := make(chan *ai.Stream, 1)
		ctx := ctxFor(t)
		go func() { returned <- w.client.Stream(ctx, scope, textTarget(f), textRequest(f), nil) }()
		var s *ai.Stream
		select {
		case s = <-returned:
			ev.Check("Stream returned while the binding resolver was blocked", true, "")
		case <-time.After(5 * time.Second):
			ev.Check("Stream returned while the binding resolver was blocked", false, "Stream blocked on resolution")
			return
		}
		ev.Check("no provider request before resolution", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
		release()
		checkStreamedText(ev, f, s)
	})
}

// entry is one public entry point.
type entry struct {
	name   string
	invoke func(ev *evidence.Case, w *world, f textFixture, scope ai.CallScope) ai.Result
}

var entries = []entry{
	{"stream-full", func(ev *evidence.Case, w *world, f textFixture, scope ai.CallScope) ai.Result {
		return checkStreamedText(ev, f, w.client.Stream(ctxFor(ev.T()), scope, textTarget(f), textRequest(f), nil))
	}},
	{"stream-simple", func(ev *evidence.Case, w *world, f textFixture, scope ai.CallScope) ai.Result {
		return checkStreamedText(ev, f, w.client.StreamSimple(ctxFor(ev.T()), scope, textTarget(f), textRequest(f), ai.SimpleOptions{}))
	}},
	{"complete-full", func(ev *evidence.Case, w *world, f textFixture, scope ai.CallScope) ai.Result {
		res, err := w.client.Complete(ctxFor(ev.T()), scope, textTarget(f), textRequest(f), ai.ResponsesOptions{})
		return checkCompleted(ev, f, res, err)
	}},
	{"complete-simple", func(ev *evidence.Case, w *world, f textFixture, scope ai.CallScope) ai.Result {
		res, err := w.client.CompleteSimple(ctxFor(ev.T()), scope, textTarget(f), textRequest(f), ai.SimpleOptions{})
		return checkCompleted(ev, f, res, err)
	}},
}

// textScope is tenant A's trusted scope, with the optional ActorID and JobID
// set so their preservation is observable.
func textScope(requestID string) ai.CallScope {
	return ai.CallScope{TenantID: tenantA.tenant, RequestID: requestID, ActorID: "actor-7", JobID: "job-42"}
}

func checkCompleted(ev *evidence.Case, f textFixture, res ai.Result, err error) ai.Result {
	ev.Record("result", res)
	ev.Check("Complete error is nil", err == nil, "err=%v", err)
	checkFinalMessage(ev, f, res.Message)
	return res
}

// checkMetadata asserts the immutable call attribution on the Result.
func checkMetadata(ev *evidence.Case, got ai.CallMetadata, scope ai.CallScope, f textFixture) {
	want := ai.CallMetadata{
		TenantID:   scope.TenantID,
		RequestID:  scope.RequestID,
		ActorID:    scope.ActorID,
		JobID:      scope.JobID,
		BindingID:  "primary",
		Resolved:   true,
		ProviderID: ai.ProviderOpenAI,
		API:        ai.APIOpenAIResponses,
		ModelID:    f.Model,
	}
	ev.Check("call metadata", got == want, "got %+v\nwant %+v", got, want)
}

// checkStreamedText consumes s fully and asserts the event sequence, the
// terminal message and the Scanner-style error contract.
func checkStreamedText(ev *evidence.Case, f textFixture, s *ai.Stream) ai.Result {
	defer s.Close()
	var events []ai.Event
	errDuring := error(nil)
	for s.Next() {
		events = append(events, s.Event())
		if err := s.Err(); err != nil && errDuring == nil {
			errDuring = err
		}
	}
	errAfter := s.Err()
	res, err := s.Result()
	ev.Record("events", eventsJSON(events))
	ev.Record("result", res)

	want := []string{"start", "text_start#0"}
	for _, d := range f.Expect.Deltas {
		want = append(want, "text_delta#0:"+d)
	}
	want = append(want, "text_end#0:"+f.Expect.Text, "done:"+f.Expect.StopReason)
	got := eventSummary(events)
	ev.Check("event sequence", reflect.DeepEqual(got, want), "got %q\nwant %q", got, want)
	ev.Check("Err is nil before the terminal", errDuring == nil, "got %v", errDuring)
	ev.Check("Err is nil after a successful terminal", errAfter == nil, "got %v", errAfter)
	ev.Check("Result error is nil", err == nil, "got %v", err)
	if len(events) > 0 {
		done, ok := events[len(events)-1].(ai.DoneEvent)
		ev.Check("done and Result carry the same final message", ok && reflect.DeepEqual(done.Message, res.Message),
			"done=%+v result=%+v", done.Message, res.Message)
	}
	checkFinalMessage(ev, f, res.Message)
	return res
}

// eventSummary renders events compactly for order assertions.
func eventSummary(events []ai.Event) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		switch e := e.(type) {
		case ai.StartEvent:
			out = append(out, "start")
		case ai.TextStartEvent:
			out = append(out, fmt.Sprintf("text_start#%d", e.ContentIndex))
		case ai.TextDeltaEvent:
			out = append(out, fmt.Sprintf("text_delta#%d:%s", e.ContentIndex, e.Delta))
		case ai.TextEndEvent:
			out = append(out, fmt.Sprintf("text_end#%d:%s", e.ContentIndex, e.Content))
		case ai.DoneEvent:
			out = append(out, "done:"+string(e.Reason))
		case ai.ErrorEvent:
			out = append(out, "error:"+string(e.Reason))
		default:
			out = append(out, fmt.Sprintf("unexpected:%T", e))
		}
	}
	return out
}

// eventsJSON keeps each event's type next to its payload for the evidence bundle.
func eventsJSON(events []ai.Event) []map[string]any {
	out := make([]map[string]any, 0, len(events))
	for _, e := range events {
		out = append(out, map[string]any{"type": e.Type(), "event": e})
	}
	return out
}

// startTextCase loads the basic text fixture, scripts the provider with it in
// the given framing and records both in the evidence bundle.
func startTextCase(t *testing.T, caseID string, framing provider.Framing) (*evidence.Case, *world, textFixture) {
	t.Helper()
	ev := run.Case(t, caseID)
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("fixture.json", raw)
	w := newWorld(t, tenantA)
	enqueue(ev, w, sseReply(t, f, framing))
	return ev, w, f
}

func ctxFor(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func textTarget(f textFixture) ai.Target {
	return ai.Target{BindingID: "primary", ModelID: f.Model}
}

func textRequest(f textFixture) ai.Request {
	return ai.Request{
		SystemPrompt: f.SystemPrompt,
		Messages:     []ai.Message{ai.UserText(f.User)},
	}
}

func checkFinalMessage(ev *evidence.Case, f textFixture, m ai.AssistantMessage) {
	ev.Check("stop reason", string(m.StopReason) == f.Expect.StopReason, "got %q", m.StopReason)
	ev.Check("no error message", m.ErrorMessage == "", "got %q", m.ErrorMessage)
	ev.Check("message identity", m.API == ai.APIOpenAIResponses && m.Provider == ai.ProviderOpenAI && m.Model == f.Model,
		"api=%q provider=%q model=%q", m.API, m.Provider, m.Model)
	ev.Check("response id", m.ResponseID == f.Expect.ResponseID, "got %q", m.ResponseID)
	ev.Check("usage", m.Usage == f.Expect.Usage, "got %+v want %+v", m.Usage, f.Expect.Usage)
	if !ev.Check("single text block", len(m.Content) == 1, "content=%+v", m.Content) {
		return
	}
	text, ok := m.Content[0].(ai.Text)
	ev.Check("text content", ok && text.Text == f.Expect.Text, "got %+v", m.Content[0])
	ev.Check("text signature", ok && text.Signature == f.Expect.TextSignature, "got %q", text.Signature)
}

// checkTextRequest asserts the single inference request the provider received.
func checkTextRequest(ev *evidence.Case, w *world, f textFixture, key tenantKey) {
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	if !ev.Check("exactly one inference request", len(reqs) == 1, "got %d", len(reqs)) {
		return
	}
	r := reqs[0]
	want := f.Expect.Request
	ev.Check("method", r.Method == want.Method, "got %q", r.Method)
	ev.Check("path", r.Path == want.Path, "got %q", r.Path)
	ev.Check("query", r.Query == want.Query, "got %q", r.Query)
	ev.Check("tenant key used", r.KeyAlias == key.alias, "got %q want %q", r.KeyAlias, key.alias)
	ev.Check("request body", jsonEqual(r.Body, want.Body), "got %s\nwant %s", r.Body, want.Body)
}
