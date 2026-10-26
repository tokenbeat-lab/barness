package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"reflect"
	"testing"
)

// The agreed candidate contract is independent from its source loader.
func xaiFoundationModel() ai.Model {
	return ai.Model{Provider: ai.ProviderXAI, API: ai.APIOpenAIResponses, ID: "grok-4.7", Name: "Grok 4.7", Reasoning: true,
		Input: []ai.Modality{ai.ModalityText, ai.ModalityImage}, ContextWindow: 500000, MaxTokens: 128000,
		ThinkingLevelMap: ai.ThinkingLevelMap{Off: ai.Null[string](), Minimal: ai.Null[string](), Low: ai.Value("low"), Medium: ai.Value("medium"), High: ai.Value("high"), XHigh: ai.Value("xhigh"), Max: ai.Null[string]()}, DefaultThinkingLevel: ai.ThinkingHigh,
		Compat: ai.ModelCompat{RequiresThinking: true, SupportsMidConvoSystemMessages: true, SupportsLongCacheRetention: ai.Value(false)},
		Cost:   ai.ModelCost{CostRates: ai.CostRates{Input: 2, CacheRead: .5, Output: 6}, Tiers: []ai.CostTier{{InputTokensAbove: 199999, CostRates: ai.CostRates{Input: 4, CacheRead: 1, Output: 12}}}}}
}

func TestXAICatalogBoundary(t *testing.T) {
	for _, kind := range []string{"model", "protocol", "context", "output", "zero-context", "zero-output", "no-text", "unknown-input", "reasoning", "required-thinking", "default-low", "default-empty", "high-disabled", "off", "unset-off", "minimal", "max", "effort", "strict", "adaptive", "budget", "explicit-cache", "mid-effort", "switch-only"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "XAI-E01-catalog-reject-"+kind)
			m := xaiFoundationModel()
			switch kind {
			case "model":
				m.ID = "grok-4.7-fast"
			case "protocol":
				m.API = ai.APIAnthropicMessages
			case "context":
				m.ContextWindow++
			case "output":
				m.MaxTokens++
			case "zero-context":
				m.ContextWindow = 0
			case "zero-output":
				m.MaxTokens = 0
			case "no-text":
				m.Input = []ai.Modality{ai.ModalityImage}
			case "unknown-input":
				m.Input = append(m.Input, ai.Modality("audio"))
			case "reasoning":
				m.Reasoning = false
			case "required-thinking":
				m.Compat.RequiresThinking = false
			case "default-low":
				m.DefaultThinkingLevel = ai.ThinkingLow
			case "default-empty":
				m.DefaultThinkingLevel = ""
			case "high-disabled":
				m.ThinkingLevelMap.High = ai.Null[string]()
			case "off":
				m.ThinkingLevelMap.Off = ai.Value("none")
			case "unset-off":
				m.ThinkingLevelMap.Off = ai.Nullable[string]{}
			case "minimal":
				m.ThinkingLevelMap.Minimal = ai.Value("low")
			case "max":
				m.ThinkingLevelMap.Max = ai.Value("xhigh")
			case "effort":
				m.ThinkingLevelMap.Low = ai.Value("high")
			case "strict":
				m.Compat.SupportsStrictMode = true
			case "adaptive":
				m.Compat.ForceAdaptiveThinking = true
			case "budget":
				m.Compat.MessagesThinkingBudgetTokens = 1024
			case "explicit-cache":
				m.Compat.SupportsExplicitPromptCacheMode = true
			case "mid-effort":
				m.Compat.SupportsMidConvoEffort = true
			case "switch-only":
				m.Compat.ThinkingSwitchOnly = true
			}
			cat := ai.Catalog{Version: "xai-controlled-invalid", Models: []ai.Model{m}}
			_, hashErr := cat.Hash()
			w := newWorld(t, tenantA)
			_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
			ev.Record("catalog", cat)
			ev.Check("Catalog and Client reject unsupported xAI metadata", hashErr != nil && err != nil, "kind=%s Hash=%v NewClient=%v", kind, hashErr, err)
		})
	}
}

func TestXAIRejectsOtherOperationCatalogs(t *testing.T) {
	for _, kind := range []string{"image", "classifier"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "XAI-E01-catalog-operation-"+kind)
			cat := ai.Catalog{Version: "xai-unimplemented-operation"}
			if kind == "image" {
				cat.ImageModels = []ai.ImageModel{{Provider: ai.ProviderXAI, API: ai.APIOpenAIImages, ID: "grok-4.7", Input: []ai.Modality{ai.ModalityText}, Output: []ai.Modality{ai.ModalityImage}, Capabilities: ai.ImageCapabilities{MaxOutputImages: 1}, Pricing: ai.ImagePricing{InputText: ai.Value(1.0), OutputImage: ai.Value(1.0)}}}
			} else {
				cat.ClassifierModels = []ai.ClassifierModel{{Provider: ai.ProviderXAI, API: ai.APITypeSafeSystemOne, ID: "grok-4.7", ContextWindow: 500000, Capabilities: ai.ClassifierCapabilities{Kinds: []ai.ClassifierQuestionKind{ai.ClassifierQuestionBool}}}}
			}
			_, err := cat.Hash()
			ev.Record("catalog", cat)
			ev.Check("other operations cannot enter xAI catalog", err != nil, "kind=%s err=%v", kind, err)
		})
	}
}

func TestXAICandidateSnapshot(t *testing.T) {
	ev := run.Case(t, "XAI-E01-independent-candidate-snapshot")
	before, _ := ai.BuiltinCatalog().Hash()
	cat := xai.CandidateCatalog()
	ev.Record("candidate-catalog", cat)
	ev.Check("one independent reviewed identity", len(cat.Models) == 1 && reflect.DeepEqual(cat.Models[0], xaiFoundationModel()), "catalog=%+v", cat)
	_, formal := ai.BuiltinCatalog().Lookup(ai.ProviderXAI, ai.APIOpenAIResponses, "grok-4.7")
	ev.Check("candidate remains distinct from the prepared unreleased Builtin snapshot", formal && cat.Version != ai.BuiltinCatalog().Version && ai.ProviderXAI == ai.ProviderID("xai"), "prepared=%t", formal)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	hash, _ := w.client.Catalog().Hash()
	cat.Models[0].Input[0] = ai.ModalityImage
	cat.Models[0].Cost.Tiers[0].Input = 999
	view := w.client.Catalog()
	view.Models[0].Input[0] = ai.ModalityImage
	view.Models[0].Cost.Tiers[0].Input = 888
	after, _ := w.client.Catalog().Hash()
	builtinAfter, _ := ai.BuiltinCatalog().Hash()
	fresh := xai.CandidateCatalog()
	ev.Check("input and returned views cannot change pinned catalog", hash == after, "hash=%s after=%s", hash, after)
	ev.Check("candidate reads remain independent and do not activate", fresh.Models[0].Cost.Tiers[0].Input == 4 && before == builtinAfter, "fresh=%+v", fresh)
}

func TestXAIHostNarrowsCandidate(t *testing.T) {
	ev := run.Case(t, "XAI-E01-host-narrowed-candidate")
	cat := xai.CandidateCatalog()
	m := &cat.Models[0]
	m.Input = []ai.Modality{ai.ModalityText}
	m.ContextWindow = 10000
	m.MaxTokens = 1024
	m.ThinkingLevelMap.Low = ai.Null[string]()
	m.ThinkingLevelMap.Medium = ai.Null[string]()
	m.ThinkingLevelMap.XHigh = ai.Null[string]()
	m.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: 5, Output: 15, CacheRead: 2}}
	w := newWorld(t, tenantA)
	c, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
	ev.Record("host-catalog", cat)
	ev.Check("Host may narrow capabilities and supply independent price snapshot", err == nil && c != nil, "err=%v", err)
}
