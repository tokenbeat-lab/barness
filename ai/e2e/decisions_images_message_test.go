package e2e

import (
	"context"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesMessageType(t *testing.T) {
	for _, name := range []string{"omitted", "message", "other", "null", "number", "alias"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD25-images-message-type-"+name)
			u := newClassifierImagesWorld(t, nil)
			valid := name == "omitted" || name == "message"
			if valid {
				enqueueClassifierImages(t, ev, u)
			}
			images := classifierAssets(t)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				messages := classifierImageMessages(images)
				m := messages[0].(map[string]any)
				switch name {
				case "message":
					m["type"] = "message"
				case "other":
					m["type"] = "item_reference"
				case "null":
					m["type"] = nil
				case "number":
					m["type"] = 1
				case "alias":
					m["Type"] = "message"
				}
				p.Body["input"] = messages
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-message-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			if valid {
				ev.Check("supported message discriminator", err == nil && len(res.Answers) == 1, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, ai.CodeCallbackFailed, ai.PhaseRequest, ai.StopReasonError)
				ev.Check("malformed message never sends", len(res.Metadata.Attempts) == 0, "sent request")
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
