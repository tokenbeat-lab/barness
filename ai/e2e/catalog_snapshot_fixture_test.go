package e2e

import (
	"os"
	"testing"
)

// The checked-in release snapshot is the independent current catalog pin.
// Provider facts keep their own immutable source dates; unrelated catalog
// additions must not make an old model's capabilities appear newly sourced.
func currentCatalogVersion(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../release/catalog-snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct{ Catalog struct{ Version string } }
	mustUnmarshal(t, raw, &snapshot)
	if snapshot.Catalog.Version == "" {
		t.Fatal("current release catalog version missing")
	}
	return snapshot.Catalog.Version
}
