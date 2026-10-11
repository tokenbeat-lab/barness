package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// S21-S23 fail when any public snapshot shares the outer/inner map or raw
// bytes. Mutation is sequential, after construction/read; hosts never write
// a Go map while the constructor is reading it.
func TestThinkingSamplingSnapshots(t *testing.T) {
	ev := run.Case(t, "E03-thinking-sampling-snapshots")
	catalog := ai.BuiltinCatalog()
	model := &catalog.Models[0]
	model.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{
		ai.ThinkingLevel("off"): {"temperature": json.RawMessage(`0.2`), "seed": json.RawMessage(`false`)},
		ai.ThinkingHigh:         {"temperature": json.RawMessage(`0.8`), "seed": json.RawMessage(`null`)},
	}
	before := mustMarshal(t, catalog)
	hash := catalogHash(t, catalog)
	ev.Record("catalog-at-construction", catalog)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &catalog }, tenantA)
	mutate := func(m *ai.Model) {
		m.SamplingParamsByThinkingLevel[ai.ThinkingHigh]["temperature"][2] = '9'
		m.SamplingParamsByThinkingLevel[ai.ThinkingLevel("off")]["top_p"] = json.RawMessage(`0`)
		m.SamplingParamsByThinkingLevel[ai.ThinkingMedium] = map[string]json.RawMessage{"top_k": json.RawMessage(`1`)}
	}
	mutate(model)
	snapshot := w.client.Catalog()
	ev.Check("construction owns every layer", jsonEqual(before, mustMarshal(t, snapshot)), "snapshot differs")
	// Each outgoing lookup/list/catalog value must own all three layers too.
	for _, source := range []string{"lookup", "list", "client-catalog"} {
		snapshot = w.client.Catalog()
		var got ai.Model
		switch source {
		case "lookup":
			got, _ = snapshot.Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4")
		case "list":
			got = snapshot.ModelsOf(ai.ProviderOpenAI, ai.APIOpenAIResponses)[0]
		default:
			got = snapshot.Models[0]
		}
		mutate(&got)
		if source != "client-catalog" {
			ev.Check(source+" owns every layer", jsonEqual(before, mustMarshal(t, snapshot)), "catalog differs")
		}
		ev.Check(source+" preserves client", jsonEqual(before, mustMarshal(t, w.client.Catalog())), "client differs")
	}
	original := w.client.Catalog()
	ev.Check("snapshot hash preserved", catalogHash(t, original) == hash, "hash differs")
	// Reversing insertion at both levels keeps identical content.
	reversed := ai.BuiltinCatalog()
	reversed.Models[0].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{}
	reversed.Models[0].SamplingParamsByThinkingLevel[ai.ThinkingHigh] = map[string]json.RawMessage{"seed": json.RawMessage(`null`), "temperature": json.RawMessage(`0.8`)}
	reversed.Models[0].SamplingParamsByThinkingLevel[ai.ThinkingLevel("off")] = map[string]json.RawMessage{"seed": json.RawMessage(`false`), "temperature": json.RawMessage(`0.2`)}
	ev.Check("insertion order does not change hash", catalogHash(t, reversed) == hash, "hash differs")
	reversed.Models[0].SamplingParamsByThinkingLevel[ai.ThinkingHigh]["temperature"] = json.RawMessage(`0.3`)
	ev.Check("profile value changes hash", catalogHash(t, reversed) != hash, "hash unchanged")
	ev.Record("original-client-catalog", w.client.Catalog())
}
