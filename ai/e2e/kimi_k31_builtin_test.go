package e2e

import (
	"bytes"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// The original approval stays independent of later catalog additions. Its
// surviving identities must retain definitions/order, and the adopted current
// bytes must still match their review baseline.
func TestKimiK31BuiltinAdoption(t *testing.T) {
	ev := run.Case(t, "KIMI-K31-builtin-adoption")
	beforeRaw, err := os.ReadFile("testdata/catalog/pre-kimi-k31-builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	approved, err := os.ReadFile("testdata/catalog/kimi-k31-approved.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("previous-builtin.json", beforeRaw)
	ev.Fixture("approved-snapshot.json", approved)
	var before, expected struct {
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}
	mustUnmarshal(t, beforeRaw, &before)
	mustUnmarshal(t, approved, &expected)
	current := ai.BuiltinCatalog()
	ev.Check("all 119 previous identities preserve definitions and relative order", reflect.DeepEqual(historicalCatalogInCurrentOrder(maintainedFixtureCatalog(t), before.Catalog), before.Catalog), "previous definitions changed")
	historical := historicalCatalogInCurrentOrder(catalogWithHistoricalNames(current, expected.Catalog), expected.Catalog)
	ev.Check("surviving approved identities retain definitions and order", reflect.DeepEqual(historicalCatalogInCurrentOrder(expected.Catalog, historical), historical), "approved historical definitions differ")
	builtin, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := os.ReadFile("../../catalog/data/baseline/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("builtin and adopted baseline preserve exact approved snapshot bytes", bytes.Equal(builtin, baseline), "snapshot bytes differ")
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			t.Run(string(market)+"/"+string(api), func(t *testing.T) {
				id := catalogContractIdentity{Operation: ai.OperationChat, Provider: market, API: api, ID: "kimi-k3.1-pro"}
				e := run.Case(t, "KIMI-K31-builtin-"+string(market)+"-"+string(api))
				m, ok := current.Lookup(market, api, id.ID)
				if !e.Check("public builtin discovers the exact new identity", ok, "identity=%+v", id) {
					return
				}
				rates := ai.CostRates{Input: 2, Output: 10, CacheRead: .2, CacheWrite: 2}
				if market == ai.ProviderMoonshotCN {
					rates = ai.CostRates{Input: 1.9360795878, Output: 10.1271855360, CacheRead: .1936079588, CacheWrite: 1.9360795878}
				}
				low, lok := m.ThinkingLevelMap.Low.Get()
				high, hok := m.ThinkingLevelMap.High.Get()
				max, mok := m.ThinkingLevelMap.Max.Get()
				e.Check("approved constrained model and market rates", m.Reasoning && m.Compat.RequiresThinking && m.ContextWindow == 1048576 && m.MaxTokens == 131072 && slices.Equal(m.Input, []ai.Modality{ai.ModalityText}) && m.DefaultThinkingLevel == ai.ThinkingMax && m.ThinkingLevelMap.Off.IsNull() && m.ThinkingLevelMap.Minimal.IsNull() && m.ThinkingLevelMap.Medium.IsNull() && m.ThinkingLevelMap.XHigh.IsNull() && lok && low == "low" && hok && high == "high" && mok && max == "max" && m.Cost.CostRates == rates && len(m.Cost.Tiers) == 0, "model=%+v", m)
				for _, scene := range []string{"basic", "cache-usage", "history-tools", "input-modalities", "parameters", "reasoning", "authorization"} {
					t.Run(scene, func(t *testing.T) {
						catalogChatContract(t, run.Case(t, "KIMI-K31-builtin-"+string(market)+"-"+string(api)+"-"+scene), current, id, scene)
					})
				}
			})
		}
	}
}
