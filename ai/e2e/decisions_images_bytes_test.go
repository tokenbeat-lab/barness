package e2e

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesByteEncoding(t *testing.T) {
	type part struct {
		Type string `json:"type"`
		Text []byte `json:"text,omitempty"`
		URL  string `json:"image_url,omitempty"`
	}
	type message struct {
		Role    string `json:"role"`
		Content []part `json:"content"`
	}
	asset := classifierAssets(t)[0]
	for _, name := range []string{"text-input-boundary", "typed-text-boundary", "text-input-limit", "typed-text-limit"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD19-images-byte-encoding-"+name)
			valid := name == "text-input-boundary" || name == "typed-text-boundary"
			text := []byte("foo")
			if !valid {
				text = make([]byte, 2<<20)
			}
			var input any = text
			expectedInput := `"Zm9v"`
			if name == "typed-text-boundary" || name == "typed-text-limit" {
				input = []message{{Role: "user", Content: []part{{Type: "input_text", Text: text}, {Type: "input_image", URL: "data:image/png;base64," + asset.Data}}}}
				expectedInput = `[{"role":"user","content":[{"type":"input_text","text":"Zm9v"},{"type":"input_image","image_url":"data:image/png;base64,` + asset.Data + `"}]}]`
			}
			wire := []byte(`{"model":"host-decisions","input":` + expectedInput + `,"questions":[{"type":"choice","name":"intent","instructions":"Classify","choices":[{"value":"other"},{"value":"track","description":"Tracking"}]}]}`)
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				c.Policy.MaxRequestBytes = 8 << 20
				c.Policy.Classifier.MaxStateBytes = 128
				if valid {
					c.Policy.MaxRequestBytes = int64(len(wire))
					c.Policy.MaxImageBytes = 128
					c.Policy.Classifier.Images.MaxTotalImageBytes = 128
				}
			})
			if valid {
				enqueueClassifierImages(t, ev, u)
			}
			req := decisionsInput()
			req.State = json.RawMessage(`""`)
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["input"] = input
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-byte-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			runtime.ReadMemStats(&after)
			if valid {
				ev.Fixture("expected-wire.json", wire)
				if ev.Check("byte slices obey real JSON boundary", err == nil && len(u.provider.Requests()) == 1, "got %v", err) {
					ev.Check("JSON base64 text is preserved", jsonEqual(u.provider.Requests()[0].Body, wire), "wrong body")
				}
			} else {
				checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseRequest, ai.StopReasonError)
				ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
				ev.Check("known base64 text limited before copying", after.TotalAlloc-before.TotalAlloc < 1<<20, "allocated %d", after.TotalAlloc-before.TotalAlloc)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
