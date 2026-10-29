package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIImagineFailureFacts(t *testing.T) {
	for _, name := range []string{"body-limit-after-usage", "image-limit-before-validation", "count-limit", "modality-tokens"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-failure-facts-"+name)
			w := newImagineWorld(t, func(c *ai.Config) {
				if name == "body-limit-after-usage" {
					c.Policy.MaxOutputBytes = 256
					c.Policy.MaxToolJSONBytes = 128
					c.Policy.MaxFrameBytes = 128
					c.Policy.Image.MaxOutputImageBytes = 64
					c.Policy.Image.MaxTotalOutputImageBytes = 128
				}
				if name == "image-limit-before-validation" {
					c.Policy.Image.MaxOutputImageBytes = 8
				}
			})
			body := `{"usage":{"cost_in_usd_ticks":400000000},"data":[{"b64_json":"` + imaginePNG + `"}]}`
			code := ai.CodeResourceLimit
			success := false
			switch name {
			case "body-limit-after-usage":
				body = `{"usage":{"cost_in_usd_ticks":400000000},"data":[{"b64_json":"` + strings.Repeat("A", 2000) + `"}]}`
			case "image-limit-before-validation":
				body = `{"data":[{"b64_json":"` + strings.Repeat("%", 100) + `"}]}`
			case "count-limit":
				body = `{"data":[` + strings.TrimSuffix(strings.Repeat(`{"b64_json":"`+imaginePNG+`"},`, 11), ",") + `]}`
			case "modality-tokens":
				success = true
				body = `{"data":[{"b64_json":"` + imaginePNG + `"}],"usage":{"input_tokens":10,"output_tokens":8,"total_tokens":18,"input_tokens_details":{"text_tokens":6,"image_tokens":4,"cached_tokens":2},"output_tokens_details":{"text_tokens":2,"image_tokens":3,"reasoning_tokens":3}}}`
			}
			enqueue(ev, w, imagineReply(body))
			res, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-facts-"+name), imagineTarget(), ai.ImagesRequest{Prompt: "Draw."}, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if success {
				ev.Check("valid token report", err == nil && res.Usage.Modalities != nil, "err=%v usage=%+v", err, res.Usage)
				if res.Usage.Modalities != nil {
					m := res.Usage.Modalities
					it, _ := m.InputText.Get()
					oi, _ := m.OutputImage.Get()
					ev.Check("modality facts", it == 6 && oi == 3, "modalities=%+v", m)
				}
			} else {
				var ae *ai.Error
				ev.Check("bounded failure", errors.As(err, &ae) && ae.Code == code, "err=%v", err)
			}
			if name == "body-limit-after-usage" {
				amount, ok := res.Metadata.Attempts[0].ReportedCostUSD.Get()
				ev.Check("parsed charge survives read limit", ok && amount == .04, "attempt=%+v", res.Metadata.Attempts[0])
			}
		})
	}
}
