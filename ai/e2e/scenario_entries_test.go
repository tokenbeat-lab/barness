package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// scenarioEntries are the four public entry points on a scenario
// protocol's binding, reporting whatever happens. The full entries pass
// empty, the protocol's options (outcomeEntries' complete-full passes
// Responses options, which another binding refuses).
func scenarioEntries(empty ai.Options) []outcomeEntry {
	return []outcomeEntry{
		{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.Stream(ctx, scope, target, req, nil))
		}},
		{"stream-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			return drain(w.client.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{}))
		}},
		{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.Complete(ctx, scope, target, req, empty)
			return outcome{result: res, err: err}
		}},
		{"complete-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
			res, err := w.client.CompleteSimple(ctx, scope, target, req, ai.SimpleOptions{})
			return outcome{result: res, err: err}
		}},
	}
}

// scenarioText is a protocol's text.json "text" scenario: the plain text
// turn its resource scenarios run, with its fixture bytes. Every protocol's
// produces "Hello, world!".
type scenarioText struct {
	sc  fixtureScenario
	raw []byte
}

func loadScenarioText(t *testing.T, proto *fixtureProtocol) scenarioText {
	t.Helper()
	f, raw := loadFixture(t, proto, "text.json")
	return scenarioText{sc: scenarioByID(t, f, "text"), raw: raw}
}

func (a scenarioText) target() ai.Target { return a.sc.target() }

func (a scenarioText) request(t *testing.T) ai.Request { return a.sc.request(t) }

// chunks lays the scenario out one SSE frame per chunk.
func (a scenarioText) chunks(t *testing.T) [][]byte {
	return a.sc.proto.sse(t, a.sc.Events, provider.FramingLF).Chunks
}

func (a scenarioText) reply(t *testing.T) provider.Reply { return a.sc.replies(t)[0] }

// text is the final text the scenario produces.
func (a scenarioText) text() string { return "Hello, world!" }

// firstDelta is the index of the first chunk carrying text.
func (a scenarioText) firstDelta(t *testing.T) int {
	t.Helper()
	for i, c := range a.chunks(t) {
		if strings.Contains(string(c), a.sc.proto.textMarker) {
			return i
		}
	}
	t.Fatal("no text delta in text.json \"text\"")
	return 0
}

// checkScenarioTextSucceeded asserts a call ended exactly as the text scenario
// expects.
func checkScenarioTextSucceeded(ev *evidence.Case, a scenarioText, o outcome) {
	m := o.result.Message
	ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	ev.Check("stop reason", m.StopReason == ai.StopReasonStop, "got %q", m.StopReason)
	ev.Check("message identity", m.API == a.sc.proto.api && m.Provider == a.sc.proto.provider && m.Model == a.sc.Model,
		"api=%q provider=%q model=%q", m.API, m.Provider, m.Model)
	ev.Check("text", textOf(m) == a.text(), "got %q", textOf(m))
	ev.Check("usage", m.Usage == *a.sc.Expect.Usage, "got %+v", m.Usage)
	if o.streamed {
		got := eventSummary(o.events)
		ev.Check("stream ends with done", len(got) > 0 && got[len(got)-1] == "done:stop", "got %q", got)
	}
}

// updateBindingOf rewrites k's tenant's stored binding bindingID.
func (w *world) updateBindingOf(k tenantKey, bindingID string, change func(*ai.Binding)) {
	b := w.host.Binding(k.tenant, bindingID)
	change(&b)
	w.host.PutBindingAt(k.tenant, bindingID, b)
}

// holdScenario starts a Stream for k on a's binding whose attempt keeps its
// permit: the provider sends the first frames through the first text delta
// and holds the connection. It returns once the provider holds.
func holdScenario(t *testing.T, ev *evidence.Case, aw admissionWorld, a scenarioText, ctx context.Context, k tenantKey, requestID string) *ai.Stream {
	t.Helper()
	held := make(chan struct{})
	r := heldReply(200, "text/event-stream", joinChunks(a.chunks(t)[:a.firstDelta(t)+1]))
	r.OnHold = func() { close(held) }
	aw.provider.Enqueue(r)
	s := aw.client.Stream(ctx, scopeFor(k, requestID), a.target(), a.request(t), nil)
	t.Cleanup(func() { _ = s.Close() })
	waitFor(ev, "provider holds "+requestID, held)
	return s
}
