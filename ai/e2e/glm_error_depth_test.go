package e2e

import (
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// A bounded error body must not expand into quadratic subtree reparsing.
// Only one documented error envelope is meaningful; deeper codes are unknown.
func TestGLMDeepErrorEnvelope(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			t.Run(r.name+"/"+model, func(t *testing.T) {
				ev := run.Case(t, "P0-GLM13-error-depth-"+r.name+"-"+model)
				w := newGLMSharedWorld(t, func(c *ai.Config) { c.Policy.MaxErrorBodyBytes = 128 << 10 })
				const depth = 5000
				body := []byte(strings.Repeat(`{"error":`, depth) + `{"code":"1210"}` + strings.Repeat("}", depth))
				ev.Fixture("malicious-error.json", body)
				reply := provider.Reply{Status: 400, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{body}}
				w.provider.EnqueueAt(r.prefix(glmA), reply)
				var before, after runtime.MemStats
				runtime.ReadMemStats(&before)
				started := time.Now()
				result, err := w.client.Complete(ctxFor(t), textScope("glm13-error-depth"), r.target(model), glmResponsesRequest(), r.options())
				elapsed := time.Since(started)
				runtime.ReadMemStats(&after)
				allocated := after.TotalAlloc - before.TotalAlloc
				ev.Record("result", result)
				ev.Record("error", errString(err))
				ev.Record("parse-resource-envelope", map[string]any{"bodyBytes": len(body), "depth": depth, "totalAllocatedBytes": allocated, "elapsedNanos": elapsed.Nanoseconds(), "maxAllocatedBytes": 64 << 20})
				ev.Check("HTTP status classification preserved", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseRequest}), "err=%v", err)
				ev.Check("unknown deep code never interpreted as legitimate envelope", !strings.Contains(result.Message.ErrorMessage, "1210"), "nested code interpreted")
				ev.Check("bounded error parsing allocation", allocated < 64<<20, "allocated %d for %d-byte body", allocated, len(body))
				ev.Check("single attempt and cleanup", len(result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "attempts=%d", len(result.Metadata.Attempts))
				w.released(ev)
			})
		}
	}
}
