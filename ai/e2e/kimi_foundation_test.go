package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// This literal is the agreed K3 supported subset, independent from the
// candidate source loader. Host metadata can narrow these capacities.
func kimiFoundationModel() ai.Model {
	return ai.Model{Provider: "moonshot", API: ai.APIOpenAICompletions, ID: "kimi-k3", Name: "K3 controlled", Reasoning: true,
		Input: []ai.Modality{ai.ModalityText}, ContextWindow: 1048576, MaxTokens: 131072,
		ThinkingLevelMap: ai.ThinkingLevelMap{Off: ai.Null[string](), Minimal: ai.Null[string](), Low: ai.Value("low"), Medium: ai.Null[string](), High: ai.Value("high"), XHigh: ai.Null[string](), Max: ai.Value("max")}, DefaultThinkingLevel: ai.ThinkingMax,
		Compat: ai.ModelCompat{RequiresThinking: true}, Cost: ai.ModelCost{CostRates: ai.CostRates{Input: 3, Output: 15, CacheRead: .3, CacheWrite: 3}}}
}

func TestKimiCatalogBoundary(t *testing.T) {
	for _, kind := range []string{"model", "protocol", "image", "context", "output", "zero-price", "tiers", "effort", "off", "strict", "adaptive", "budget", "default-low", "default-empty", "max-disabled", "thinking-optional"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "KIMI-INTL-CHAT-K01-catalog-reject-"+kind)
			m := kimiFoundationModel()
			switch kind {
			case "default-low":
				m.DefaultThinkingLevel = ai.ThinkingLow
			case "default-empty":
				m.DefaultThinkingLevel = ""
			case "max-disabled":
				m.DefaultThinkingLevel = ""
				m.ThinkingLevelMap.Max = ai.Null[string]()
			case "thinking-optional":
				m.Compat.RequiresThinking = false
			case "model":
				m.ID = "kimi-k2.5"
			case "protocol":
				m.API = ai.APIGoogleGenerativeAI
			case "image":
				m.Input = append(m.Input, ai.ModalityImage)
			case "context":
				m.ContextWindow++
			case "output":
				m.MaxTokens++
			case "zero-price":
				m.Cost.CacheWrite = 0
			case "tiers":
				m.Cost.Tiers = []ai.CostTier{{InputTokensAbove: 1, CostRates: m.Cost.CostRates}}
			case "effort":
				m.ThinkingLevelMap.Low = ai.Value("max")
			case "off":
				m.ThinkingLevelMap.Off = ai.Value("none")
			case "strict":
				m.Compat.SupportsStrictMode = true
			case "adaptive":
				m.Compat.ForceAdaptiveThinking = true
			case "budget":
				m.Compat.MessagesThinkingBudgetTokens = 1024
			}
			cat := ai.Catalog{Version: "kimi-controlled-reject", Models: []ai.Model{m}}
			w := newWorld(t, tenantA)
			_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
			ev.Record("catalog", cat)
			ev.Check("public Client rejects metadata beyond implemented K3 subset", err != nil, "kind=%s err=%v", kind, err)
		})
	}
}
