package e2e

import (
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Failure modes: the six approved identities are refused; a request is routed
// under the wrong model or authority; text/usage/cost or low/high/max effort is
// lost; or the new identity expands the existing Kimi modality, capacity,
// reasoning, protocol, price, or compatibility boundaries. Each market uses
// its separately confirmed standard USD price; China is a fixed conversion.
func TestKimiK31PublicCatalogContract(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			t.Run(string(market)+"/"+string(api), func(t *testing.T) {
				id := catalogContractIdentity{Operation: ai.OperationChat, Provider: market, API: api, ID: "kimi-k3.1-pro"}
				ev := run.Case(t, "KIMI-K31-accept-"+string(market)+"-"+string(api))
				m, found := ai.BuiltinCatalog().Lookup(market, api, "kimi-k3")
				if !ev.Check("existing K3 gives the implemented protocol subset", found, "provider=%s api=%s", market, api) {
					return
				}
				m.ID, m.Name = id.ID, "K3.1 Pro controlled"
				m.Cost.CostRates = ai.CostRates{Input: 2, Output: 10, CacheRead: .2, CacheWrite: 2}
				if market == ai.ProviderMoonshotCN {
					m.Cost.CostRates = ai.CostRates{Input: 1.9360795878, Output: 10.1271855360, CacheRead: .1936079588, CacheWrite: 1.9360795878}
				}
				cat := ai.Catalog{Version: "kimi-k31-controlled", Models: []ai.Model{m}}
				w := newWorld(t, tenantA)
				client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
				ev.Record("catalog", cat)
				if !ev.Check("public Client accepts approved K3.1 identity", err == nil, "provider=%s api=%s err=%v", market, api, err) {
					return
				}
				got, ok := client.Catalog().Lookup(market, api, id.ID)
				ev.Check("exact identity and confirmed market price remain discoverable", ok && got.Cost.CostRates == m.Cost.CostRates, "model=%+v", got)
				for _, scene := range []string{"load-query", "basic", "cache-usage", "history-tools", "input-modalities", "parameters", "reasoning", "authorization"} {
					t.Run(scene, func(t *testing.T) {
						catalogChatContract(t, run.Case(t, "KIMI-K31-"+string(market)+"-"+string(api)+"-"+scene), cat, id, scene)
					})
				}
			})
		}
	}
}

func TestKimiK31CatalogRejectsExpandedCapabilities(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			for _, kind := range []string{"empty-id", "protocol", "image", "context", "output", "zero-price", "tiers", "off", "effort", "default", "thinking-optional", "strict", "adaptive", "budget"} {
				t.Run(fmt.Sprintf("%s/%s/%s", market, api, kind), func(t *testing.T) {
					ev := run.Case(t, "KIMI-K31-reject-"+string(market)+"-"+string(api)+"-"+kind)
					m, ok := ai.BuiltinCatalog().Lookup(market, api, "kimi-k3")
					if !ok {
						t.Fatal("missing existing K3 identity")
					}
					m.ID = "kimi-k3.1-pro"
					switch kind {
					case "empty-id":
						m.ID = ""
					case "protocol":
						m.API = ai.APIGoogleGenerativeAI
					case "image":
						m.Input = append(m.Input, ai.ModalityImage)
					case "context":
						m.ContextWindow = 1048577
					case "output":
						m.MaxTokens = 131073
					case "zero-price":
						m.Cost.CacheWrite = 0
					case "tiers":
						m.Cost.Tiers = []ai.CostTier{{InputTokensAbove: 1, CostRates: m.Cost.CostRates}}
					case "off":
						m.ThinkingLevelMap.Off = ai.Value("none")
					case "effort":
						m.ThinkingLevelMap.Low = ai.Value("max")
					case "default":
						m.DefaultThinkingLevel = ai.ThinkingLow
					case "thinking-optional":
						m.Compat.RequiresThinking = false
					case "strict":
						m.Compat.SupportsStrictMode = true
					case "adaptive":
						m.Compat.ForceAdaptiveThinking = true
					case "budget":
						m.Compat.MessagesThinkingBudgetTokens = 1024
					}
					cat := ai.Catalog{Version: "kimi-k31-reject", Models: []ai.Model{m}}
					w := newWorld(t, tenantA)
					_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
					ev.Check("public Client rejects unapproved expansion", err != nil, "kind=%s err=%v", kind, err)
				})
			}
		}
	}
}
