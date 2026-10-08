package e2e

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// The host owns the supplied large value. The library must reject its known
// size before making encoded/decoded copies. This intentionally measures a
// generous allocation ceiling rather than an internal function or call order.
func TestUnaryCallbackAllocationBound(t *testing.T) {
	large := strings.Repeat("x", 8<<20)
	numbers := make([]int, 2<<20)
	emptyStrings := make([]string, 2<<20)
	stringKeys := map[unaryShortText]string{unaryShortText(large): "ok"}
	for _, name := range []string{"state", "nested-state", "question", "extra-request-field", "numbers", "empty-strings", "string-key"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E08-unary-callback-allocation-"+name)
			u := newUnaryWorld(t, nil)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "string-key":
					p.Body["state"] = stringKeys
				case "numbers":
					p.Body["state"] = numbers
				case "empty-strings":
					p.Body["state"] = emptyStrings
				case "state":
					p.Body["state"] = large
				case "nested-state":
					p.Body["state"] = map[string]any{"nested": []any{large}}
				case "question":
					p.Body["questions"].(map[string]any)["intent"].(map[string]any)["instructions"] = large
				case "extra-request-field":
					p.Body["padding"] = large
				}
				return ai.KeepPayload(), nil
			}}
			// Construct input and hooks before measuring; only the public call counts.
			input := classifierInput(t, u.sc)
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-allocation-"+name), u.sc.target(), input, nil)
			runtime.ReadMemStats(&after)
			allocated := after.TotalAlloc - before.TotalAlloc
			ev.Record("allocated-bytes", allocated)
			code := ai.CodeCallbackFailed
			if name == "extra-request-field" {
				code = ai.CodeResourceLimit
			}
			checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
			ev.Check("known oversized input refused before copies", allocated < 4<<20, "allocated %d bytes for a known oversized host value", allocated)
			ev.Check("no admission or HTTP", len(u.adm.Requests()) == 0 && len(u.provider.Requests()) == 0, "sent %d", len(u.provider.Requests()))
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
