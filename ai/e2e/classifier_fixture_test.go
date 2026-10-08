package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

var classifierProtocol = &fixtureProtocol{dir: "typesafe", binding: "classifier", api: ai.APITypeSafeSystemOne, provider: ai.ProviderTypeSafe, key: func(k tenantKey) tenantKey { return k }, account: "acct-tenant-a"}

// The host catalog and price are synthetic. Vendor inclusion and live evidence
// belong to issue 08, so this suite makes no built-in support claim.
func configureClassifier(c *ai.Config) {
	cat := ai.BuiltinCatalog()
	cat.ClassifierModels = []ai.ClassifierModel{{Provider: ai.ProviderTypeSafe, API: ai.APITypeSafeSystemOne, ID: "host-jev", Name: "Synthetic classifier", ContextWindow: 64000, Capabilities: ai.ClassifierCapabilities{Kinds: []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore, ai.ClassifierQuestionBool}, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10}, Cost: ai.ModelCost{CostRates: ai.CostRates{Input: 1}}}}
	c.Catalog = &cat
	c.Policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
}
func installClassifierBinding(w *world, k tenantKey) {
	b := primaryBinding(k, w.provider.URL())
	b.BindingID = "classifier"
	b.ProviderID = ai.ProviderTypeSafe
	b.API = ai.APITypeSafeSystemOne
	b.Operation = ai.OperationClassifier
	b.AllowedModels = []string{"host-jev"}
	w.host.PutBinding(b)
}
func classifierInput(t *testing.T, sc fixtureScenario) ai.ClassifierRequest {
	t.Helper()
	var req ai.ClassifierRequest
	mustUnmarshal(t, sc.Classifier, &req)
	return req
}
func invokeClassifierScenario(t *testing.T, ev *evidence.Case, w *world, sc fixtureScenario) {
	t.Helper()
	var opts ai.Options
	if sc.ForeignOptions {
		opts = ai.ResponsesOptions{}
	}
	res, err := w.client.Classify(ctxFor(t), textScope("req-"+sc.ID), sc.target(), classifierInput(t, sc), opts)
	checkClassifierScenario(ev, w, sc, res, err)
}
func checkClassifierScenario(ev *evidence.Case, w *world, sc fixtureScenario, res ai.ClassifierResult, err error) {
	ev.Record("result", res)
	ev.Record("error", errString(err))
	ev.Record("requests", w.provider.Requests())
	x := sc.Expect
	reqs := w.provider.Requests()
	want := max(x.Requests, 0)
	ev.Check("request count", len(reqs) == want, "got %d want %d", len(reqs), want)
	if want > 0 && len(reqs) > 0 && x.Request != nil {
		r := reqs[len(reqs)-1]
		ev.Check("request path and tenant key", r.Method == "POST" && r.Path == x.Request.Path && r.KeyAlias == tenantA.alias, "got %+v", r)
		ev.Check("request JSON", jsonEqual(r.Body, x.Request.Body), "got %s", r.Body)
		for _, number := range []string{"9007199254740993", "12345678901234567890"} {
			if count := strings.Count(string(x.Request.Body), number); count > 0 {
				ev.Check("native numeric precision "+number, strings.Count(string(r.Body), number) == count, "got %s", r.Body)
			}
		}
	}
	ev.Check("terminal", string(res.StopReason) == x.StopReason, "got %q", res.StopReason)
	if x.Code != "" {
		var ae *ai.Error
		ev.Check("classified failure", errors.As(err, &ae) && ae.Code == x.Code && ae.Phase == x.Phase, "got %v", err)
		ev.Check("error message retained", ae != nil && ae.Message == res.ErrorMessage, "got %v", err)
		if ae != nil {
			ev.Check("HTTP error metadata", ae.HTTPStatus == x.HTTPStatus && ae.ProviderRequestID == x.ProviderRequestID, "got %+v", ae)
		}
	} else {
		ev.Check("success", err == nil, "got %v", err)
	}
	if x.Answers != nil && string(x.Answers) != "null" {
		ev.Check("typed answers", jsonEqual(mustMarshal(ev.T(), res.Answers), x.Answers), "got %+v", res.Answers)
		a, ok := res.Answers["intent"].(ai.ChoiceAnswer)
		ev.Check("answer Go type", ok && a.Choice == "track", "got %T", res.Answers["intent"])
		if _, requested := classifierInput(ev.T(), sc).Questions["anger"]; requested {
			_, ok := res.Answers["anger"].(ai.ScoreAnswer)
			ev.Check("score Go type", ok, "got %T", res.Answers["anger"])
		}
		if _, requested := classifierInput(ev.T(), sc).Questions["urgent"]; requested {
			_, ok := res.Answers["urgent"].(ai.BoolAnswer)
			ev.Check("bool Go type", ok, "got %T", res.Answers["urgent"])
		}
	} else {
		ev.Check("failed answers empty", len(res.Answers) == 0, "got %+v", res.Answers)
	}
	model, _ := res.ResponseModel.Get()
	ev.Check("response model", model == x.ResponseModel, "got %q", model)
	if x.Usage != nil {
		ev.Check("usage retained and input priced", res.Usage == *x.Usage, "got %+v want %+v", res.Usage, *x.Usage)
	}
	meta := res.Metadata
	ev.Check("operation attributed", meta.Operation == ai.OperationClassifier, "got %+v", meta)
	ev.Check("attempts", len(meta.Attempts) == want, "got %+v", meta.Attempts)
	if want > 0 {
		ev.Check("authorized identity retained", meta.Resolved && meta.ModelID == sc.Model && meta.ProviderID == ai.ProviderTypeSafe && meta.API == ai.APITypeSafeSystemOne, "got %+v", meta)
		if len(meta.Attempts) > 0 {
			a := meta.Attempts[len(meta.Attempts)-1]
			ev.Check("usage reporting", a.UsageReporting == x.UsageReporting && a.Usage == res.Usage, "got %+v", a)
			ev.Check("vendor id", a.ProviderRequestID == "req-typesafe", "got %+v", a)
		}
	} else {
		ev.Check("preflight unresolved", !meta.Resolved, "got %+v", meta)
		ev.Check("no credential read", w.host.ReadsFor("req-"+sc.ID).Credentials == 0, "got %+v", w.host.ReadsFor("req-"+sc.ID))
	}
}
func TestClassifierChoiceFixtures(t *testing.T) {
	f, raw := loadFixture(t, classifierProtocol, "choice.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) { ev := run.Case(t, "P07-"+sc.ID); runScenario(t, ev, sc, raw, modeComplete) })
	}
}

func TestClassifierMixedFixtures(t *testing.T) {
	f, raw := loadFixture(t, classifierProtocol, "mixed.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) { ev := run.Case(t, "P07-"+sc.ID); runScenario(t, ev, sc, raw, modeComplete) })
	}
}

func TestClassifierParityFixtures(t *testing.T) {
	f, raw := loadFixture(t, classifierProtocol, "parity.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P07-"+sc.ID+"-parity")
			runScenario(t, ev, sc, raw, modeComplete)
		})
	}
}
