package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevStateBoundaries(t *testing.T) {
	for _, sc := range []struct{ name, state, pointer string }{
		{"missing", `{}`, "/state/photo"}, {"non-null", `{"photo":"text"}`, "/state/photo"},
		{"object", `{"photo":{}}`, "/state/photo"}, {"array-container", `{"photo":[]}`, "/state/photo"},
		{"array-leading-zero", `[null]`, "/state/00"}, {"array-plus", `[null]`, "/state/+0"},
		{"array-negative-zero", `[null]`, "/state/-0"}, {"array-negative", `[null]`, "/state/-1"},
		{"array-append", `[null]`, "/state/-"}, {"array-oob", `[null]`, "/state/1"},
		{"array-overflow", `[null]`, "/state/9999999999999999999999999"},
		{"bad-escape", `{"~2":null}`, "/state/~2"}, {"dangling-escape", `{"~":null}`, "/state/~"},
		{"fragment", `null`, "#/state"}, {"request-root", `{}`, "/"},
		{"questions", `{}`, "/questions"}, {"question", `{}`, "/questions/q"},
		{"type", `{}`, "/questions/q/type"}, {"source", `{}`, "/images/0"},
		{"duplicate", `{"photo":null}`, "/state/photo"},
		{"unused", `{"photo":null}`, "/state/photo"},
		{"dangling-source", `{"photo":null}`, "/state/photo"},
		{"root-unreferenced", `null`, "/questions/q/instructions/1"},
	} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC05-scope-"+sc.name+"-"+entry)
				u := newJevStateWorld(t, ev, nil)
				req := jevContextInput(t)
				req.State, req.Images = json.RawMessage(sc.state), req.Images[:1]
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: sc.pointer}}
				switch sc.name {
				case "duplicate":
					req.ImageReferences = append(req.ImageReferences, req.ImageReferences[0])
				case "unused":
					req.Images = jevContextImages(t)
				case "dangling-source":
					req.ImageReferences[0].ImageIndex = 1
				case "root-unreferenced":
					req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`["ask",null]`)}
				}
				res, err := jevStateCall(t, u, entry, "req-state-scope-"+sc.name+entry, req, ai.Hooks{})
				checkUnaryFailure(ev, res, err, ai.CodeInvalidRequest, ai.PhaseScope, ai.StopReasonError)
				ev.Check("scope failure before credentials or Attempt", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "unexpected work")
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevStateCapabilities(t *testing.T) {
	for _, name := range []string{"context-only", "instructions", "criteria", "decisions", "state-only-context", "binding", "experiment"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC05-capability-"+name+"-"+entry)
				var u *unaryWorld
				if name == "decisions" {
					u = newReferenceWorld(t, ev, nil)
				} else {
					u = newJevStateWorld(t, ev, func(c *ai.Config) {
						if name == "context-only" {
							c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}
						}
						if name == "state-only-context" {
							c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
						}
					})
				}
				req := jevContextInput(t)
				req.State, req.Images = json.RawMessage(`{"photo":null}`), req.Images[:1]
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}}
				if name == "instructions" || name == "criteria" {
					req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`["ask",null]`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`{"photo":null}`), False: json.RawMessage(`"other"`)}}
					if name == "instructions" {
						req.ImageReferences[0].Pointer = "/questions/q/instructions/1"
					} else {
						req.ImageReferences[0].Pointer = "/questions/q/criteria/true/photo"
					}
				}
				if name == "state-only-context" || name == "decisions" {
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
				}
				target, opts := jevContextTarget(), ai.Options(jevContextOptions())
				if name == "decisions" {
					target.ModelID, opts = "host-decisions", nil
				}
				code := ai.CodeInvalidRequest
				if name == "binding" {
					code = ai.CodeTenantDenied
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				}
				if name == "experiment" {
					opts = ai.TypeSafeOptions{}
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), textScope("req-state-cap-"+name+entry), target, req, opts)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-state-cap-"+name+entry), target, req, opts)
				}
				checkUnaryFailure(ev, res, err, code, ai.PhaseCapability, ai.StopReasonError)
				ev.Check("unsupported request rejected as a whole before pin", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "unexpected send")
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
