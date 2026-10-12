package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsNativeLimits(t *testing.T) {
	for _, name := range []string{"initial-name", "initial-body", "hook-name", "hook-question", "hook-count", "hook-input", "hook-body", "custom-question"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD25-limits-"+name+"-hooked")
			u := newDecisionsWorld(t, func(c *ai.Config) {
				c.Policy.Classifier.MaxQuestionBytes = 220
				c.Policy.Classifier.MaxStateBytes = 256
				c.Policy.Classifier.MaxQuestions = 2
				if name == "initial-body" || name == "hook-body" {
					c.Policy.MaxRequestBytes = 256
					if name == "initial-body" {
						c.Policy.MaxRequestBytes = 180
					}
					c.Policy.MaxImageBytes = 128
				}
			})
			req := decisionsInput()
			if name == "initial-name" {
				req.Questions[strings.Repeat("x", 240)] = req.Questions["intent"]
				delete(req.Questions, "intent")
			}
			payloads := 0
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				payloads++
				q := p.Body["questions"].([]any)[0].(map[string]any)
				switch name {
				case "initial-name", "initial-body":
					q["name"] = "q"
					q["instructions"] = "x"
					p.Body["input"] = ""
				case "hook-name":
					q["name"] = strings.Repeat("x", 240)
				case "hook-question":
					q["instructions"] = strings.Repeat("x", 240)
				case "hook-count":
					p.Body["questions"] = []any{q, q, q}
				case "hook-input":
					p.Body["input"] = strings.Repeat("x", 257)
				case "hook-body":
					p.Body["input"] = strings.Repeat("x", 200)
				case "custom-question":
					p.Body["questions"] = decisionsRaw(`[{"type":"choice","name":"` + strings.Repeat("x", 240) + `","instructions":"x","choices":[{"value":"a"},{"value":"b"}]}]`)
				}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-limit-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseRequest, ai.StopReasonError)
			if strings.HasPrefix(name, "initial") {
				ev.Check("initial native limit checked before callback", payloads == 0, "got %d", payloads)
			} else {
				ev.Check("final limit actually exercised after callback", payloads == 1, "got %d", payloads)
			}
			ev.Check("limits never send", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
