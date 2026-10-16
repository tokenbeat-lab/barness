package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestJevCriteriaLegend(t *testing.T) {
	for _, name := range []string{"root", "nested", "no-legend", "no-distribution", "missing", "extra", "wrong-image", "wrong-text", "duplicate", "distribution", "without-probabilities"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC07-legend-"+name+"-"+entry)
				u := newJevCriteriaWorld(t, ev, nil)
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
				req := jevContextInput(t)
				// Entry order differs from first final use. Both levels and repeated
				// occurrences must point into the final result, never entry indices.
				criteria := []json.RawMessage{json.RawMessage(`null`), json.RawMessage(`null`)}
				tail := ""
				if name == "nested" {
					criteria[0] = json.RawMessage(`{"a/b":{"~":[null,null]},"n":9007199254740993,"d":1e999999999999999999999}`)
					tail = "/a~1b/~0/0"
				}
				req.Questions["score/~"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"rate"`), Criteria: criteria}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/score~1~0/criteria/1"}, {ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/questions/score~1~0/criteria/0" + tail}}
				chunks := make([][]byte, 1)
				reply := provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json"}, Chunks: chunks, OnRequest: func(capture provider.Request) {
					var native struct {
						Questions map[string]struct{ Criteria []json.RawMessage }
					}
					mustUnmarshal(t, capture.Body, &native)
					levels := native.Questions["score/~"].Criteria
					legend := map[string]json.RawMessage{"0": levels[0], "1": levels[1]}
					answer := map[string]any{"type": "score", "score": 1, "confidence": 1, "probabilities": map[string]float64{"0": 0, "1": 1}, "legend": legend}
					switch name {
					case "missing":
						delete(legend, "1")
					case "extra":
						legend["2"] = json.RawMessage(`"extra"`)
					case "wrong-image":
						legend["0"] = levels[1]
					case "wrong-text":
						legend["0"] = json.RawMessage(`"different"`)
					case "distribution":
						answer["probabilities"] = map[string]float64{"0": 0.5, "1": 0.5}
					case "without-probabilities":
						delete(answer, "probabilities")
					case "no-legend":
						delete(answer, "legend")
					case "no-distribution":
						delete(answer, "legend")
						delete(answer, "probabilities")
					}
					body := mustMarshal(t, map[string]any{"model": "response-jev", "answers": map[string]any{"q": map[string]any{"type": "noul", "noul": 1}, "score/~": answer}, "usage": map[string]int{"input_tokens": 100, "output_tokens": 2}})
					if name == "duplicate" {
						body = []byte(strings.Replace(string(body), `"legend":{"0":`, `"legend":{"0":"duplicate","0":`, 1))
					}
					chunks[0] = body
					ev.Record("final-response", json.RawMessage(body))
				}}
				enqueue(ev, u.world, reply)
				res, err := jevStateCall(t, u, entry, "req-criteria-legend-"+name+entry, req, ai.Hooks{})
				success := name == "root" || name == "nested" || name == "no-legend" || name == "no-distribution"
				if success {
					ev.Check("final native rubric accepted", err == nil && len(res.Answers) == 2, "got %v", err)
					checkImageInputs(ev, res, []ai.Image{req.Images[1], req.Images[0]})
					if a, ok := res.Answers["score/~"].(ai.ScoreAnswer); ok {
						var public map[string]json.RawMessage
						mustUnmarshal(t, mustMarshal(t, a), &public)
						if name == "root" || name == "nested" {
							ev.Check("legend restores public null slots", jevCriteriaJSONEqual(a.Legend[0], criteria[0]) && jevCriteriaJSONEqual(a.Legend[1], criteria[1]), "wrong legend")
							expected := `[{"imageIndex":0,"placement":"value","pointer":"/0` + tail + `"},{"imageIndex":1,"placement":"value","pointer":"/1"}]`
							ev.Check("legend references address final ImageInputs", jsonEqual(public["legendImageReferences"], []byte(expected)), "wrong refs: %s", public["legendImageReferences"])
						} else {
							ev.Check("absent legend stays absent", len(a.Legend) == 0 && public["legendImageReferences"] == nil, "invented legend")
						}
					}
				} else {
					checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
					ev.Check("one legal usage retained", res.Usage.Input == 100 && res.ResponseModel == ai.Value("response-jev"), "lost usage/model")
				}
				ev.Check("2xx never replayed", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "replayed")
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevCriteriaLegendFinalHook(t *testing.T) {
	for _, name := range []string{"final", "stale", "changed-image", "duplicate-use", "state-first"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC07-legend-hook-"+name)
			u := newJevCriteriaWorld(t, ev, nil)
			req := jevContextInput(t)
			req.Images = req.Images[:1]
			req.Questions["old"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"rate"`), Criteria: []json.RawMessage{json.RawMessage(`null`), json.RawMessage(`"old"`)}}
			req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/old/criteria/0"}}
			images := jevContextImages(t)
			recordClassifierAssets(ev, images)
			levels := `["visual-B",{"a/b":["visual-A","visual-B",null],"n":9007199254740993},"top"]`
			legend := `{"0":"visual-B","1":{"a/b":["visual-A","visual-B",null],"n":9007199254740993},"2":"top"}`
			if name == "stale" {
				legend = `{"0":"visual-A","1":"old","2":"top"}`
			}
			if name == "changed-image" {
				legend = strings.Replace(legend, `"0":"visual-B"`, `"0":"visual-A"`, 1)
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"renamed":{"type":"score","score":1,"confidence":1,"probabilities":{"0":0,"1":1,"2":0},"legend":` + legend + `}},"usage":{"input_tokens":100,"output_tokens":2}}`}.script(t, classifierProtocol, ""))
			var borrowed map[string]any
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["questions"] = map[string]any{"renamed": map[string]any{"type": "score", "instructions": "rate final", "criteria": json.RawMessage(levels)}}
				p.Body["assets"] = map[string]any{"visual-A": map[string]any{"format": "png", "data_base64": images[0].Data}, "visual-B": map[string]any{"format": "jpeg", "data_base64": images[1].Data}}
				if name == "state-first" {
					p.Body["state"] = "visual-A"
				}
				if name == "duplicate-use" {
					p.Body["assets"].(map[string]any)["visual-B"] = p.Body["assets"].(map[string]any)["visual-A"]
				}
				borrowed = p.Body
				ev.Record("hook-final", p.Body)
				return ai.KeepPayload(), nil
			}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				borrowed["questions"], borrowed["assets"] = nil, nil
				return nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-legend-final-"+name), jevContextTarget(), req, jevContextOptions())
			if name == "final" || name == "duplicate-use" || name == "state-first" {
				ev.Check("final renamed rubric succeeds", err == nil && res.Answers["renamed"] != nil, "got %v", err)
				want := []ai.Image{images[1], images[0]}
				if name == "duplicate-use" {
					want = []ai.Image{images[0], images[0]}
				}
				if name == "state-first" {
					want = images
				}
				checkImageInputs(ev, res, want)
				if a, ok := res.Answers["renamed"].(ai.ScoreAnswer); ok {
					var public map[string]json.RawMessage
					mustUnmarshal(t, mustMarshal(t, a), &public)
					ev.Check("final legend nulls and exact numbers", jsonEqual(public["legend"], []byte(`{"0":null,"1":{"a/b":[null,null,null],"n":9007199254740993},"2":"top"}`)), "wrong legend")
					refs := `[{"imageIndex":0,"placement":"value","pointer":"/0"},{"imageIndex":1,"placement":"value","pointer":"/1/a~1b/0"},{"imageIndex":0,"placement":"value","pointer":"/1/a~1b/1"}]`
					if name == "state-first" {
						refs = `[{"imageIndex":1,"placement":"value","pointer":"/0"},{"imageIndex":0,"placement":"value","pointer":"/1/a~1b/0"},{"imageIndex":1,"placement":"value","pointer":"/1/a~1b/1"}]`
					}
					ev.Check("repeated final asset references", jsonEqual(public["legendImageReferences"], []byte(refs)), "wrong refs")
				}
			} else {
				ev.Check("stale/modified rubric rejected atomically", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && res.Answers == nil && res.ImageInputs == nil && res.Usage.Input == 100, "got %v", err)
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}
