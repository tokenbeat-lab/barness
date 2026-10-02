package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestDeepSeekResponsesByteLimits is P05/E08 (spec I9, ADR-0002,
// ADR-0007): the request body DeepSeek's capabilities shape, and the SSE
// frame and streamed output of its replies, at their boundary and at
// limit+1 on every entry point. Only these differ from OpenAI's Responses
// calls; the tool JSON and error body bounds are the shared adapter's,
// asserted on P01 by TestByteLimits.
func TestDeepSeekResponsesByteLimits(t *testing.T) {
	a := loadScenarioText(t, deepseekResponsesProtocol)
	chunks := a.chunks(t)
	largestFrame, streamBytes := largestAndTotal(chunks)
	script := func(ev *evidence.Case, lw limitWorld) {
		ev.Fixture("text.json", a.raw)
		enqueue(ev, lw.world, provider.SSE(chunks))
	}
	scripted := func(_ *testing.T, ev *evidence.Case, lw limitWorld, _ bool) { script(ev, lw) }
	within := func(ev *evidence.Case, _ limitWorld, o outcome) { checkScenarioTextSucceeded(ev, a, o) }
	keepsReceived := func(ev *evidence.Case, _ limitWorld, o outcome) {
		got := textOf(o.result.Message)
		ev.Check("message keeps the text received before the limit", strings.HasPrefix(a.text(), got), "got %q", got)
	}
	nothingSent := func(ev *evidence.Case, lw limitWorld, o outcome) {
		ev.Check("no inference request sent", len(lw.provider.Requests()) == 0, "got %d", len(lw.provider.Requests()))
		ev.Check("no attempt recorded", len(o.result.Metadata.Attempts) == 0, "got %+v", o.result.Metadata.Attempts)
	}
	boundaryOn := func(bc boundary) boundary {
		bc.entries, bc.casePrefix = scenarioEntries(ai.ResponsesOptions{}), "P05-E08"
		bc.target, bc.request, bc.script = a.target(), a.request(t), scripted
		return bc
	}

	t.Run("request-body", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "request-body", field: "MaxRequestBytes", phase: ai.PhaseRequest, set: setRequestBytes,
			size: func(t *testing.T, e outcomeEntry) int64 {
				w := newWorld(t, tenantA)
				w.provider.Enqueue(a.reply(t))
				if o := e.invoke(ctxFor(t), w, textScope("req-p05-measure-"+e.name), a.target(), a.request(t)); o.err != nil {
					t.Fatalf("measuring call failed: %v", o.err)
				}
				return int64(len(w.provider.Requests()[0].Body))
			},
			within: within, over: nothingSent,
		}))
	})
	t.Run("frame", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "frame", field: "MaxFrameBytes", phase: ai.PhaseStream, set: setFrameBytes, size: fixedSize(largestFrame),
			within: within, over: keepsReceived,
		}))
	})
	t.Run("output", func(t *testing.T) {
		runBoundary(t, boundaryOn(boundary{
			id: "output", field: "MaxOutputBytes", phase: ai.PhaseStream, set: setOutputBytes, size: fixedSize(streamBytes),
			within: within, over: keepsReceived,
		}))
	})
}
