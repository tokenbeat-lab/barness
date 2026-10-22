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
	current := ai.BuiltinCatalog()
	ev.Fixture("catalog-snapshot.json", raw)
	ev.Check("published catalog and price snapshot matches public discovery", reflect.DeepEqual(current, snapshot.Catalog) && catalogHash(t, current) == snapshot.Hash, "catalog snapshot is stale")
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
