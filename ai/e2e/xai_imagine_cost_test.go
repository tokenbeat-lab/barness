package e2e

import (
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIImagineCostEstimates(t *testing.T) {
	cases := []struct {
		name, quality, size string
		refs, n             int
		bad                 bool
		status              ai.CostEstimateStatus
		low, high           float64
	}{
		{"low-1k", "low", "1k", 0, 1, false, ai.CostResolved, .04, .04}, {"low-1.5k", "low", "1.5k", 0, 1, false, ai.CostResolved, .05, .05}, {"low-2k", "low", "2k", 0, 1, false, ai.CostResolved, .06, .06},
		{"medium-1k", "medium", "1k", 0, 1, false, ai.CostResolved, .06, .06}, {"medium-1.5k", "medium", "1.5k", 0, 1, false, ai.CostResolved, .07, .07}, {"medium-2k", "medium", "2k", 0, 1, false, ai.CostResolved, .08, .08},
		{"three-outputs", "low", "1k", 0, 3, false, ai.CostResolved, .12, .12}, {"two-inputs", "medium", "2k", 2, 1, false, ai.CostResolved, .10, .10}, {"auto", "auto", "1k", 0, 1, false, ai.CostBounded, .04, .06}, {"omitted-quality", "", "1k", 0, 1, false, ai.CostPartial, 0, 0}, {"omitted-size", "low", "", 0, 1, false, ai.CostPartial, 0, 0}, {"invalid-output", "low", "1k", 0, 1, true, ai.CostPartial, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-cost-"+tc.name)
			w := newImagineWorld(t)
			req := ai.ImagesRequest{Prompt: "Draw."}
			for range tc.refs {
				req.ReferenceImages = append(req.ReferenceImages, ai.Image{Data: imaginePNG, MimeType: "image/png"})
			}
			opts := ai.XAIImagesOptions{N: ai.Value(tc.n)}
			if tc.quality != "" {
				opts.Quality = ai.Value(tc.quality)
			}
			if tc.size != "" {
				opts.Resolution = ai.Value(tc.size)
			}
			data := make([]any, tc.n)
			for i := range data {
				data[i] = map[string]any{"b64_json": imaginePNG}
			}
			if tc.bad {
				data[0] = map[string]any{"b64_json": "%%%"}
			}
			enqueue(ev, w, imagineReply(string(mustMarshal(t, map[string]any{"data": data}))))
			res, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-cost-"+tc.name), imagineTarget(), req, opts)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("expected outcome", (err != nil) == tc.bad, "err=%v", err)
			if len(res.Metadata.Attempts) == 0 {
				ev.Check("attempt exists", false, "metadata=%+v", res.Metadata)
				return
			}
			estimate := res.Metadata.Attempts[0].CostEstimate
			ev.Check("estimate status", estimate.Status == tc.status, "got %+v", estimate)
			minimum, hasMinimum := estimate.Minimum.Get()
			maximum, hasMaximum := estimate.Maximum.Get()
			amounts := tc.status == ai.CostResolved || tc.status == ai.CostBounded
			ev.Check("estimate presence", hasMinimum == amounts && hasMaximum == amounts, "got %+v", estimate)
			if amounts {
				ev.Check("independent spec rates", math.Abs(minimum.Total-tc.low) < 1e-12 && math.Abs(maximum.Total-tc.high) < 1e-12, "got %+v", estimate)
			}
			ev.Check("money does not fabricate tokens", res.Usage.Input == 0 && res.Usage.Output == 0 && res.Usage.Cost.Total == 0 && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "usage=%+v", res.Usage)
		})
	}
}
