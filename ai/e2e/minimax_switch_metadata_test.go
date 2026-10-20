package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

// A model's binary switch and mandatory-thinking constraint are independent.
// Catalog JSON, Hash and Client snapshots must preserve that distinction.
func TestMiniMaxThinkingSwitchOnlyMetadata(t *testing.T) {
	ev := run.Case(t, "MM04-switch-metadata")
	cat := minimax.CandidateCatalog()
	for _, m := range cat.Models {
		fields := bodyFields(t, mustMarshal(t, m.Compat))
		want := m.ID == "MiniMax-M3"
		ev.Check("each protocol declares its own switch semantics "+string(m.API)+"/"+m.ID, (string(fields["thinkingSwitchOnly"]) == "true") == want, "%s", mustMarshal(t, m.Compat))
	}
	base, _ := ai.BuiltinCatalog().Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-5")
	base.Compat = ai.ModelCompat{}
	declared := patchModel(base, json.RawMessage(`{"compat":{"thinkingSwitchOnly":true}}`))
	owned := ai.Catalog{Version: "synthetic-switch-metadata", Models: []ai.Model{declared}}
	h := host.New()
	client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &owned})
	if !ev.Check("valid explicit metadata accepted", err == nil, "%v", err) {
		return
	}
	before := mustMarshal(t, client.Catalog())
	hash := catalogHash(t, owned)
	owned.Models[0] = base
	snapshot := client.Catalog()
	snapshot.Models[0] = base
	ev.Check("switch metadata owns construction and output snapshots", jsonEqual(before, mustMarshal(t, client.Catalog())), "%s", before)
	ev.Check("explicit switch participates in public catalog hash", catalogHash(t, owned) != hash, "hash unchanged")
	fields := bodyFields(t, mustMarshal(t, client.Catalog().Models[0].Compat))
	ev.Check("JSON retains switch metadata", string(fields["thinkingSwitchOnly"]) == "true", "%s", before)
	ev.Record("catalog", client.Catalog())
	ev.Record("hash", hash)
}
