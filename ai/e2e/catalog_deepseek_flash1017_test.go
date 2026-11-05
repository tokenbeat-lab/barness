package e2e

import (
	"os"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Failure inventory: stale release label, invented dated wire ID, duplicate
// identities, changed capability/price/order, or protocol/image/history/auth
// regressions. The upstream release uses the existing rolling deepseek-flash ID.
func TestDeepSeekFlash1017Builtin(t *testing.T) {
	ev := run.Case(t, "CATALOG-DEEPSEEK-FLASH1017-builtin")
	cat := ai.BuiltinCatalog()
	raw, err := os.ReadFile("testdata/catalog/pre-deepseek-flash1017.json")
	if err != nil {
		t.Fatal(err)
	}
	var before struct {
		Hash    string
		Catalog ai.Catalog
	}
	mustUnmarshal(t, raw, &before)
	ev.Check("pre-revision snapshot stays valid", catalogHash(t, before.Catalog) == before.Hash, "invalid snapshot")
	for _, api := range []ai.API{ai.APIOpenAIResponses, ai.APIOpenAICompletions} {
		t.Run(string(api), func(t *testing.T) {
			i := catalogContractIdentity{Operation: ai.OperationChat, Provider: ai.ProviderDeepSeek, API: api, ID: "deepseek-flash"}
			m, ok := cat.Lookup(i.Provider, i.API, i.ID)
			if !ev.Check("current served release is discoverable "+string(api), ok && m.Name == "DeepSeek V4.1 Flash 1017", "model=%+v", m) {
				return
			}
			old, existed := before.Catalog.Lookup(i.Provider, i.API, i.ID)
			m.Name = old.Name
			ev.Check("capabilities and prices stay unchanged "+string(api), existed && reflect.DeepEqual(m, old), "definition changed beyond label")
			for _, scene := range []string{"load-query", "basic", "cache-usage", "history-tools", "input-modalities", "parameters", "reasoning", "authorization"} {
				t.Run(scene, func(t *testing.T) {
					catalogChatContract(t, run.Case(t, "CATALOG-DEEPSEEK-FLASH1017-"+string(api)+"-"+scene), cat, i, scene)
				})
			}
		})
	}
}
