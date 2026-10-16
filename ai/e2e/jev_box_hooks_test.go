package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevBoxFinalHooks(t *testing.T) {
	for _, name := range []string{"to-box", "to-bool", "to-choice", "to-score", "remove-box-images", "remove-images", "other-instructions", "own-image", "replaced", "criteria", "null-criteria", "empty-instructions", "unsupported-kind", "question-budget", "references", "expanded", "header-removed", "header-duplicate", "long-context", "model", "freeze"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC08-hook-"+name)
			u := newJevBoxWorld(t, ev, func(c *ai.Config) {
				switch name {
				case "unsupported-kind":
					c.Catalog.ClassifierModels[0].Capabilities.Kinds = c.Catalog.ClassifierModels[0].Capabilities.Kinds[:3]
				case "question-budget":
					c.Policy.Classifier.MaxQuestionBytes = 128
				case "references":
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 1, MaxExpandedImageBytes: 1 << 19}
				case "expanded":
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 2000}
				}
			})
			req := jevBoxInput(t)
			if name == "to-box" || name == "unsupported-kind" {
				req.Questions["location"] = ai.BoolQuestion{Instructions: []byte(`"present?"`)}
			}
			answer := jevBoxAnswer
			switch name {
			case "to-bool", "remove-box-images":
				answer = `{"type":"noul","noul":0}`
			case "to-choice":
				answer = `{"type":"choice","choice":"only","probabilities":{"only":1},"confidence":1}`
			case "to-score":
				answer = `{"type":"score","score":1,"confidence":1,"probabilities":{"0":0,"1":1},"legend":{"0":"low","1":"high"}}`
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"final":` + answer + `},"usage":{"input_tokens":100,"output_tokens":20}}`}.script(t, classifierProtocol, ""))
			success := name == "to-box" || name == "to-bool" || name == "to-choice" || name == "to-score" || name == "remove-box-images" || name == "own-image" || name == "replaced" || name == "freeze"
			images := jevContextImages(t)
			recordClassifierAssets(ev, images)
			var borrowed map[string]any
			hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
				switch name {
				case "header-removed":
					h.Del("TypeSafe-Danger-Zone-Images")
				case "header-duplicate":
					h["typesafe-danger-zone-images"] = []string{"true"}
				case "long-context":
					h.Set("TypeSafe-Danger-Zone-Long-Context", "true")
				}
				return nil
			}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				q := map[string]any{"type": "box", "instructions": "locate final target"}
				p.Body["questions"] = map[string]any{"final": q}
				p.Body["assets"] = map[string]any{"visual-A": map[string]any{"type": "image", "format": "png", "data_base64": images[0].Data}}
				state := map[string]any{"state": "kept", "images": []any{"visual-A"}}
				p.Body["state"] = state
				switch name {
				case "to-bool", "remove-box-images":
					q["type"] = "noul"
				case "to-choice":
					q["type"] = "choice"
					q["criteria"] = map[string]any{"only": nil}
				case "to-score":
					q["type"] = "score"
					q["criteria"] = []any{"low", "high"}
				case "criteria":
					q["criteria"] = []any{}
				case "null-criteria":
					q["criteria"] = nil
				case "empty-instructions":
					q["instructions"] = ""
				case "question-budget":
					q["instructions"] = strings.Repeat("x", 200)
				case "references", "expanded":
					state["images"] = []any{"visual-A", "visual-A", "visual-A"}
				case "other-instructions":
					state["images"] = []any{}
					p.Body["questions"].(map[string]any)["other"] = map[string]any{"type": "noul", "instructions": "visual-A"}
				case "own-image":
					state["images"] = []any{}
					q["instructions"] = []any{"find", "visual-A"}
				case "replaced":
					p.Body["assets"] = map[string]any{"visual-B": map[string]any{"format": "jpg", "data_base64": images[1].Data}}
					state["images"] = []any{}
					q["instructions"] = "visual-B"
				case "model":
					p.Body["model"] = "jev-preview"
				}
				if name == "remove-images" || name == "remove-box-images" {
					delete(p.Body, "assets")
					p.Body["state"] = "kept"
				}
				borrowed = q
				ev.Record("hook-final", p.Body)
				return ai.KeepPayload(), nil
			}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				if name == "freeze" {
					borrowed["type"] = "score"
					borrowed["instructions"] = nil
				}
				return nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-box-hook-"+name), jevContextTarget(), req, jevContextOptions())
			if success {
				ev.Check("answers follow final authorized questions", err == nil && len(res.Answers) == 1 && res.Answers["final"] != nil, "got %v", err)
				want := images[:1]
				if name == "remove-box-images" {
					want = nil
				}
				if name == "replaced" {
					want = images[1:]
				}
				checkImageInputs(ev, res, want)
			} else {
				code := ai.CodeCallbackFailed
				switch name {
				case "question-budget", "references", "expanded":
					code = ai.CodeResourceLimit
				case "header-removed", "header-duplicate", "long-context", "model":
					code = ai.CodeTenantDenied
				}
				checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
				ev.Check("final guard prevents Attempt and HTTP send", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}

func TestJevBoxSnapshots(t *testing.T) {
	ev := run.Case(t, "UC08-snapshot-input")
	var source *ai.ClassifierModel
	u := newJevBoxWorld(t, ev, func(c *ai.Config) { source = &c.Catalog.ClassifierModels[0] })
	req := jevBoxInput(t)
	original := string(mustMarshal(t, req))
	hooks := ai.Hooks{TransformHeaders: func(context.Context, ai.CallScope, http.Header) error {
		q := req.Questions["location"].(ai.BoxQuestion)
		q.Instructions[0] = 'x'
		delete(req.Questions, "exists")
		req.Images[0].Data = "invalid"
		source.Capabilities.Kinds[3] = "unknown"
		return nil
	}}
	enqueueJevBox(t, ev, u, jevBoxAnswer, "0")
	res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-box-snapshot"), jevContextTarget(), req, jevContextOptions())
	ev.Check("input and catalog copies own Box instructions", err == nil && len(res.Answers) == 2, "got %v", err)
	var input ai.ClassifierRequest
	mustUnmarshal(t, []byte(original), &input)
	recordJevContext(ev, u, input, res, err)
	a, ok := res.Answers["location"].(ai.BoxAnswer)
	ev.Check("Box answer has value ownership", ok, "missing box")
	a.Box.Left = 0
	a.Confidence = ai.Value(1.0)
	saved := res.Answers["location"].(ai.BoxAnswer)
	ev.Check("borrowed typed answer copy is independent", saved.Box.Left == 0.44 && saved.Confidence.IsNull(), "aliased")
}
