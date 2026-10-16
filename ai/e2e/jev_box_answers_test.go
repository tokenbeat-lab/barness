package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevBoxAnswers(t *testing.T) {
	for _, confidence := range []string{"null", "0", "0.00001", "1"} {
		for _, existence := range []string{"0", "0.01", "1"} {
			for _, entry := range []string{"plain", "hooked"} {
				t.Run(confidence+"/"+existence+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "UC08-answer-"+confidence+"-"+existence+"-"+entry)
					u := newJevBoxWorld(t, ev, nil)
					req := jevBoxInput(t)
					req.Questions["choice"] = jevInstructionsQuestion("choice", json.RawMessage(`"choose"`))
					req.Questions["score"] = jevInstructionsQuestion("score", json.RawMessage(`"rate"`))
					box := strings.Replace(jevBoxAnswer, `"confidence":null`, `"confidence":`+confidence, 1)
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"location":` + box + `,"exists":{"type":"noul","noul":` + existence + `},"choice":` + jevInstructionsAnswer("choice") + `,"score":` + jevInstructionsAnswer("score") + `}}`}.script(t, classifierProtocol, ""))
					res, err := jevStateCall(t, u, entry, "req-box-answer-"+confidence+existence+entry, req, ai.Hooks{})
					ev.Check("all four question types evaluated together", err == nil && len(res.Answers) == 4, "got %v", err)
					a, ok := res.Answers["location"].(ai.BoxAnswer)
					ev.Check("sealed typed Box with unchanged edges", ok && a.Box == (ai.Box{Top: 0.25, Bottom: 0.82, Left: 0.44, Right: 0.84}), "got %+v", a)
					value, has := a.Confidence.Get()
					if confidence == "null" {
						ev.Check("explicit null is set without a numeric value", a.Confidence.IsNull() && !a.Confidence.IsZero() && !has, "lost null")
					} else {
						var want float64
						mustUnmarshal(t, []byte(confidence), &want)
						ev.Check("numeric confidence including zero retained", has && !a.Confidence.IsNull() && value == want, "got %v", value)
					}
					ev.Check("public answer has exactly specified fields", jsonEqual(mustMarshal(t, a), []byte(box)), "invented or changed fields")
					var want float64
					mustUnmarshal(t, []byte(existence), &want)
					exists, ok := res.Answers["exists"].(ai.BoolAnswer)
					ev.Check("existence remains independent and unthresholded", ok && exists.Probability == want, "got %+v", exists)
					ev.Check("one batch and one Attempt", len(u.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "split call")
					checkImageInputs(ev, res, req.Images)
					recordJevContext(ev, u, req, res, err)
				})
			}
		}
	}
}

func TestJevBoxInvalidAnswers(t *testing.T) {
	cases := map[string]string{}
	for _, field := range []string{"top", "bottom", "left", "right"} {
		for _, bad := range []string{"missing", "null", `"0.5"`, "false", "[]", "{}", "-0.1", "1.1", "1e9999", "NaN", "Infinity"} {
			edges := map[string]json.RawMessage{"top": []byte(`0.25`), "bottom": []byte(`0.82`), "left": []byte(`0.44`), "right": []byte(`0.84`)}
			if bad == "missing" {
				delete(edges, field)
			} else if bad != "NaN" && bad != "Infinity" {
				edges[field] = []byte(bad)
			}
			raw := string(mustMarshal(t, edges))
			if bad == "NaN" || bad == "Infinity" {
				raw = strings.Replace(raw, `"`+field+`":`+string(edges[field]), `"`+field+`":`+bad, 1)
			}
			cases[field+"-"+bad] = `{"type":"box","box":` + raw + `,"confidence":null}`
		}
	}
	for _, confidence := range []string{"missing", `"0"`, "false", "[]", "{}", "-0.1", "1.1", "1e9999"} {
		if confidence == "missing" {
			cases["confidence-"+confidence] = strings.Replace(jevBoxAnswer, `,"confidence":null`, "", 1)
		} else {
			cases["confidence-"+confidence] = strings.Replace(jevBoxAnswer, `"confidence":null`, `"confidence":`+confidence, 1)
		}
	}
	for name, box := range map[string]string{"zero-width": `{"top":0.25,"bottom":0.82,"left":0.44,"right":0.44}`, "zero-height": `{"top":0.25,"bottom":0.25,"left":0.44,"right":0.84}`, "inverted-x": `{"top":0.25,"bottom":0.82,"left":0.9,"right":0.84}`, "inverted-y": `{"top":0.9,"bottom":0.82,"left":0.44,"right":0.84}`, "box-null": `null`, "box-array": `[]`, "box-string": `"box"`, "case-edge": `{"Top":0.25,"bottom":0.82,"left":0.44,"right":0.84}`, "duplicate-edge": `{"top":0.25,"top":0.3,"bottom":0.82,"left":0.44,"right":0.84}`, "extra-edge": `{"top":0.25,"bottom":0.82,"left":0.44,"right":0.84,"image_id":"a"}`} {
		cases[name] = `{"type":"box","box":` + box + `,"confidence":null}`
	}
	cases["missing-box"] = `{"type":"box","confidence":null}`
	cases["wrong-type"] = strings.Replace(jevBoxAnswer, `"type":"box"`, `"type":"noul"`, 1)
	cases["refusal"] = `{"type":"refusal"}`
	cases["case-confidence"] = strings.Replace(jevBoxAnswer, `"confidence"`, `"Confidence"`, 1)
	cases["duplicate-confidence"] = strings.Replace(jevBoxAnswer, `"confidence":null`, `"confidence":null,"confidence":0`, 1)
	for _, field := range []string{"image_id", "not_found", "criteria", "probability"} {
		cases["extra-"+field] = strings.TrimSuffix(jevBoxAnswer, "}") + `,"` + field + `":null}`
	}
	for name, answer := range cases {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC08-invalid-answer-"+name+"-"+entry)
				u := newJevBoxWorld(t, ev, nil)
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				req := jevBoxInput(t)
				enqueueJevBox(t, ev, u, answer, "0")
				res, err := jevStateCall(t, u, entry, "req-box-bad-"+name+entry, req, ai.Hooks{})
				checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
				// Non-JSON number literals invalidate the whole envelope; no usage can
				// safely be recovered from those bytes. All valid envelopes retain usage.
				if !strings.Contains(name, "NaN") && !strings.Contains(name, "Infinity") {
					model, ok := res.ResponseModel.Get()
					ev.Check("legal usage and response identity survive bad Box", res.Usage.Input == 100 && res.Usage.Output == 20 && ok && model == "jev-1.14.0", "usage lost")
				}
				ev.Check("all answers and image metadata withheld; 2xx never replayed", res.Answers == nil && res.ImageInputs == nil && len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "partial or replay")
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
