package e2e

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesRawTextBudget(t *testing.T) {
	assets := classifierAssets(t)
	for _, name := range []string{"raw-input", "raw-value", "typed-value", "escaped-large", "escaped-boundary", "surrogate-boundary", "prefix-surrogate-boundary", "text-input"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD26-images-raw-text-"+name)
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				c.Policy.MaxRequestBytes = 8 << 20
				c.Policy.Classifier.MaxStateBytes = 128
				if name == "surrogate-boundary" {
					c.Policy.Classifier.MaxStateBytes = 64
				}
				if name == "prefix-surrogate-boundary" {
					c.Policy.Classifier.MaxStateBytes = 5
				}
			})
			text := json.RawMessage(`"` + strings.Repeat("x", 2<<20) + `"`)
			switch name {
			case "escaped-large":
				text = json.RawMessage(`"` + strings.Repeat(`\u0041`, 1000000) + `"`)
			case "escaped-boundary":
				text = json.RawMessage(`"` + strings.Repeat(`\u0041`, 128) + `"`)
			case "surrogate-boundary":
				text = json.RawMessage(`"` + strings.Repeat(`\ud83d\ude00`, 16) + `"`)
			case "prefix-surrogate-boundary":
				text = json.RawMessage(`"\u0041\ud83d\ude00"`)
			}
			messages := classifierImageMessages(assets[:1])
			content := messages[0].(map[string]any)["content"].([]any)
			content[0].(map[string]any)["text"] = text
			var input any = messages
			if name == "raw-input" || name == "escaped-large" || name == "escaped-boundary" || name == "surrogate-boundary" {
				input = json.RawMessage(mustMarshal(t, messages))
			}
			if name == "typed-value" {
				input = []struct {
					Role    string `json:"role"`
					Content []any  `json:"content"`
				}{{Role: "user", Content: content}}
			}
			if name == "text-input" {
				input = text
			}
			valid := strings.HasSuffix(name, "boundary")
			if valid {
				enqueueClassifierImages(t, ev, u)
			}
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			req := decisionsInput()
			if name == "prefix-surrogate-boundary" {
				req.State = json.RawMessage(`""`)
			}
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["input"] = input
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-rawtext-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			runtime.ReadMemStats(&after)
			if valid {
				ev.Check("raw escaped decoded byte boundary", err == nil, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseRequest, ai.StopReasonError)
				ev.Record("allocation-bytes", after.TotalAlloc-before.TotalAlloc)
				ev.Check("known raw text rejected before proportional copies", after.TotalAlloc-before.TotalAlloc < 1<<20, "allocated %d", after.TotalAlloc-before.TotalAlloc)
				ev.Check("raw text never attempts", len(res.Metadata.Attempts) == 0, "sent request")
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
