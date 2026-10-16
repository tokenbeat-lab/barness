package e2e

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevCriteriaScope(t *testing.T) {
	for _, sc := range []struct{ name, kind, raw, path string }{
		{"container", "score", `null`, "/questions/q/criteria"},
		{"question", "score", `null`, "/questions/q"},
		{"type", "score", `null`, "/questions/q/type"},
		{"absent", "choice", `null`, "/questions/q/criteria/missing"},
		{"bad-escape", "choice", `null`, "/questions/q/criteria/only~2"},
		{"negative", "score", `null`, "/questions/q/criteria/-1"},
		{"leading-zero", "score", `null`, "/questions/q/criteria/00"},
		{"plus", "score", `null`, "/questions/q/criteria/+0"},
		{"out-of-range", "score", `null`, "/questions/q/criteria/2"},
		{"append", "score", `null`, "/questions/q/criteria/-"},
		{"not-null", "score", `"text"`, "/questions/q/criteria/0"},
		{"object-root", "score", `{"x":null}`, "/questions/q/criteria/0"},
		{"duplicate", "score", `{"x":null}`, "/questions/q/criteria/0/x"},
		{"unused", "score", `{"x":null}`, "/questions/q/criteria/0/x"},
		{"unreferenced-score", "score", `null`, "/state/photo"},
		{"unreferenced-bool", "bool", `null`, "/state/photo"},
	} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC07-scope-"+sc.name+"-"+entry)
				u := newJevCriteriaWorld(t, ev, nil)
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.State = json.RawMessage(`{"photo":null}`)
				q, _, _ := jevCriteriaQuestion(sc.kind, json.RawMessage(sc.raw))
				req.Questions = map[string]ai.ClassifierQuestion{"q": q}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: sc.path}}
				if sc.name == "duplicate" {
					req.ImageReferences = append(req.ImageReferences, req.ImageReferences[0])
				}
				if sc.name == "unused" {
					req.Images = jevContextImages(t)
				}
				res, err := jevStateCall(t, u, entry, "req-criteria-scope-"+sc.name+entry, req, ai.Hooks{})
				checkUnaryFailure(ev, res, err, ai.CodeInvalidRequest, ai.PhaseScope, ai.StopReasonError)
				ev.Check("no credentials or send", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevCriteriaCapabilities(t *testing.T) {
	for _, kind := range []string{"choice", "score", "bool"} {
		for _, name := range []string{"root-format", "container-format", "location", "decisions", "binding", "experiment", "policy", "small", "mime"} {
			for _, entry := range []string{"plain", "hooked"} {
				t.Run(kind+"/"+name+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "UC07-capability-"+kind+"-"+name+"-"+entry)
					var u *unaryWorld
					if name == "decisions" {
						u = newReferenceWorld(t, ev, nil)
					} else {
						u = newJevCriteriaWorld(t, ev, func(c *ai.Config) {
							if name == "root-format" || name == "container-format" {
								c.Catalog.ClassifierModels[0].Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
							}
							if name == "location" {
								c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationInstructions}
							}
							if name == "policy" {
								c.Policy.Classifier.Images = nil
							}
						})
					}
					req := jevContextInput(t)
					req.Images = req.Images[:1]
					raw, tail := `null`, ""
					if name == "container-format" {
						raw, tail = `[null]`, "/0"
					}
					q, paths, answer := jevCriteriaQuestion(kind, json.RawMessage(raw))
					req.Questions = map[string]ai.ClassifierQuestion{"q": q}
					for _, path := range paths {
						req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/criteria" + path + tail})
					}
					target, opts := jevContextTarget(), ai.Options(jevContextOptions())
					code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
					switch name {
					case "decisions":
						target.ModelID, opts = "host-decisions", nil
					case "binding":
						code = ai.CodeTenantDenied
						u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
					case "experiment":
						opts = ai.TypeSafeOptions{}
					case "policy":
						phase = ai.PhaseScope
					case "small":
						req.Images[0] = jevPNG(t, 256, 256)
					case "mime":
						req.Images[0] = classifierAssets(t)[2]
					}
					if name == "root-format" {
						enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":` + answer + `}}`}.script(t, classifierProtocol, ""))
					}
					var res ai.ClassifierResult
					var err error
					if entry == "plain" {
						res, err = u.client.Classify(ctxFor(t), textScope("req-criteria-cap-"+kind+name+entry), target, req, opts)
					} else {
						res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-criteria-cap-"+kind+name+entry), target, req, opts)
					}
					if name == "root-format" {
						ev.Check("image root independent of text format", err == nil, "got %v", err)
						checkImageInputs(ev, res, req.Images)
					} else {
						checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
						ev.Check("capability before credentials", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
					}
					recordJevContext(ev, u, req, res, err)
				})
			}
		}
	}
}

func TestJevCriteriaBudgets(t *testing.T) {
	for _, name := range []string{"reuse", "references", "expanded", "source", "question", "body"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC07-budget-"+name+"-"+entry)
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.Questions = map[string]ai.ClassifierQuestion{"a": ai.BoolQuestion{Instructions: json.RawMessage(`"judge"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`[null,null]`), False: json.RawMessage(`null`)}}, "z": ai.ChoiceQuestion{Instructions: json.RawMessage(`"unrelated"`), Criteria: map[string]json.RawMessage{"only": json.RawMessage(`null`)}}}
				for _, path := range []string{"/false", "/true/1", "/true/0"} {
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/a/criteria" + path})
				}
				u := newJevCriteriaWorld(t, ev, func(c *ai.Config) {
					c.Policy.Classifier.Images.MaxImages = 1
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 3, MaxExpandedImageBytes: 1 << 19}
					switch name {
					case "references":
						c.Policy.Classifier.Images.References.MaxReferences = 2
					case "expanded":
						data, _ := base64.StdEncoding.DecodeString(req.Images[0].Data)
						c.Policy.Classifier.Images.References.MaxExpandedImageBytes = int64(len(data))*3 - 1
					case "source":
						c.Policy.Classifier.Images.MaxTotalImageBytes = 32
					case "question":
						c.Policy.Classifier.MaxQuestionBytes = 120
					case "body":
						c.Policy.MaxRequestBytes = 1650
						c.Policy.MaxImageBytes = 1650
						c.Policy.Classifier.Images.MaxTotalImageBytes = 1650
					}
				})
				if name == "reuse" {
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"a":{"type":"noul","noul":1},"z":{"type":"choice","choice":"only","probabilities":{"only":1},"confidence":1}}}`}.script(t, classifierProtocol, ""))
				}
				res, err := jevStateCall(t, u, entry, "req-criteria-budget-"+name+entry, req, ai.Hooks{})
				if name == "reuse" {
					ev.Check("source reused three times locally; Choice null unchanged", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseScope, ai.StopReasonError)
					ev.Check("budget before pin", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0, "got %v", err)
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
