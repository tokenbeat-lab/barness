package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
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

func TestKimiCandidateIdentityAndSnapshot(t *testing.T) {
	ev := run.Case(t, "KIMI-K01-K11-K14-independent-candidate-snapshots")
	beforeBuiltin, _ := ai.BuiltinCatalog().Hash()
	cat := kimi.CandidateCatalog()
	ev.Record("candidate-catalog", cat)
	ev.Record("native-prices", kimi.Prices())
	ev.Record("source-snapshot-hash", kimi.SnapshotHash())
	ev.Check("six own candidate identities", len(cat.Models) == 6, "models=%d", len(cat.Models))
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	active := w.client.Catalog()
	hash, err := active.Hash()
	ev.Check("public Client accepts source-backed candidates", err == nil && hash != "", "err=%v", err)
	for _, provider := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			m, ok := active.Lookup(provider, api, "kimi-k3")
			expected := ai.CostRates{Input: 3, Output: 15, CacheRead: .3, CacheWrite: 3}
			if provider == ai.ProviderMoonshotCN {
				expected = ai.CostRates{Input: 2.9753049688, Output: 14.8765248438, CacheRead: .2975304969, CacheWrite: 2.9753049688}
			}
			ev.Check("full identity retains own USD snapshot", ok && m.Cost.CostRates == expected && m.DefaultThinkingLevel == ai.ThinkingMax && m.MaxTokens == 131072 && m.ContextWindow == 1048576, "provider=%s api=%s model=%+v", provider, api, m)
		}
	}
	afterBuiltin, _ := ai.BuiltinCatalog().Hash()
	ev.Check("candidate loading does not change formal activation", beforeBuiltin == afterBuiltin, "before=%s after=%s", beforeBuiltin, afterBuiltin)
	cat.Models[0].Input[0] = ai.ModalityImage
	cat.Models[0].Cost.Input = 999
	view := w.client.Catalog()
	view.Models[0].Input[0] = ai.ModalityImage
	view.Models[0].Cost.Input = 777
	after, _ := w.client.Catalog().Hash()
	ev.Check("caller and returned view cannot mutate fixed snapshot", after == hash, "before=%s after=%s", hash, after)
}

func TestKimiMessagesCatalogSamplingRejected(t *testing.T) {
	for _, profile := range []bool{false, true} {
		t.Run(fmt.Sprint(profile), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("KIMI-INTL-MESSAGES-K01-unsupported-model-sampling-%t", profile))
			model := kimiFoundationModel()
			model.API = ai.APIAnthropicMessages
			if profile {
				model.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingMax: {"temperature": json.RawMessage(`1`)}}
			} else {
				model.SamplingParams = map[string]json.RawMessage{"temperature": json.RawMessage(`1`)}
			}
			cat := ai.Catalog{Version: "kimi-messages-unsupported-sampling", Models: []ai.Model{model}}
			w := newWorld(t, tenantA)
			_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
			ev.Record("catalog", cat)
			ev.Check("Messages model cannot silently grant unsupported sampling", err != nil, "err=%v", err)
		})
	}
}
