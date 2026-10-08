package e2e

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Official facts and this route's live wire, rather than another provider's
// pass, are the independent oracle for inclusion. Aliases must not enter it.
func TestClassifierBuiltinOfficial(t *testing.T) {
	ev := run.Case(t, "P07-E11-builtin-official")
	raw, err := os.ReadFile("testdata/typesafe/official-capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("official-capabilities.json", raw)
	var facts struct{ Model ai.ClassifierModel }
	mustUnmarshal(t, raw, &facts)
	catalog := ai.BuiltinCatalog()
	ev.Record("catalog", catalog)
	m, ok := catalog.LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.13.0")
	ev.Check("fixed version has official capabilities and rates", ok && reflect.DeepEqual(m, facts.Model), "got %+v", m)
	ev.Check("only first fixed classifier included", len(catalog.ClassifierModels) == 1, "listed %d", len(catalog.ClassifierModels))
	for _, id := range []string{"jev-latest", "jev-preview"} {
		_, ok := catalog.LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, id)
		ev.Check("moving alias excluded "+id, !ok, "alias listed")
	}
	ev.Check("content revision has its own version", catalog.Version == "2026-10-08.4", "version %s", catalog.Version)
	// Mutating returned capabilities cannot change the next host's snapshot.
	if len(m.Capabilities.Kinds) > 0 {
		m.Capabilities.Kinds[0] = "invalid"
	}
	fresh, ok := ai.BuiltinCatalog().LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.13.0")
	ev.Check("fresh catalog owns its capabilities", ok && reflect.DeepEqual(fresh, facts.Model), "shared state")
	var roundtrip ai.Catalog
	mustUnmarshal(t, mustMarshal(t, catalog), &roundtrip)
	ev.Check("official classifier roundtrip", reflect.DeepEqual(catalog, roundtrip), "roundtrip changed")
}

func TestClassifierLiveWireReplay(t *testing.T) {
	raw, err := os.ReadFile("testdata/typesafe/live-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			ID       string               `json:"id"`
			Request  ai.ClassifierRequest `json:"request"`
			Status   int                  `json:"status"`
			Headers  map[string]string    `json:"responseHeaders"`
			Response json.RawMessage      `json:"response"`
		}
	}
	mustUnmarshal(t, raw, &fixture)
	for _, sc := range fixture.Cases {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P07-E11-live-wire-"+sc.ID)
			ev.Fixture("live-contract.json", raw)
			w := newWorldWith(t, func(c *ai.Config) {
				c.Policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 3, MaxStateBytes: 131072, MaxQuestionBytes: 16384}
			}, tenantA)
			installClassifierBinding(w, tenantA)
			w.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowedModels = []string{"jev-1.13.0"} })
			enqueue(ev, w, fixtureReply{Status: sc.Status, Header: sc.Headers, Body: string(sc.Response)}.script(t, classifierProtocol, ""))
			res, err := w.client.Classify(ctxFor(t), textScope("req-live-wire-"+sc.ID), ai.Target{BindingID: "classifier", ModelID: "jev-1.13.0"}, sc.Request, nil)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			ev.Record("error", errString(err))
			ev.Check("one authorized classifier attempt", res.Metadata.Operation == ai.OperationClassifier && res.Metadata.ModelID == "jev-1.13.0" && len(res.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "attempts %+v", res.Metadata.Attempts)
			if len(res.Metadata.Attempts) != 1 {
				return
			}
			a := res.Metadata.Attempts[0]
			ev.Check("captured official status and request ID", a.HTTPStatus == sc.Status && a.ProviderRequestID == sc.Headers["x-typesafe-request-id"], "attempt %+v", a)
			if sc.Status == 200 {
				ev.Check("real success converts complete typed answers", err == nil && len(res.Answers) == len(sc.Request.Questions) && a.UsageReporting == ai.UsageComplete, "err=%v", err)
				m, ok := res.ResponseModel.Get()
				ev.Check("real version retained", ok && m == "jev-1.13.0", "model %s", m)
				ev.Check("official input-only estimate", res.Usage.Cost.Output == 0 && math.Abs(res.Usage.Cost.Input-float64(res.Usage.Input)*0.042/1e6) < 1e-12, "usage %+v", res.Usage)
			} else {
				ev.Check("real context guard classification", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseRequest}) && sc.Status == 400 && len(res.Answers) == 0, "err=%v", err)
				ev.Check("missing error usage stays unknown", a.UsageReporting == ai.UsageUnreported, "reporting %s", a.UsageReporting)
			}
		})
	}
}

func TestClassifierLiveScoreRounding(t *testing.T) {
	raw, err := os.ReadFile("testdata/typesafe/live-contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Request  ai.ClassifierRequest
			Response json.RawMessage
		}
	}
	mustUnmarshal(t, raw, &fixture)
	for _, tc := range []struct {
		name   string
		score  float64
		accept bool
	}{{"observed-hundredths", 1.32, true}, {"opposite-rounding", 1.30, true}, {"beyond-rounding", 1.34, false}, {"higher-precision-mismatch", 1.315, false}} {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "P07-E02-live-score-"+tc.name)
			w := newWorldWith(t, func(c *ai.Config) {
				c.Policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 3, MaxStateBytes: 131072, MaxQuestionBytes: 16384}
			}, tenantA)
			installClassifierBinding(w, tenantA)
			w.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowedModels = []string{"jev-1.13.0"} })
			var body map[string]any
			mustUnmarshal(t, fixture.Cases[0].Response, &body)
			// The independent observed vector is .01/.67/.32 (weighted 1.31).
			body["answers"].(map[string]any)["severity"].(map[string]any)["score"] = tc.score
			enqueue(ev, w, fixtureReply{Status: 200, Header: map[string]string{"x-typesafe-request-id": "test-rounding"}, Body: string(mustMarshal(t, body))}.script(t, classifierProtocol, ""))
			res, err := w.client.Classify(ctxFor(t), textScope("req-rounding-"+tc.name), ai.Target{BindingID: "classifier", ModelID: "jev-1.13.0"}, fixture.Cases[0].Request, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("only bounded hundredth rounding accepted", (err == nil) == tc.accept, "err=%v", err)
			if !tc.accept {
				ev.Check("invalid answer is atomic and preserves known usage", len(res.Answers) == 0 && res.Usage.Input > 0, "usage %+v", res.Usage)
			}
		})
	}
}
