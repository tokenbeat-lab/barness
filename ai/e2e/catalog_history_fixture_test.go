package e2e

import (
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func TestBuiltinCatalogReleaseSnapshot(t *testing.T) {
	ev := run.Case(t, "catalog-release-snapshot")
	raw, err := os.ReadFile("../release/catalog-snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}
	mustUnmarshal(t, raw, &snapshot)
	current := maintainedFixtureCatalog(t)
	ev.Fixture("catalog-snapshot.json", raw)
	// This file remains the historical measurement pin. The migration changes
	// composition identity, while preserving every reviewed definition.
	ev.Check("historical catalog pin remains valid", catalogHash(t, snapshot.Catalog) == snapshot.Hash, "historical snapshot changed")
	projection := historicalCatalogInCurrentOrder(current, snapshot.Catalog)
	ev.Check("historical definitions and relative order survive", reflect.DeepEqual(projection, snapshot.Catalog), "catalog definitions or order changed")
}

// Historical inclusion tests share this projection because later independent
// identities must not change the exact content proof for any earlier identity.
// Values come only from public discovery, never from the expected snapshot.
func historicalCatalogProjection(t *testing.T, ev *evidence.Case, current, baseline ai.Catalog) ai.Catalog {
	t.Helper()
	out := ai.Catalog{Version: baseline.Version}
	if baseline.Models != nil {
		out.Models = make([]ai.Model, 0, len(baseline.Models))
	}
	for _, old := range baseline.Models {
		model, ok := current.Lookup(old.Provider, old.API, old.ID)
		if !ok {
			model, ok = maintainedFixtureCatalog(t).Lookup(old.Provider, old.API, old.ID)
		}
		ev.Check("historical chat identity survives "+string(old.Provider)+"/"+string(old.API)+"/"+old.ID, ok, "missing identity")
		out.Models = append(out.Models, model)
	}
	if baseline.ImageModels != nil {
		out.ImageModels = make([]ai.ImageModel, 0, len(baseline.ImageModels))
	}
	for _, old := range baseline.ImageModels {
		model, ok := current.LookupImage(old.Provider, old.API, old.ID)
		ev.Check("historical image identity survives "+old.ID, ok, "missing identity")
		out.ImageModels = append(out.ImageModels, model)
	}
	if baseline.ClassifierModels != nil {
		out.ClassifierModels = make([]ai.ClassifierModel, 0, len(baseline.ClassifierModels))
	}
	for _, old := range baseline.ClassifierModels {
		model, ok := current.LookupClassifier(old.Provider, old.API, old.ID)
		ev.Check("historical classifier identity survives "+old.ID, ok, "missing identity")
		out.ClassifierModels = append(out.ClassifierModels, model)
	}
	return out
}

// Filter in current order so additions cannot conceal reordering historical identities.
func historicalCatalogInCurrentOrder(current, baseline ai.Catalog) ai.Catalog {
	out := ai.Catalog{Version: baseline.Version, Models: make([]ai.Model, 0), ImageModels: make([]ai.ImageModel, 0), ClassifierModels: make([]ai.ClassifierModel, 0)}
	for _, m := range current.Models {
		if _, ok := baseline.Lookup(m.Provider, m.API, m.ID); ok {
			out.Models = append(out.Models, m)
		}
	}
	for _, m := range current.ImageModels {
		if _, ok := baseline.LookupImage(m.Provider, m.API, m.ID); ok {
			out.ImageModels = append(out.ImageModels, m)
		}
	}
	for _, m := range current.ClassifierModels {
		if _, ok := baseline.LookupClassifier(m.Provider, m.API, m.ID); ok {
			out.ClassifierModels = append(out.ClassifierModels, m)
		}
	}
	return out
}
