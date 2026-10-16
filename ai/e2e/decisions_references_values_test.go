package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsReferencesValues(t *testing.T) {
	for _, sc := range []struct {
		name, state, question, pointer string
		local                          bool
	}{
		{"state-root", `null`, ``, `/state`, true},
		{"state-null", `{"photo":null,"ordinary":null}`, ``, `/state/photo`, true},
		{"escaped-keys", `{"a/b":{"~":{"":null}}}`, ``, `/state/a~1b/~0/`, true},
		{"numeric-object-key", `{"01":null}`, ``, `/state/01`, true},
		{"array", `[null,{"x":null}]`, ``, `/state/1/x`, true},
		{"literal-percent", `{"%2F":null}`, ``, `/state/%2F`, true},
		{"unicode", `{"图":null}`, ``, `/state/图`, true},
		{"instructions-root", `{}`, `{"type":"choice","instructions":null,"criteria":{"a":null,"b":null}}`, `/questions/q/instructions`, true},
		{"escaped-question", `{}`, `{"type":"choice","instructions":["compare",null],"criteria":{"a":null,"b":null}}`, `/questions/q~1~0/instructions/1`, true},
		{"instructions-array", `{}`, `{"type":"choice","instructions":["compare",null],"criteria":{"a":null,"b":null}}`, `/questions/q/instructions/1`, true},
		{"choice", `{}`, `{"type":"choice","instructions":"classify","criteria":{"a/b":{"image":null},"b":null}}`, `/questions/q/criteria/a~1b/image`, true},
		{"choice-root", `{}`, `{"type":"choice","instructions":"classify","criteria":{"a":null,"b":null}}`, `/questions/q/criteria/a`, true},
		{"score-root", `{}`, `{"type":"score","instructions":"rate","criteria":[null,"other"]}`, `/questions/q/criteria/0`, true},
		{"bool-root", `{}`, `{"type":"bool","instructions":"judge","criteria":{"true":null,"false":"other"}}`, `/questions/q/criteria/true`, true},
		{"bool-descendant", `{}`, `{"type":"bool","instructions":"judge","criteria":{"true":{"x":[null]},"false":"other"}}`, `/questions/q/criteria/true/x/0`, true},
		{"not-null", `{"photo":"text"}`, ``, `/state/photo`, false},
		{"absent", `{}`, ``, `/state/photo`, false},
		{"array-leading-zero", `[null]`, ``, `/state/00`, false},
		{"array-negative", `[null]`, ``, `/state/-1`, false},
		{"array-append", `[null]`, ``, `/state/-`, false},
		{"array-oob", `[null]`, ``, `/state/1`, false},
		{"bad-escape", `{"~2":null}`, ``, `/state/~2`, false},
		{"fragment", `null`, ``, `#/state`, false},
		{"root", `{}`, ``, `/`, false},
		{"questions-container", `{}`, ``, `/questions`, false},
		{"question-container", `{}`, ``, `/questions/q`, false},
		{"criteria-container", `{}`, `{"type":"choice","instructions":"classify","criteria":{"a":null,"b":null}}`, `/questions/q/criteria`, false},
		{"type-field", `{}`, ``, `/questions/intent/type`, false},
		{"images-field", `{}`, ``, `/images/0`, false},
		{"references-field", `{}`, ``, `/imageReferences/0/pointer`, false},
		{"unknown-field", `{}`, ``, `/allowClassifierImages`, false},
		{"ordinary-state-null", `null`, ``, `/questions/intent/criteria/other`, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC02-value-"+sc.name)
			u := newClassifierImagesWorld(t, nil)
			req := decisionsInput()
			req.State = json.RawMessage(sc.state)
			req.Images = classifierAssets(t)[:1]
			if sc.question != "" {
				raw := `{"state":{},"questions":{"q":` + sc.question + `}}`
				mustUnmarshal(t, []byte(raw), &req)
				req.State = json.RawMessage(sc.state)
				req.Images = classifierAssets(t)[:1]
			}
			if sc.name == "escaped-question" {
				req.Questions["q/~"] = req.Questions["q"]
				delete(req.Questions, "q")
			}
			refs := `[{"imageIndex":0,"placement":"context"},{"imageIndex":0,"placement":"value","pointer":` + string(mustMarshal(t, sc.pointer)) + `}]`
			req = decisionsReferenceRequest(t, req, refs)
			ev.Fixture("public-request.json", mustMarshal(t, req))
			res, err := u.client.Classify(ctxFor(t), textScope("req-value-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			phase := ai.PhaseScope
			if sc.local {
				phase = ai.PhaseCapability
			}
			checkUnaryFailure(ev, res, err, ai.CodeInvalidRequest, phase, ai.StopReasonError)
			ev.Check("invalid or unsupported reference never pins credentials or sends", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsReferencesInvalidPlan(t *testing.T) {
	for _, sc := range []struct {
		name, refs string
		images     int
	}{
		{"dangling", `[{"imageIndex":1,"placement":"context"}]`, 1},
		{"no-source", `[{"imageIndex":0,"placement":"context"}]`, 0},
		{"unused-source", `[{"imageIndex":0,"placement":"context"}]`, 2},
		{"duplicate-value", `[{"imageIndex":0,"placement":"value","pointer":"/state/photo"},{"imageIndex":0,"placement":"value","pointer":"/state/photo"}]`, 1},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC02-invalid-plan-"+sc.name)
			u := newClassifierImagesWorld(t, nil)
			req := decisionsInput()
			req.State = json.RawMessage(`{"photo":null}`)
			req.Images = classifierAssets(t)[:sc.images]
			req = decisionsReferenceRequest(t, req, sc.refs)
			ev.Fixture("public-request.json", mustMarshal(t, req))
			res, err := u.client.Classify(ctxFor(t), textScope("req-plan-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			checkUnaryFailure(ev, res, err, ai.CodeInvalidRequest, ai.PhaseScope, ai.StopReasonError)
			ev.Check("no credential or Attempt", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
