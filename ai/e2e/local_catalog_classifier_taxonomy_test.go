package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/catalog"
)

// The classifier protocol, rather than a particular model name, defines its
// capability subset. Invalid values and unsupported routes stay distinct.
func TestLocalCatalogClassifierFailureTaxonomy(t *testing.T) {
	var jev, decisions ai.ClassifierModel
	for _, m := range ai.BuiltinCatalog().ClassifierModels {
		if m.ID == "jev-1.14.0" {
			jev = m
		}
		if m.Provider == ai.ProviderOpenAI && m.API == ai.APIOpenAIDecisions {
			decisions = m
		}
	}
	if jev.ID == "" || decisions.ID == "" {
		t.Fatal("independent catalog baseline identities unavailable")
	}
	for _, mode := range []string{"fixed-jev", "future-jev-images-box", "unsupported-box-protocol", "invalid-image-limit"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "CAT-T08-classifier-taxonomy-"+mode)
			model := jev
			wantReason := catalog.Reason("")
			wantError := ""
			switch mode {
			case "future-jev-images-box":
				model.ID = "jev-future-catalog-fixture"
			case "unsupported-box-protocol":
				model = decisions
				model.Capabilities.Kinds = append(append([]ai.ClassifierQuestionKind(nil), model.Capabilities.Kinds...), ai.ClassifierQuestionBox)
				wantReason = catalog.UnsupportedDefinition
				wantError = "ai: invalid config Catalog: classifier model " + model.ID + ": Box requires the SystemOne image protocol and a visible image location"
			case "invalid-image-limit":
				model.Capabilities.MaxInputImages = 0
				wantReason = catalog.InvalidDefinition
				wantError = "ai: invalid config Catalog: classifier model jev-1.14.0: classifier image declarations are inconsistent"
			}
			input := ai.Catalog{Version: "classifier-taxonomy", Models: []ai.Model{}, ClassifierModels: []ai.ClassifierModel{model}}
			body := mustMarshal(t, input)
			raw := localCatalogFile(string(body))
			effective, report, err := catalog.Load(raw)
			ev.Record("source", json.RawMessage(raw))
			ev.Record("report", report)
			ev.Check("model failures stay isolated from file validity", err == nil && len(report.Models) == 1, "err=%v report=%+v", err, report)
			if err != nil || len(report.Models) != 1 {
				t.FailNow()
			}
			ev.Check("unsupported subset differs from invalid values", report.Models[0].Reason == wantReason, "got=%s want=%s", report.Models[0].Reason, wantReason)
			wantLoaded := 0
			if wantReason == "" {
				wantLoaded = 1
			}
			ev.Check("only implemented valid model is usable", report.Loaded == wantLoaded && report.Failed == 1-wantLoaded && len(effective.ClassifierModels) == wantLoaded, "report=%+v", report)
			_, hashErr := input.Hash()
			h := host.New()
			_, clientErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &input})
			ev.Check("legacy Hash error and acceptance unchanged", errString(hashErr) == wantError, "error=%v want=%s", hashErr, wantError)
			ev.Check("legacy Client error and acceptance unchanged", errString(clientErr) == wantError, "error=%v want=%s", clientErr, wantError)
		})
	}
}
