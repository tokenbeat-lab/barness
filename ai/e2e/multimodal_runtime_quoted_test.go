package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMultimodalRuntimeQuotedFields(t *testing.T) {
	for _, kind := range []string{"int", "number", "float", "bool", "string"} {
		for _, mode := range []string{"exact", "count", "expanded"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "UC09-quoted-"+kind+"-"+mode)
				req := jevBoxInput(t)
				raw, _ := base64.StdEncoding.DecodeString(req.Images[0].Data)
				u := newJevBoxWorld(t, ev, func(c *ai.Config) {
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 3, MaxExpandedImageBytes: 3 * int64(len(raw))}
					if mode == "count" {
						c.Policy.Classifier.Images.References.MaxReferences = 2
					}
					if mode == "expanded" {
						c.Policy.Classifier.Images.References.MaxExpandedImageBytes--
					}
				})
				enqueueJevBox(t, ev, u, jevBoxAnswer, "0")
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					key := "424242"
					var field any
					switch kind {
					case "int":
						field = struct {
							Image int `json:"image,string"`
						}{424242}
					case "number":
						field = struct {
							Image json.Number `json:"image,string"`
						}{"424242"}
					case "float":
						field = struct {
							Image float64 `json:"image,string"`
						}{424242}
					case "bool":
						key = "true"
						field = struct {
							Image bool `json:"image,string"`
						}{true}
					case "string":
						key = `"image-key"`
						field = struct {
							Image string `json:"image,string"`
						}{"image-key"}
					}
					var asset any
					for _, a := range p.Body["assets"].(map[string]any) {
						asset = a
					}
					p.Body["assets"] = map[string]any{key: asset}
					p.Body["state"] = map[string]any{"state": field, "images": []string{key}}
					p.Body["questions"].(map[string]any)["exists"].(map[string]any)["instructions"] = []string{"compare", key}
					return ai.KeepPayload(), nil
				}}
				res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-quoted-"+kind+mode), jevContextTarget(), req, jevContextOptions())
				if mode == "exact" {
					ev.Check("quoted field counts actual JSON string", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					ev.Check("quoted field obeys occurrence budgets", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0, "got %v", err)
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
