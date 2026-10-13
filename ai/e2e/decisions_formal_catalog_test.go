package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// OD52 public catalog seam: inclusion must not rewrite the accepted candidate,
// import chat prices, materialize Jev defaults, or change any existing entry.
func TestDecisionsFormalCatalog(t *testing.T) {
	ev := run.Case(t, "P10-OD52-formal-catalog")
	raw, err := os.ReadFile("testdata/decisions/formal-before.json")
	if err != nil {
		t.Fatal(err)
	}
	var before struct {
		Hash    string
		Catalog ai.Catalog
	}
	mustUnmarshal(t, raw, &before)
	cat := ai.BuiltinCatalog()
	m, ok := cat.LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIDecisions, "gpt-6-luna")
	ev.Record("formal-catalog", cat)
	ev.Check("Host discovers formal classifier", ok, "Luna absent")
	if !ok {
		return
	}
	raw, err = os.ReadFile("../live/testdata/decisions/candidate.json")
	if err != nil {
		t.Fatal(err)
	}
	var candidate ai.Catalog
	mustUnmarshal(t, raw, &candidate)
	ev.Check("accepted sourced capabilities and input-only prices", reflect.DeepEqual(m, candidate.ClassifierModels[0]), "changed model")
	hash := catalogHash(t, cat)
	ev.Check("fresh formal version and hash", cat.Version == "2026-10-13.1" && hash != before.Hash && hash != catalogHash(t, candidate), "wrong formal identity")
	cat.ClassifierModels = cat.ClassifierModels[:len(cat.ClassifierModels)-1]
	cat.Version = before.Catalog.Version
	ev.Check("all old entries and omitted defaults unchanged", reflect.DeepEqual(cat, before.Catalog) && catalogHash(t, cat) == before.Hash, "historical content changed")
	raw, err = os.ReadFile("../live/testdata/decisions/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	ev.Check("independent source pin preserved", hex.EncodeToString(sum[:]) == "61be0c194e7231aa5c3b986c849c2b23935ae6a1830038413530eb0d3618b8fe", "source changed")
	var snapshot map[string]json.RawMessage
	mustUnmarshal(t, raw, &snapshot)
	ev.Record("sources", snapshot)
}
