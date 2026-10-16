package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevInstructionsFinalHooks(t *testing.T) {
	for _, name := range []string{"replace", "retype", "state-first", "remove", "from-text", "root-format", "substring", "object-key", "question-key", "criteria", "unused", "null", "empty", "url", "part", "references", "expanded", "question-bytes", "raw-question-bytes", "location", "container-format", "unselected", "model", "authority", "bad-answer", "stale-answer"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC06-hook-"+name)
			u := newJevInstructionsWorld(t, ev, func(c *ai.Config) {
				c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 1 << 19}
				switch name {
				case "references":
					c.Policy.Classifier.Images.References.MaxReferences = 2
				case "expanded":
					c.Policy.Classifier.Images.References.MaxExpandedImageBytes = 4000
				case "question-bytes", "raw-question-bytes":
					c.Policy.Classifier.MaxQuestionBytes = 256
				case "location":
					c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
				case "root-format", "container-format":
					c.Catalog.ClassifierModels[0].Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
				}
			})
			req := jevContextInput(t)
			req.Images = req.Images[:1]
			req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`null`)}
			req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/instructions"}}
			opts := jevContextOptions()
			if name == "from-text" || name == "location" || name == "unselected" {
				req.Images, req.ImageReferences = nil, nil
				req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`"ask"`)}
			}
			if name == "unselected" {
				opts.ExperimentalImages = false
			}
			images := jevContextImages(t)
			recordClassifierAssets(ev, images)
			want := []ai.Image{images[1], images[0]}
			success := name == "replace" || name == "retype" || name == "state-first" || name == "remove" || name == "from-text" || name == "root-format"
			answers := `"a":{"type":"noul","noul":0.5},"z":{"type":"noul","noul":1}`
			if name == "retype" {
				answers = `"a":` + jevInstructionsAnswer("score") + `,"z":` + jevInstructionsAnswer("choice")
			}
			if name == "bad-answer" {
				answers = `"a":{"type":"noul","noul":0.5},"z":{"type":"noul","noul":2}`
			}
			if name == "stale-answer" {
				answers = `"q":{"type":"noul","noul":0.5}`
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"model":"final-jev","answers":{` + answers + `},"usage":{"input_tokens":100,"output_tokens":2}}`}.script(t, classifierProtocol, ""))
			var borrowed map[string]any
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				assets := map[string]any{"visual-A": map[string]any{"type": "image", "format": "png", "data_base64": images[0].Data}, "visual-B": map[string]any{"type": "image", "format": "jpg", "data_base64": images[1].Data}}
				p.Body["assets"] = assets
				a := map[string]any{"type": "noul", "instructions": []any{"before", "visual-B", nil, "visual-A", "after"}}
				z := map[string]any{"type": "noul", "instructions": "visual-A"}
				questions := map[string]any{"z": z, "a": a}
				p.Body["questions"] = questions
				switch name {
				case "retype":
					a["type"], a["criteria"] = "score", []any{"low", "high"}
					z["type"], z["criteria"] = "choice", map[string]any{"yes/~": nil, "no": "other"}
				case "replace":
					assets["visual-B"] = assets["visual-A"]
					want = []ai.Image{images[0], images[0]}
				case "state-first":
					p.Body["state"] = "visual-A"
					want = images
				case "remove":
					delete(p.Body, "assets")
					a["instructions"], z["instructions"] = "plain a", "plain z"
					want = nil
				case "root-format":
					a["instructions"] = "visual-B"
				case "substring":
					a["instructions"] = "before visual-B after"
				case "object-key":
					a["instructions"] = map[string]any{"visual-B": nil}
				case "question-key":
					questions["visual-B"] = a
					delete(questions, "a")
				case "criteria":
					a["instructions"] = "plain"
					a["criteria"] = map[string]any{"true": "visual-B", "false": "plain"}
				case "unused":
					a["instructions"] = "visual-A"
				case "null":
					a["instructions"] = nil
				case "empty":
					a["instructions"] = []any{}
				case "url":
					assets["visual-B"] = map[string]any{"type": "image", "url": "https://example.invalid/private"}
				case "part":
					assets["visual-B"].(map[string]any)["part"] = []any{0, 0, 100, 100}
				case "references", "expanded":
					a["instructions"] = []any{"visual-B", "visual-B", "visual-A"}
				case "question-bytes":
					a["instructions"] = []any{"visual-B", strings.Repeat("x", 300)}
				case "raw-question-bytes":
					questions["a"] = json.RawMessage(`{"type":"noul","instructions":["visual-B","` + strings.Repeat("x", 300) + `"]}`)
				case "model":
					p.Body["model"] = "jev-latest"
				case "authority":
					p.Body["credential"] = "foreign"
				}
				ev.Record("hook-final", p.Body)
				borrowed = p.Body
				return ai.KeepPayload(), nil
			}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				borrowed["questions"], borrowed["assets"] = nil, nil
				return nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-instructions-hook-"+name), jevContextTarget(), req, opts)
			if success {
				ev.Check("final renamed questions determine answers", err == nil && len(res.Answers) == 2 && res.Answers["a"] != nil && res.Answers["z"] != nil, "got %v", err)
				checkImageInputs(ev, res, want)
			} else {
				code, phase := ai.CodeCallbackFailed, ai.PhaseRequest
				switch name {
				case "references", "expanded", "question-bytes", "raw-question-bytes":
					code = ai.CodeResourceLimit
				case "unselected", "model", "authority":
					code = ai.CodeTenantDenied
				case "bad-answer", "stale-answer":
					code, phase = ai.CodeProtocol, ai.PhaseResponse
				}
				ev.Check("final request rejected at boundary", errors.Is(err, &ai.Error{Code: code, Phase: phase}), "got %v", err)
				ev.Check("atomic failure", res.Answers == nil && res.ImageInputs == nil, "partial result")
				if phase == ai.PhaseRequest {
					ev.Check("no Attempt or send", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "unexpected send")
				} else {
					ev.Check("reported usage retained", res.Usage.Input == 100 && res.ResponseModel == ai.Value("final-jev") && len(res.Metadata.Attempts) == 1, "usage lost or replayed")
				}
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}
