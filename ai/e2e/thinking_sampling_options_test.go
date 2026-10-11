package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

var thinkingSamplingProtocols = []struct {
	prefix string
	proto  *fixtureProtocol
}{
	{"P01", responsesProtocol}, {"P04", chatProtocol}, {"P05", deepseekResponsesProtocol}, {"P06", deepseekChatProtocol},
}

// S01-S17 use the existing option/scenario seam, complete requests and final
// results. S01-S03 additionally run all four WithHooks entries; callbacks must
// see exactly the final merged payload. Inputs/catalogs stay unchanged.
func TestThinkingSamplingOptions(t *testing.T) {
	for _, p := range thinkingSamplingProtocols {
		f, raw := loadFixture(t, p.proto, "thinking-sampling.json")
		for _, sc := range f.Scenarios {
			for _, mode := range []callMode{modeStream, modeComplete} {
				hooks := []bool{false}
				if strings.Contains(sc.ID, "S01") || strings.Contains(sc.ID, "S02") || strings.Contains(sc.ID, "S03") || strings.Contains(sc.ID, "S07") || strings.Contains(sc.ID, "S08") || strings.Contains(sc.ID, "S09") || strings.Contains(sc.ID, "S11") {
					hooks = append(hooks, true)
				}
				for _, hooked := range hooks {
					name := sc.ID + "-" + string(mode)
					if hooked {
						name += "-hooks"
					}
					t.Run(p.proto.dir+"/"+name, func(t *testing.T) {
						ev := run.Case(t, p.prefix+"-E04-"+name)
						ev.Fixture("fixture.json", raw)
						w := scenarioWorld(t, sc)
						enqueue(ev, w, sc.replies(t)...)
						before := mustMarshal(t, w.client.Catalog())
						count := 0
						var o outcome
						if !hooked {
							o = invokeScenario(t, w, sc, mode, "req-"+name)
						} else {
							o = invokeThinkingSamplingHooks(t, w, sc, mode, ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
								count++
								ev.Check("callback sees resolved sampling", jsonEqual(mustMarshal(t, payload.Body), sc.Expect.Request.Body), "got %s", mustMarshal(t, payload.Body))
								return ai.KeepPayload(), nil
							}})
							ev.Check("callback once", count == 1, "count=%d", count)
						}
						o.record(ev)
						checkScenario(ev, w, sc, mode, o)
						ev.Check("catalog unchanged", jsonEqual(before, mustMarshal(t, w.client.Catalog())), "catalog changed")
						ev.Record("catalog", w.client.Catalog())
					})
				}
			}
		}
	}
}

func invokeThinkingSamplingHooks(t *testing.T, w *world, sc fixtureScenario, mode callMode, hooks ai.Hooks) outcome {
	t.Helper()
	return invokeThinkingSamplingAt(t, ctxFor(t), w, sc, mode, hooks, textScope("req-"+sc.ID))
}

func invokeThinkingSamplingAt(t *testing.T, ctx context.Context, w *world, sc fixtureScenario, mode callMode, hooks ai.Hooks, scope ai.CallScope) outcome {
	t.Helper()
	c := w.client.WithHooks(hooks)
	target, req := sc.target(), sc.request(t)
	if mode == modeComplete {
		if sc.entry() == "streamSimple" {
			r, err := c.CompleteSimple(ctx, scope, target, req, sc.simpleOptions(t))
			return outcome{result: r, err: err}
		}
		r, err := c.Complete(ctx, scope, target, req, sc.fullOptions(t))
		return outcome{result: r, err: err}
	}
	if sc.entry() == "streamSimple" {
		return drain(c.StreamSimple(ctx, scope, target, req, sc.simpleOptions(t)))
	}
	return drain(c.Stream(ctx, scope, target, req, sc.fullOptions(t)))
}
