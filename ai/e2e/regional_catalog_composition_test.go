package e2e

import (
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func TestRegionalQualificationCatalogComposition(t *testing.T) {
	ev := run.Case(t, "RP18-qualification-catalog-composition")
	candidate, err := regional.CandidateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	builtin := ai.BuiltinCatalog()
	// Remove only the nine target identities to exercise both possible formal
	// states with independent candidate facts, irrespective of current activation.
	var old []ai.Model
	for _, model := range builtin.Models {
		if _, owned := candidate.Lookup(model.Provider, model.API, model.ID); !owned {
			old = append(old, model)
		}
	}
	builtin.Models = old
	for _, active := range []bool{false, true} {
		base := builtin
		base.Models = append([]ai.Model{}, builtin.Models...)
		if active {
			base.Models = append(base.Models, candidate.Models...)
		}
		composed, err := regional.QualificationCatalog(base)
		if !ev.Check("compose both formal states", err == nil, "active=%t err=%v", active, err) {
			continue
		}
		ev.Check("nine identities included exactly once", len(composed.Models) == len(old)+9, "count=%d", len(composed.Models))
		for _, want := range candidate.Models {
			got, ok := composed.Lookup(want.Provider, want.API, want.ID)
			ev.Check("own facts preserved", ok && reflect.DeepEqual(got, want), "model=%s/%s/%s", want.Provider, want.API, want.ID)
		}
		w := newWorld(t, tenantA)
		_, err = ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &composed})
		ev.Check("composed facts construct public client", err == nil, "%v", err)
	}
	for _, mutation := range []string{"partial", "name", "price", "duplicate"} {
		base := builtin
		base.Models = append(append([]ai.Model{}, old...), candidate.Models...)
		switch mutation {
		case "partial":
			base.Models = base.Models[:len(old)+1]
		case "name":
			base.Models[len(old)].Name = "different accepted metadata"
		case "price":
			base.Models[len(old)].Cost.Input = 999
		case "duplicate":
			base.Models = append(base.Models, candidate.Models[0])
		}
		_, err := regional.QualificationCatalog(base)
		ev.Check("conflicting formal state refused "+mutation, err != nil, "silently accepted")
	}
}
