package e2e

import (
	"context"
	"encoding/base64"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesTypedRequestBudget(t *testing.T) {
	type part struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
		URL  string `json:"image_url,omitempty"`
	}
	type message struct {
		Role    string `json:"role"`
		Content []part `json:"content"`
	}
	asset := classifierAssets(t)[0]
	for _, name := range []string{"text", "image"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD26-images-typed-request-"+name)
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				c.Policy.MaxRequestBytes = 256 << 10
				c.Policy.MaxImageBytes = 256 << 10
				c.Policy.Classifier.Images.MaxTotalImageBytes = 256 << 10
				c.Policy.Classifier.MaxStateBytes = 8 << 20
			})
			text := "ok"
			url := "data:image/png;base64," + asset.Data
			if name == "text" {
				text = strings.Repeat("x", 2<<20)
			} else {
				raw := make([]byte, 230000)
				copy(raw, []byte("\x89PNG\r\n\x1a\n"))
				url = "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
			}
			input := []message{{Role: "user", Content: []part{{Type: "input_text", Text: text}, {Type: "input_image", URL: url}}}}
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["input"] = input
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-typed-budget-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			runtime.ReadMemStats(&after)
			checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseRequest, ai.StopReasonError)
			ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
			ev.Check("known typed input respects whole JSON before copies", after.TotalAlloc-before.TotalAlloc < 256<<10, "allocated %d", after.TotalAlloc-before.TotalAlloc)
			ev.Check("no transmitted attempt", len(res.Metadata.Attempts) == 0, "sent request")
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
