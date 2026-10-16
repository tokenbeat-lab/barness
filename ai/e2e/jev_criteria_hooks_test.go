package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevCriteriaFinalHooks(t *testing.T) {
	for _, name := range []string{"choice", "bool", "score", "from-text", "remove", "state-first", "root-format", "substring", "object-key", "candidate-key", "question-key", "type-key", "unused", "null", "url", "part", "references", "expanded", "source", "question-bytes", "raw-question-bytes", "location", "container-format", "binding", "unselected"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC07-hook-"+name)
			u := newJevCriteriaWorld(t, ev, func(c *ai.Config) {
				c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 1 << 19}
				switch name {
				case "references":
					c.Policy.Classifier.Images.References.MaxReferences = 2
				case "expanded":
					c.Policy.Classifier.Images.References.MaxExpandedImageBytes = 4000
				case "source":
					c.Policy.Classifier.Images.MaxImages = 1
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
			req.Questions = map[string]ai.ClassifierQuestion{"old": ai.BoolQuestion{Instructions: json.RawMessage(`"judge"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`null`), False: json.RawMessage(`"other"`)}}}
			req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/old/criteria/true"}}
			opts := jevContextOptions()
			if name == "from-text" || name == "location" || name == "binding" || name == "unselected" {
				req.Images, req.ImageReferences = nil, nil
				req.Questions = map[string]ai.ClassifierQuestion{"old": ai.BoolQuestion{Instructions: json.RawMessage(`"judge"`)}}
			}
			if name == "binding" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				opts.ExperimentalImages = false
			}
			if name == "unselected" {
				opts.ExperimentalImages = false
			}
			images := jevContextImages(t)
			recordClassifierAssets(ev, images)
			answer := `{"type":"noul","noul":0.75}`
			if name == "choice" {
				answer = `{"type":"choice","choice":"only","probabilities":{"only":1},"confidence":1}`
			}
			if name == "score" {
				answer = `{"type":"score","score":1,"confidence":1,"probabilities":{"0":0,"1":1}}`
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"new":` + answer + `},"usage":{"input_tokens":100,"output_tokens":2}}`}.script(t, classifierProtocol, ""))
			want := []ai.Image{images[1], images[0]}
			success := name == "choice" || name == "bool" || name == "score" || name == "from-text" || name == "remove" || name == "state-first" || name == "root-format"
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				assets := map[string]any{"visual-A": map[string]any{"format": "png", "data_base64": images[0].Data}, "visual-B": map[string]any{"format": "jpg", "data_base64": images[1].Data}}
				p.Body["assets"] = assets
				q := map[string]any{"type": "noul", "instructions": "judge final", "criteria": map[string]any{"false": []any{"visual-B", nil, "visual-A"}, "true": "visual-A"}}
				questions := map[string]any{"new": q}
				p.Body["questions"] = questions
				switch name {
				case "choice":
					q["type"], q["criteria"] = "choice", map[string]any{"only": []any{"visual-B", nil, "visual-A"}}
				case "score":
					q["type"], q["criteria"] = "score", []any{"visual-B", "visual-A"}
				case "remove":
					delete(p.Body, "assets")
					delete(q, "criteria")
					want = nil
				case "state-first":
					p.Body["state"] = "visual-A"
					want = images
				case "root-format":
					q["criteria"] = map[string]any{"false": "visual-B", "true": "visual-A"}
				case "substring":
					q["criteria"] = map[string]any{"true": "before visual-B after", "false": "visual-A"}
				case "object-key":
					q["criteria"] = map[string]any{"true": map[string]any{"visual-B": nil}, "false": "visual-A"}
				case "candidate-key":
					q["type"], q["criteria"] = "choice", map[string]any{"visual-B": "visual-A"}
				case "question-key":
					questions["visual-B"] = q
					delete(questions, "new")
				case "type-key":
					assets["noul"] = assets["visual-A"]
				case "unused":
					q["criteria"] = map[string]any{"true": "visual-A", "false": "other"}
				case "null":
					q["criteria"] = map[string]any{"true": nil, "false": "visual-A"}
				case "url":
					assets["visual-B"] = map[string]any{"type": "image", "url": "https://example.invalid/image"}
				case "part":
					assets["visual-B"].(map[string]any)["part"] = []int{0, 0, 100, 100}
				case "expanded":
					q["criteria"] = map[string]any{"false": []any{"visual-B", "visual-B", "visual-B", "visual-A"}, "true": "visual-A"}
				case "question-bytes":
					q["criteria"] = map[string]any{"false": []any{"visual-B", strings.Repeat("x", 300)}, "true": "visual-A"}
				case "raw-question-bytes":
					questions["new"] = json.RawMessage(`{"type":"noul","instructions":"judge","criteria":{"false":["visual-B","` + strings.Repeat("x", 300) + `"],"true":"visual-A"}}`)
				}
				ev.Record("hook-final", p.Body)
				return ai.KeepPayload(), nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-criteria-hook-"+name), jevContextTarget(), req, opts)
			if success {
				ev.Check("final local criteria succeed", err == nil && len(res.Answers) == 1 && res.Answers["new"] != nil, "got %v", err)
				checkImageInputs(ev, res, want)
			} else {
				code := ai.CodeCallbackFailed
				switch name {
				case "references", "expanded", "source", "question-bytes", "raw-question-bytes":
					code = ai.CodeResourceLimit
				case "binding", "unselected":
					code = ai.CodeTenantDenied
				}
				ev.Check("final guard rejects atomically before Attempt", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && res.Answers == nil && res.ImageInputs == nil && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}
