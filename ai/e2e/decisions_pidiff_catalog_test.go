package e2e

import (
	"runtime"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

func TestDecisionsPiCatalog(t *testing.T) {
	ev := run.Case(t, "PIDIFF-P10-OD48-catalog")
	ev.ReplayEnv(pioracle.EnableEnv + "=1")
	o := openOracle(t)
	models, err := o.Models(ctxFor(t), []pioracle.ModelKey{{Provider: "openai", API: "openai-decisions", ID: "gpt-6-luna"}, {Provider: "openai", API: "openai-responses", ID: "gpt-6-luna"}})
	if !ev.Check("separate classifier and chat catalogs", err == nil && len(models) == 2 && string(models[0]) != "null" && string(models[1]) != "null", "%v", err) {
		return
	}
	classifier, chat := jsonObject(t, models[0]), jsonObject(t, models[1])
	ev.Check("same named types remain independent", classifier["type"] == "classifier" && chat["type"] == "chat" && classifier["api"] == "openai-decisions" && chat["api"] == "openai-responses", "got %s", mustMarshal(t, models))
	u := newClassifierImagesWorld(t, func(c *ai.Config) {
		m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
		m.ID, m.Name, m.ContextWindow = "gpt-6-luna", "GPT-6 Luna", 922000
		m.Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore, ai.ClassifierQuestionBool}
		m.Capabilities.MinScoreLevels, m.Capabilities.MaxScoreLevels = 2, 10
		m.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: .1}, Tiers: []ai.CostTier{{InputTokensAbove: 272000, CostRates: ai.CostRates{Input: .2}}}}
	})
	cat := u.client.Catalog()
	m, found := cat.LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIDecisions, "gpt-6-luna")
	ev.Check("public operation lookup selects classifier", found, "not found")
	bmodel := jsonObject(t, m)
	bmodel["type"] = "classifier"
	// Keep the complete host model, including additive capability metadata.
	// Binding owns the endpoint. Neither catalog is pruned wholesale.
	b := pioracle.Observation{Result: mustMarshal(t, bmodel)}
	pi := pioracle.Observation{Result: models[0]}
	ev.Check("frozen endpoint remains official", classifier["baseUrl"] == "https://api.openai.com/v1", "got %v", classifier["baseUrl"])
	ev.Check("host capability metadata has its own public evidence", m.Capabilities.MinChoices == 2 && m.Capabilities.MaxChoices == 255 && m.Capabilities.MaxInputImages == 128 && m.Capabilities.MaxScoreLevels == 10, "got %+v", m.Capabilities)
	ev.Record("pi-catalog", models)
	ev.Record("barness-catalog", cat)
	ev.Record("pi-observation", pi)
	ev.Record("barness-observation", b)
	diffs, err := pioracle.Compare(pi, b)
	if !ev.Check("full catalogs compare", err == nil, "%v", err) {
		return
	}
	v := loadLedger(t).Classify("openai-decisions", ev.ID(), diffs)
	checkDecisionsLedger(ev, diffs)
	ev.Record("pidiff", o.Record(ev.ID(), "openai-decisions", v, pioracle.Sides{BarnessVersions: map[string]string{"go": runtime.Version()}, ModelCatalogHash: catalogHash(t, cat), PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), PiResult: pi.Result, BarnessResult: b.Result, Frames: mustMarshal(t, models)}))
	ev.Check("catalog zero pending", v.Pass, "%+v", v.Findings)
}

// OD50: adding classifier metadata cannot hide old Jev fields from the
// catalog projection. Preserve all old fields and explicitly prove omission
// of only the new optional metadata, as required by ADR-0026 / issue 01.
func TestJevDifferentialMetadata(t *testing.T) {
	ev := run.Case(t, "P07-OD50-differential-metadata-regression")
	m, ok := ai.BuiltinCatalog().LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.13.0")
	ev.Check("fixed official Jev still listed", ok, "not found")
	raw := jsonObject(t, m)
	ev.Record("catalog-model", m)
	expected := []byte(`{"provider":"typesafe","api":"typesafe-system-one","id":"jev-1.13.0","name":"Jev","contextWindow":64000,"capabilities":{"kinds":["choice","score","bool"],"maxChoices":255,"minScoreLevels":2,"maxScoreLevels":10},"cost":{"input":0.042,"output":0,"cacheRead":0,"cacheWrite":0}}`)
	ev.Check("complete old Jev serialization retained", jsonEqual(mustMarshal(t, raw), expected), "got %s", mustMarshal(t, raw))
}
