package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestGeminiFailures is P03/E02: every failure and truncation terminal of
// testdata/gemini/failures.json through the public Client. Each ends with a
// complete message and, when it failed, a classified error that Result,
// Complete and Err agree on; the stream's terminals also run through Result
// without Next and through Complete.
func TestGeminiFailures(t *testing.T) {
	f, raw := loadFixture(t, geminiProtocol, "failures.json")
	for _, sc := range f.Scenarios {
		modes := []callMode{modeStream}
		if sc.CancelAfterEvents == 0 {
			modes = append(modes, modeResult, modeComplete)
		}
		for _, mode := range modes {
			t.Run(sc.ID+"/"+string(mode), func(t *testing.T) {
				ev := run.Case(t, "P03-E02-"+sc.ID+"-"+string(mode))
				_, o := runScenario(t, ev, sc, raw, mode)
				if sc.ID == "length-truncated-tool" {
					_, err := ai.ValidateToolCall(o.result.Message, 0, sc.Context.Tools)
					ev.Check("a truncated tool call is not executable", err != nil, "ValidateToolCall accepted it")
				}
			})
		}
	}

	// A deadline that passes while the provider holds the stream open aborts
	// the call as deadline_exceeded, keeping what arrived.
	t.Run("deadline-mid-stream", func(t *testing.T) {
		ev := run.Case(t, "P03-E02-deadline-mid-stream")
		sc := scenarioByID(t, f, "canceled-mid-stream")
		ev.Fixture("fixture.json", raw)
		w := scenarioWorld(t, sc)
		enqueue(ev, w, sc.replies(t)...)
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		res, err := w.client.Complete(ctx, textScope("req-deadline"), sc.target(), sc.request(t), ai.GeminiOptions{})
		ev.Record("result", res)
		var aiErr *ai.Error
		ev.Check("deadline_exceeded in the stream phase", errors.As(err, &aiErr) && aiErr.Code == ai.CodeDeadlineExceeded &&
			aiErr.Phase == ai.PhaseStream, "got %v", err)
		ev.Check("message ends aborted with pi's text", res.Message.StopReason == ai.StopReasonAborted &&
			res.Message.ErrorMessage == "Request was aborted", "got %q %q", res.Message.StopReason, res.Message.ErrorMessage)
		text, _ := firstText(res.Message)
		ev.Check("content received before the deadline is kept", text == "Hel", "got %q", text)
	})

	// A setup failure ends the call before start without contacting the
	// provider: here a pi model the catalog leaves out (a Live API model).
	t.Run("setup-failure-before-start", func(t *testing.T) {
		ev := run.Case(t, "P03-E02-setup-failure-before-start")
		w := newWorld(t, tenantA)
		s := w.client.Stream(ctxFor(t), textScope("req-setup"), ai.Target{BindingID: "gemini", ModelID: "gemini-3.1-flash-live-preview"},
			ai.Request{Messages: []ai.Message{ai.UserText("hi")}}, nil)
		var events []ai.Event
		for s.Next() {
			events = append(events, s.Event())
		}
		res, err := s.Result()
		ev.Record("result", res)
		ev.Check("only the error terminal", len(events) == 1 && events[0].Type() == ai.EventError, "got %q", eventSummary(events))
		ev.Check("tenant_denied", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied}), "got %v", err)
		ev.Check("message ends as error", res.Message.StopReason == ai.StopReasonError && res.Message.ErrorMessage != "", "got %+v", res.Message)
		ev.Check("no request reached the provider", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	})
}
