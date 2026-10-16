package e2e

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevInstructionsBoundaries(t *testing.T) {
	for _, sc := range []struct{ name, raw, pointer string }{
		{"missing", `{}`, "/questions/q/instructions/photo"},
		{"non-null", `{"photo":"text"}`, "/questions/q/instructions/photo"},
		{"object", `{"photo":{}}`, "/questions/q/instructions/photo"},
		{"array", `{"photo":[]}`, "/questions/q/instructions/photo"},
		{"duplicate-json", `{"photo":null,"photo":null}`, "/questions/q/instructions/photo"},
		{"zero", `[null]`, "/questions/q/instructions/00"},
		{"negative", `[null]`, "/questions/q/instructions/-1"},
		{"append", `[null]`, "/questions/q/instructions/-"},
		{"oob", `[null]`, "/questions/q/instructions/1"},
		{"escape", `{"~2":null}`, "/questions/q/instructions/~2"},
		{"question-missing", `{"photo":null}`, "/questions/missing/instructions/photo"},
		{"type", `[null]`, "/questions/q/type"},
		{"question-container", `[null]`, "/questions/q"},
		{"criteria-container", `[null]`, "/questions/q/criteria"},
		{"authority", `[null]`, "/tenantId"},
		{"references", `[null]`, "/imageReferences/0"},
		{"duplicate", `[null]`, "/questions/q/instructions/0"},
		{"unused", `[null]`, "/questions/q/instructions/0"},
		{"unreferenced-null", `null`, "/state/photo"},
		{"empty-string", `""`, "/state/photo"},
		{"empty-object", `{}`, "/state/photo"},
		{"empty-array", `[]`, "/state/photo"},
	} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC06-scope-"+sc.name+"-"+entry)
				u := newJevInstructionsWorld(t, ev, nil)
				req := jevContextInput(t)
				req.State, req.Images = json.RawMessage(`{"photo":null}`), req.Images[:1]
				req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(sc.raw)}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: sc.pointer}}
				if sc.name == "duplicate" {
					req.ImageReferences = append(req.ImageReferences, req.ImageReferences[0])
				}
				if sc.name == "unused" {
					req.Images = jevContextImages(t)
				}
				res, err := jevStateCall(t, u, entry, "req-instructions-scope-"+sc.name+entry, req, ai.Hooks{})
				checkUnaryFailure(ev, res, err, ai.CodeInvalidRequest, ai.PhaseScope, ai.StopReasonError)
				ev.Check("scope rejected before credentials and Attempt", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "unexpected work")
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevInstructionsCapabilities(t *testing.T) {
	for _, name := range []string{"root-format", "container-format", "location", "criteria", "decisions", "binding", "policy", "experiment", "mime", "small"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC06-capability-"+name+"-"+entry)
				var u *unaryWorld
				if name == "decisions" {
					u = newReferenceWorld(t, ev, nil)
				} else {
					u = newJevInstructionsWorld(t, ev, func(c *ai.Config) {
						if name == "root-format" || name == "container-format" {
							c.Catalog.ClassifierModels[0].Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
						}
						if name == "location" {
							c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
						}
						if name == "policy" {
							c.Policy.Classifier.Images = nil
						}
					})
				}
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`null`)}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/instructions"}}
				code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
				target, opts := jevContextTarget(), ai.Options(jevContextOptions())
				switch name {
				case "container-format":
					req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`[null]`)}
					req.ImageReferences[0].Pointer += "/0"
				case "criteria":
					req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`"ask"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`{"photo":null}`), False: json.RawMessage(`"other"`)}}
					req.ImageReferences[0].Pointer = "/questions/q/criteria/true/photo"
				case "decisions":
					target.ModelID, opts = "host-decisions", nil
				case "binding":
					code = ai.CodeTenantDenied
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				case "policy":
					phase = ai.PhaseScope
				case "experiment":
					opts = ai.TypeSafeOptions{}
				case "mime":
					req.Images[0] = classifierAssets(t)[2]
				case "small":
					req.Images[0] = jevPNG(t, 256, 256)
				}
				if name == "root-format" {
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), textScope("req-instructions-cap-"+name+entry), target, req, opts)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-instructions-cap-"+name+entry), target, req, opts)
				}
				if name == "root-format" {
					ev.Check("root image independent of text formats", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
					ev.Check("capability rejection before credentials and Attempt", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "unexpected work")
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevInstructionsBudgets(t *testing.T) {
	for _, name := range []string{"reuse", "references", "expanded", "question", "source", "body"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC06-budget-"+name+"-"+entry)
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.Questions = map[string]ai.ClassifierQuestion{"a": ai.BoolQuestion{Instructions: json.RawMessage(`[null,null]`)}, "b": ai.BoolQuestion{Instructions: json.RawMessage(`null`)}}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/b/instructions"}, {ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/a/instructions/0"}, {ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/a/instructions/1"}}
				u := newJevInstructionsWorld(t, ev, func(c *ai.Config) {
					c.Policy.Classifier.Images.MaxImages = 1
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 3, MaxExpandedImageBytes: 1 << 19}
					switch name {
					case "references":
						c.Policy.Classifier.Images.References.MaxReferences = 2
					case "expanded":
						data, _ := base64.StdEncoding.DecodeString(req.Images[0].Data)
						c.Policy.Classifier.Images.References.MaxExpandedImageBytes = int64(len(data))*3 - 1
					case "question":
						c.Policy.Classifier.MaxQuestionBytes = 80
					case "source":
						c.Policy.Classifier.Images.MaxTotalImageBytes = 32
					case "body":
						c.Policy.MaxRequestBytes = 1650
						c.Policy.MaxImageBytes = 1650
						c.Policy.Classifier.Images.MaxTotalImageBytes = 1650
					}
				})
				if name == "reuse" {
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"a":{"type":"noul","noul":1},"b":{"type":"noul","noul":0}}}`}.script(t, classifierProtocol, ""))
				}
				res, err := jevStateCall(t, u, entry, "req-instructions-budget-"+name+entry, req, ai.Hooks{})
				if name == "reuse" {
					ev.Check("three occurrences upload once", err == nil && len(res.Answers) == 2, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseScope, ai.StopReasonError)
					ev.Check("known expansion fails before pin", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0, "got %v", err)
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
