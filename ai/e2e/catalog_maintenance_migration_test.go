package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// The historical snapshot is an independent source: expected definitions do
// not come from the maintenance source or the generated builtin artifact.
func TestMaintenanceCatalogMigration(t *testing.T) {
	ev := run.Case(t, "CATALOG-T01-maintenance-migration")
	historical, err := os.ReadFile("../release/catalog-snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("historical-2026-10-29.1.json", historical)
	var old struct {
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}
	mustUnmarshal(t, historical, &old)
	ev.Check("historical identity is unchanged", old.Catalog.Version == "2026-10-29.1" && old.Hash == "sha256:f44d07c6969346f056cc46d31fccbafb68827267910801c05bb314ac0c55348d", "historical source changed")
	state, err := maintenance.BuiltinState()
	if err != nil {
		t.Fatal(err)
	}
	generated, err := maintenance.Generate(state)
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("generated-catalog.json", generated)
	var snapshot struct {
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}
	mustUnmarshal(t, generated, &snapshot)
	current := ai.BuiltinCatalog()
	ev.Check("new identity is explicit", current.Version != old.Catalog.Version && current.Version == snapshot.Catalog.Version, "version=%s", current.Version)
	historicalCurrent := historicalCatalogInCurrentOrder(current, old.Catalog)
	ev.Check("historical 118 records retain definitions and order", len(historicalCurrent.Models) == 112 && len(historicalCurrent.ImageModels) == 3 && len(historicalCurrent.ClassifierModels) == 3 && reflect.DeepEqual(historicalCurrent, old.Catalog), "historical definitions or order differ")
	ev.Check("builtin and external JSON have the same definitions and hash", reflect.DeepEqual(current, snapshot.Catalog) && catalogHash(t, current) == snapshot.Hash, "generated/builtin mismatch")
	checkedIn, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("checked-in artifact is exactly reproducible", bytes.Equal(checkedIn, generated), "generation is stale")
	again, err := maintenance.Generate(state)
	ev.Check("generation is deterministic", err == nil && bytes.Equal(generated, again), "err=%v", err)
	gaps, retained := 0, 0
	for _, p := range state.Providers {
		for _, entry := range p.Entries {
			if entry.Source == nil || entry.Source.Kind != "historical" {
				continue
			}
			retained++
			ev.Check("every record retains historical provenance", entry.Source != nil && entry.Source.Digest == old.Hash && entry.Source.Path == "ai/release/catalog-snapshot.json", "source=%+v", entry.Source)
			if entry.Source != nil && entry.Source.MissingServing {
				gaps++
			}
		}
	}
	ev.Check("six image/classifier serving gaps remain explicit", retained == 118 && gaps == 6, "records=%d gaps=%d", retained, gaps)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &snapshot.Catalog }, tenantA)
	ev.Check("generated snapshot reaches the public Client", reflect.DeepEqual(w.client.Catalog(), current), "client catalog differs")
	// Mutating a returned catalog or source must not change future defaults.
	current.Models[0].Input[0] = ai.ModalityImage
	state.Providers[0].Entries[0].Definition = json.RawMessage(`{}`)
	ev.Check("public defaults remain independent", reflect.DeepEqual(ai.BuiltinCatalog(), snapshot.Catalog), "builtin shares mutable state")
}
