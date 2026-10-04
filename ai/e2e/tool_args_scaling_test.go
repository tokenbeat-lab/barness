package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// toolScalingSizes are the argument sizes the scaling case streams: a ×4
// series reaching past the largest example limit, so a cost growing with the
// square of the size shows as a ×16 step and a linear one as ×4.
var toolScalingSizes = []int{128 << 10, 512 << 10, 2 << 20}

// toolScalingRuns is how often each size runs; the fastest run counts, which
// filters out a GC or scheduler pause landing on one of them.
const toolScalingRuns = 3

// maxToolScalingGrowth bounds the per-byte cost of the largest size against
// the smallest. Linear processing stays near 1; a quadratic one is 16 here.
const maxToolScalingGrowth = 3.0

// toolScalingPoint is one size's fastest single call.
type toolScalingPoint struct {
	Bytes        int     `json:"bytes"`
	Duration     string  `json:"duration"`
	NanosPerByte float64 `json:"nanosPerByte"`
}

// toolScalingReport is the scaling case's outcome, the numerical basis the
// examples' MaxToolJSONBytes comments cite.
type toolScalingReport struct {
	Points []toolScalingPoint `json:"points"`
	// Growth is the largest size's per-byte cost over the smallest's.
	Growth float64 `json:"growth"`
}

// TestToolArgumentsScaling streams one tool call's arguments in 64-byte
// deltas, as Responses streams them, at growing sizes (issue 32): the time a
// call takes must grow about linearly with its argument size, so that
// MaxToolJSONBytes is bounded by memory rather than CPU. The consumer reads
// every event without reading the view, as a host relaying deltas does;
// reading the view while arguments stream parses them at that moment.
//
// Opt-in: BARNESS_AI_PRESSURE=1.
func TestToolArgumentsScaling(t *testing.T) {
	ev := pressureCase(t, "E08-policy-pressure-tool-arguments-scaling")
	policy := *localassembly.LocalPolicy()
	largest := toolScalingSizes[len(toolScalingSizes)-1]
	// Room for the largest call: its frames, and the terminal frames that
	// repeat its arguments.
	policy.MaxToolJSONBytes = int64(largest)
	policy.MaxFrameBytes = int64(4 * largest)
	pw := newPressureWorld(t, policy, 1)
	defer checkPressureReleased(ev, pw)

	var rep toolScalingReport
	for _, size := range toolScalingSizes {
		frames := pressureToolCall(t, size)
		chunks := coalesce(frames, 64)
		best := time.Duration(0)
		for run := range toolScalingRuns {
			pw.provider.Enqueue(pressureReply(chunks))
			d, ok := timeToolCall(ev, pw, fmt.Sprintf("req-tool-scaling-%d-%d", size, run), size)
			if !ok {
				return
			}
			if best == 0 || d < best {
				best = d
			}
		}
		rep.Points = append(rep.Points, toolScalingPoint{Bytes: size,
			Duration: best.Round(time.Microsecond).String(), NanosPerByte: float64(best.Nanoseconds()) / float64(size)})
	}
	first, last := rep.Points[0], rep.Points[len(rep.Points)-1]
	rep.Growth = last.NanosPerByte / first.NanosPerByte
	ev.Record("tool-arguments-scaling", rep)
	t.Logf("tool arguments scaling: %+v", rep)
	ev.Check("a call's time grows about linearly with its argument size", rep.Growth <= maxToolScalingGrowth,
		"per-byte cost ×%.1f from %d to %d bytes", rep.Growth, first.Bytes, last.Bytes)
}

// timeToolCall runs one streamed tool call to its result and reports how
// long it took, checking the whole call arrived.
func timeToolCall(ev *evidence.Case, pw pressureWorld, requestID string, size int) (time.Duration, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), pressureDeadline)
	defer cancel()
	req := ai.Request{Messages: []ai.Message{ai.UserText("Write the summary to a file.")}, Tools: pressureRequest(designLoad{}).Tools}
	started := time.Now()
	s := pw.client.Stream(ctx, scopeFor(pw.tenants[0], requestID), ai.Target{BindingID: "primary", ModelID: pressureModel}, req, nil)
	defer s.Close()
	for s.Next() {
	}
	res, err := s.Result()
	elapsed := time.Since(started)
	var raw int
	for _, c := range res.Message.Content {
		if call, ok := c.(ai.ToolCall); ok {
			raw = len(call.RawArguments)
		}
	}
	return elapsed, ev.Check("the whole tool call arrived", err == nil && raw == size, "%s: err %v, %d of %d bytes", requestID, err, raw, size)
}
