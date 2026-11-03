package e2e

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// Protocol regressions replay their original model definitions explicitly.
// The immutable pre-retirement approval remains valid input to custom Catalog;
// current default discovery and withdrawal are tested independently.
func protocolFixtureCatalog() ai.Catalog {
	raw, err := os.ReadFile("testdata/catalog/kimi-k31-approved.json")
	if err != nil {
		panic(err)
	}
	var snapshot struct {
		Hash    string
		Catalog ai.Catalog
	}
	if err = json.Unmarshal(raw, &snapshot); err != nil {
		panic(err)
	}
	hash, err := snapshot.Catalog.Hash()
	if err != nil || hash != snapshot.Hash {
		panic("invalid historical protocol fixture catalog")
	}
	return snapshot.Catalog
}

// Withdrawal preserves reviewed definitions in maintenance State. Historical
// preservation assertions inspect those values without restoring runtime IDs.
func maintainedFixtureCatalog(t *testing.T) ai.Catalog {
	t.Helper()
	state, err := maintenance.BuiltinState()
	if err != nil {
		t.Fatal(err)
	}
	for pi := range state.Providers {
		for ei := range state.Providers[pi].Entries {
			state.Providers[pi].Entries[ei].Lifecycle = maintenance.Active
		}
	}
	raw, err := maintenance.Generate(state)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct{ Catalog ai.Catalog }
	mustUnmarshal(t, raw, &snapshot)
	return snapshot.Catalog
}
