package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Independent calls have independent elapsed times. Check the terminal value
// before normalizing only that number for cross-run message comparisons; the
// field remains present, and the raw outcomes remain in the evidence bundle.
func messageWithoutElapsed(ev *evidence.Case, label string, m ai.AssistantMessage) ai.AssistantMessage {
	value, ok := m.DurationMs.Get()
	ev.Check(label+" has a nonnegative integer duration", ok && value >= 0, "duration=%v", m.DurationMs)
	m.DurationMs = ai.Value[int64](0)
	return m
}

// pi 1.1.0 migration: every terminal (success or setup failure) is timed,
// including Complete. Pending snapshots may be untimed; zero is a value.
// The controlled callback delay proves actual elapsed time is measured.
func TestChatDuration(t *testing.T) {
	for _, e := range outcomeEntries {
		for _, failing := range []bool{false, true} {
			name := e.name + "-success"
			if failing {
				name = e.name + "-failure"
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "E11-chat-duration-"+name)
				f, raw := loadTextFixture(t, "text-basic.json")
				ev.Fixture("text-basic.json", raw)
				w := newWorld(t, tenantA)
				enqueue(ev, w, sseReply(t, f, provider.FramingLF))
				hooked := w.client.WithHooks(ai.Hooks{OnPayload: func(ctx context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
					timer := time.NewTimer(15 * time.Millisecond)
					defer timer.Stop()
					select {
					case <-timer.C:
					case <-ctx.Done():
						return ai.KeepPayload(), ctx.Err()
					}
					if failing {
						return ai.ReplacePayload(map[string]any{"model": "unauthorized"}), nil
					}
					return ai.KeepPayload(), nil
				}})
				start := time.Now()
				var o outcome
				target, req, scope := textTarget(f), textRequest(f), textScope("req-duration-"+name)
				switch e.name {
				case "stream-full":
					o = drain(hooked.Stream(ctxFor(t), scope, target, req, nil))
				case "stream-simple":
					o = drain(hooked.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{}))
				case "complete-full":
					r, err := hooked.Complete(ctxFor(t), scope, target, req, nil)
					o = outcome{result: r, err: err}
				case "complete-simple":
					r, err := hooked.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{})
					o = outcome{result: r, err: err}
				}
				o.record(ev)
				var message struct {
					Duration *int64 `json:"durationMs"`
				}
				mustUnmarshal(t, mustMarshal(t, o.result.Message), &message)
				ev.Check("outcome", (o.err != nil) == failing, "err=%v", o.err)
				ev.Check("monotonic duration includes callback", message.Duration != nil && *message.Duration >= 14 && *message.Duration <= time.Since(start).Milliseconds()+1, "duration=%v", message.Duration)
				for _, event := range o.events {
					switch event := event.(type) {
					case ai.DoneEvent:
						ev.Check("terminal duration equals result", jsonEqual(mustMarshal(t, event.Message), mustMarshal(t, o.result.Message)), "terminal differs")
					case ai.ErrorEvent:
						ev.Check("error duration equals result", jsonEqual(mustMarshal(t, event.Message), mustMarshal(t, o.result.Message)), "terminal differs")
					}
				}
				ev.Record("duration-input", json.RawMessage(`{"callbackDelayMs":15}`))
			})
		}
	}
}
