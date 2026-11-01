package e2e

import (
	"os"
	"testing"
)

// The generated builtin artifact carries the current catalog identity.
// Provider facts keep their own immutable source dates; unrelated catalog
// additions must not make an old model's capabilities appear newly sourced.
func currentCatalogVersion(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct{ Catalog struct{ Version string } }
	mustUnmarshal(t, raw, &snapshot)
	if snapshot.Catalog.Version == "" {
		t.Fatal("generated builtin catalog version missing")
	}
	return snapshot.Catalog.Version
}
