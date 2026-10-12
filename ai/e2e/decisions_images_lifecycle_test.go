package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesSnapshot(t *testing.T) {
	for _, entry := range []string{"initial", "hook"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "P10-OD22-images-snapshot-"+entry)
			u := newClassifierImagesWorld(t, nil)
			assets := classifierAssets(t)
			req := decisionsInput()
			original := assets[0]
			if entry == "initial" {
				req.Images = assets
			}
			var borrowed map[string]any
			enqueueClassifierImages(t, ev, u)
			res, err := u.client.WithHooks(ai.Hooks{
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					// OnPayload is after entry copying. Mutations are synchronous.
					if entry == "hook" {
						p.Body["input"] = classifierImageMessages(assets)
					}
					borrowed = p.Body
					assets[0] = ai.Image{Data: "changed", MimeType: "image/gif"}
					req.Questions["intent"] = nil
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false; b.Version = "b2" })
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					borrowed["input"].([]any)[0].(map[string]any)["content"].([]any)[1].(map[string]any)["image_url"] = "changed"
					return nil
				},
			}).Classify(ctxFor(t), textScope("req-image-snapshot-"+entry), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			ev.Check("pinned permission and copied images", err == nil && res.Metadata.BindingVersion == "b1" && len(res.Answers) == 1, "got %v", err)
			if ev.Check("one image request", len(u.provider.Requests()) == 1, "unexpected request count") {
				var body map[string]any
				mustUnmarshal(t, u.provider.Requests()[0].Body, &body)
				url := body["input"].([]any)[0].(map[string]any)["content"].([]any)[1].(map[string]any)["image_url"]
				ev.Check("captured bytes independent of references", url == "data:"+original.MimeType+";base64,"+original.Data, "frozen image changed")
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsImagesCancellation(t *testing.T) {
	// Fixed valid header plus a large synthetic tail: encoding/header validation
	// is deliberately streaming and does not claim complete raster validity.
	raw := make([]byte, 32<<20)
	copy(raw, []byte("\x89PNG\r\n\x1a\n"))
	data := base64.StdEncoding.EncodeToString(raw)
	u := newClassifierImagesWorld(t, func(c *ai.Config) {
		c.Policy.MaxRequestBytes = 64 << 20
		c.Policy.MaxImageBytes = 40 << 20
		c.Policy.Classifier.Images.MaxTotalImageBytes = 40 << 20
	})
	ev := run.Case(t, "P10-OD42-images-validation-cancellation")
	req := decisionsInput()
	req.Images = []ai.Image{{Data: data, MimeType: "image/png"}}
	ctx, cancel := context.WithTimeout(ctxFor(t), time.Millisecond)
	defer cancel()
	started := time.Now()
	res, err := u.client.Classify(ctx, textScope("req-image-cancellation"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	ev.Record("elapsed-ms", time.Since(started).Milliseconds())
	checkUnaryFailure(ev, res, err, ai.CodeDeadlineExceeded, ai.PhaseScope, ai.StopReasonError)
	ev.Check("validation ends before resolving", u.host.ReadsFor(res.Metadata.RequestID).Bindings == 0 && len(u.provider.Requests()) == 0, "unexpected network")
	// Recovery is a new public call after the failed validation.
	enqueueClassifierImages(t, ev, u)
	req.Images = classifierAssets(t)
	req.State = json.RawMessage(`{}`)
	res, err = u.client.Classify(ctxFor(t), textScope("req-image-recovery"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	ev.Check("image-only judgment with required empty State", err == nil && len(res.Answers) == 1, "got %v", err)
	u.records(ev, res, err)
	u.released(ev)
}
