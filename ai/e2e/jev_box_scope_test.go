package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevBoxVisibility(t *testing.T) {
	for _, name := range []string{"context", "state", "own-root", "own-nested", "each-own", "no-image", "other-instructions", "other-criteria", "second-invisible", "null-instructions", "empty-instructions", "primitive-instructions", "question-budget"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC08-visibility-"+name+"-"+entry)
				u := newJevBoxWorld(t, ev, func(c *ai.Config) {
					if name == "question-budget" {
						c.Policy.Classifier.MaxQuestionBytes = 128
					}
				})
				req := jevBoxInput(t)
				success := name == "context" || name == "state" || name == "own-root" || name == "own-nested" || name == "each-own"
				switch name {
				case "state":
					req.State = []byte(`{"photo":null}`)
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}}
				case "own-root", "own-nested", "each-own", "second-invisible":
					raw, tail := `null`, ""
					if name == "own-nested" {
						raw, tail = `["find",{"a/b":[null,"after"]}]`, "/1/a~1b/0"
					}
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(raw)}
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/location/instructions" + tail}}
					if name == "each-own" || name == "second-invisible" {
						req.Questions["second/~"] = ai.BoxQuestion{Instructions: []byte(`"other target"`)}
						if name == "each-own" {
							req.Questions["second/~"] = ai.BoxQuestion{Instructions: []byte(`{"~":null}`)}
							req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/second~1~0/instructions/~0"})
						}
					}
				case "no-image":
					req.Images = nil
				case "other-instructions":
					req.Questions["exists"] = ai.BoolQuestion{Instructions: []byte(`null`)}
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/exists/instructions"}}
				case "other-criteria":
					req.Questions["exists"] = ai.BoolQuestion{Instructions: []byte(`"present?"`), Criteria: &ai.BoolCriteria{True: []byte(`null`), False: []byte(`"no"`)}}
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/exists/criteria/true"}}
				case "null-instructions":
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(`null`)}
				case "empty-instructions":
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(`""`)}
				case "primitive-instructions":
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(`42`)}
				case "question-budget":
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(`"` + strings.Repeat("x", 130) + `"`)}
				}
				if success {
					extra := ""
					if name == "each-own" {
						extra = `,"second/~":` + jevBoxAnswer
					}
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"location":` + jevBoxAnswer + `,"exists":{"type":"noul","noul":0}` + extra + `}}`}.script(t, classifierProtocol, ""))
				}
				original := string(mustMarshal(t, req))
				res, err := jevStateCall(t, u, entry, "req-box-visible-"+name+entry, req, ai.Hooks{})
				ev.Check("borrowed request unchanged", original == string(mustMarshal(t, req)), "mutated")
				if success {
					ev.Check("every Box sees a valid image in its scope", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					code := ai.CodeInvalidRequest
					if name == "question-budget" {
						code = ai.CodeResourceLimit
					}
					checkUnaryFailure(ev, res, err, code, ai.PhaseScope, ai.StopReasonError)
					ev.Check("scope denial before binding and credential reads", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "sent")
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevBoxCapabilities(t *testing.T) {
	for _, name := range []string{"kind", "jev-1.13", "decisions", "policy", "binding", "experiment", "location", "format", "root-format"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC08-capability-"+name+"-"+entry)
				var u *unaryWorld
				switch name {
				case "jev-1.13":
					u = newUnaryWorld(t, func(c *ai.Config) {
						c.Policy.Classifier.Images = &ai.ClassifierImagePolicy{MaxImages: 8, MaxTotalImageBytes: 1 << 19}
					})
				case "decisions":
					u = newReferenceWorld(t, ev, nil)
				default:
					u = newJevBoxWorld(t, ev, func(c *ai.Config) {
						m := &c.Catalog.ClassifierModels[0]
						switch name {
						case "kind":
							m.Capabilities.Kinds = m.Capabilities.Kinds[:3]
						case "policy":
							c.Policy.Classifier.Images = nil
						case "location":
							m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
						case "format", "root-format":
							m.Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
						}
					})
				}
				req := jevBoxInput(t)
				target, opts := jevContextTarget(), ai.Options(jevContextOptions())
				code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
				switch name {
				case "jev-1.13":
					target.ModelID = "host-jev"
					opts = nil

				case "decisions":
					target.ModelID = "host-decisions"
					opts = nil
				case "binding":
					code = ai.CodeTenantDenied
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				case "experiment":
					opts = ai.TypeSafeOptions{}
				case "policy":
					phase = ai.PhaseScope
				case "format":
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(`["find",null]`)}
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/location/instructions/1"}}
				case "root-format":
					req.Questions["location"] = ai.BoxQuestion{Instructions: []byte(`null`)}
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/location/instructions"}}
					enqueueJevBox(t, ev, u, jevBoxAnswer, "0")
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), textScope("req-box-cap-"+name+entry), target, req, opts)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-box-cap-"+name+entry), target, req, opts)
				}
				if name == "root-format" {
					ev.Check("root image bypasses only text format", err == nil, "got %v", err)
				} else {
					checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
					ev.Check("unsupported requests never pin or send", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
