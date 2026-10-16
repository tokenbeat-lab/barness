package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevContextBorrowedImageBudget(t *testing.T) {
	for _, name := range []string{"map", "raw", "typed"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC04-borrowed-budget-"+name)
			u := newJevContextWorld(t, ev, func(c *ai.Config) { c.Policy.MaxImageBytes = 3000 })
			req := jevContextInput(t)
			// 512 KiB of input is already allocated by the trusted Host. A known
			// decoded single-file violation must fail before serializing/copying it.
			data := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", 256<<10)))
			var before, after runtime.MemStats
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				asset := map[string]any{"type": "image", "format": "png", "data_base64": data}
				switch name {
				case "map":
					p.Body["assets"] = map[string]any{"asset": asset}
				case "raw":
					p.Body["assets"] = json.RawMessage(mustMarshal(t, map[string]any{"asset": asset}))
				case "typed":
					p.Body["assets"] = map[string]struct {
						Type   string `json:"type"`
						Format string `json:"format"`
						Data   string `json:"data_base64"`
					}{"asset": {Type: "image", Format: "png", Data: data}}
				}
				p.Body["state"] = map[string]any{"state": "ok", "images": []string{"asset"}}
				runtime.ReadMemStats(&before)
				return ai.KeepPayload(), nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-jev-borrowed-"+name), jevContextTarget(), req, jevContextOptions())
			runtime.ReadMemStats(&after)
			ev.Check("known bytes rejected before encoded copies", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && after.TotalAlloc-before.TotalAlloc < 128<<10, "got %v allocated %d", err, after.TotalAlloc-before.TotalAlloc)
			recordJevContext(ev, u, req, res, err)
			ev.Record("input-recipe", map[string]any{"decodedXBytes": 256 << 10, "encoding": name})
		})
	}
}
