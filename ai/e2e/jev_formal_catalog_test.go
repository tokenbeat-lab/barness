package e2e

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevFormalCatalog(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC13-formal-catalog-"+entry)
			cat := ai.BuiltinCatalog()
			m, ok := cat.LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, "jev-1.14.0")
			ev.Record("catalog", cat)
			if !ev.Check("Host discovers independently accepted fixed Jev", ok, "formal Jev 1.14 absent") {
				return
			}
			ev.Check("four explicit image positions and Box", reflect.DeepEqual(m.Capabilities.ImageLocations, []ai.ClassifierImageLocation{"context", "state", "instructions", "criteria"}) && slices.Contains(m.Capabilities.Kinds, ai.ClassifierQuestionBox) && reflect.DeepEqual(m.Capabilities.ImageMIMETypes, []string{"image/png", "image/jpeg"}) && m.Capabilities.MaxInputImages == 24 && m.Cost.Input == 0.042 && m.Cost.Output == 0, "unsupported declaration")
			for _, alias := range []string{"jev-latest", "jev-preview"} {
				_, found := cat.LookupClassifier(ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, alias)
				ev.Check("moving alias excluded "+alias, !found, "alias expanded")
			}
			var before struct {
				Hash    string
				Catalog ai.Catalog
			}
			raw, err := os.ReadFile("testdata/typesafe/formal-before-1.14.json")
			if err != nil {
				t.Fatal(err)
			}
			mustUnmarshal(t, raw, &before)
			trimmed := cat
			// GLM adds independent chat identities after this historical Jev
			// inclusion baseline; compare the exact unchanged prior model set.
			trimmed.Models = slices.DeleteFunc(slices.Clone(cat.Models), func(m ai.Model) bool { return m.Provider == ai.ProviderZhipu })
			trimmed.ClassifierModels = slices.DeleteFunc(slices.Clone(cat.ClassifierModels), func(m ai.ClassifierModel) bool { return m.ID == "jev-1.14.0" })
			trimmed.Version = before.Catalog.Version
			ev.Check("old model content and hash preserved", reflect.DeepEqual(trimmed, before.Catalog) && catalogHash(t, trimmed) == before.Hash && cat.Version != before.Catalog.Version && catalogHash(t, cat) != before.Hash, "historical model changed")
			u := newJevBoxWorld(t, ev, func(c *ai.Config) { *c.Catalog = cat })
			req := jevBoxInput(t)
			enqueueJevBox(t, ev, u, jevBoxAnswer, "0.9")
			var res ai.ClassifierResult
			if entry == "plain" {
				res, err = u.client.Classify(ctxFor(t), textScope("req-jev-formal-"+entry), jevContextTarget(), req, jevContextOptions())
			} else {
				res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-jev-formal-"+entry), jevContextTarget(), req, jevContextOptions())
			}
			ev.Check("formal catalog supports public Box and image identity", err == nil && len(res.Answers) == 2 && len(res.ImageInputs) == 1 && res.Metadata.CatalogHash == catalogHash(t, cat), "call failed: %v", err)
			recordJevContext(ev, u, req, res, err)
			borrowed := u.client.Catalog()
			for i := range borrowed.ClassifierModels {
				if borrowed.ClassifierModels[i].ID == m.ID {
					borrowed.ClassifierModels[i].Capabilities.ImageLocations[0] = "criteria"
					borrowed.ClassifierModels[i].Capabilities.ImageMIMETypes[0] = "image/webp"
					borrowed.ClassifierModels[i].Capabilities.Kinds[0] = "bad"
				}
			}
			fresh, _ := u.client.Catalog().LookupClassifier(m.Provider, m.API, m.ID)
			ev.Check("borrowed discovery cannot change pinned model", reflect.DeepEqual(fresh, m), "shared capabilities")
			var round ai.Catalog
			mustUnmarshal(t, mustMarshal(t, cat), &round)
			ev.Check("formal catalog round trips", reflect.DeepEqual(cat, round), "JSON drift")
		})
	}
}
