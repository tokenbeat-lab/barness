package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
)

// Modality is an input kind a model accepts.
type Modality string

const (
	ModalityText  Modality = "text"
	ModalityImage Modality = "image"
)

// Model is secret-free, versioned model metadata located by (Provider, API, ID).
type Model struct {
	Provider      ProviderID `json:"provider"`
	API           API        `json:"api"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Reasoning     bool       `json:"reasoning"`
	Input         []Modality `json:"input"`
	ContextWindow int        `json:"contextWindow"`
	MaxTokens     int        `json:"maxTokens"`
	// ThinkingLevelMap maps a reasoning model's levels to provider values
	// and marks unsupported ones.
	ThinkingLevelMap ThinkingLevelMap `json:"thinkingLevelMap,omitzero"`
	// SamplingParams are the model's default sampling parameters; the simple
	// entry merges a call's over them key by key.
	SamplingParams map[string]json.RawMessage `json:"samplingParams,omitempty"`
	Compat         ModelCompat                `json:"compat,omitzero"`
	// Cost is the model's price, which estimates every call's Usage.Cost.
	Cost ModelCost `json:"cost"`
}

// ModelCompat holds the protocol compatibility flags of pi-ai's model data
// that barness-ai implements. An absent flag has pi's default.
type ModelCompat struct {
	// SupportsStrictMode: Responses function tools carry a "strict" field.
	SupportsStrictMode bool `json:"supportsStrictMode,omitempty"`
	// SupportsMidConvoSystemMessages: later system messages are sent in place
	// as instruction updates instead of folding into the leading one. No
	// built-in model sets it yet: pi sets it only on models that also send
	// added tools as additional_tools items, which barness-ai does not
	// implement, so those models are not listed.
	SupportsMidConvoSystemMessages bool `json:"supportsMidConvoSystemMessages,omitempty"`
	// SupportsLongCacheRetention: long cache retention is requested
	// (Responses: a 24h retention; Anthropic: a 1h cache_control ttl). Unset
	// or null means true, as in pi.
	SupportsLongCacheRetention Nullable[bool] `json:"supportsLongCacheRetention,omitzero"`
	// SupportsExplicitPromptCacheMode: cache retention is sent as
	// prompt_cache_options instead of prompt_cache_retention.
	SupportsExplicitPromptCacheMode bool `json:"supportsExplicitPromptCacheMode,omitempty"`
	// ForceAdaptiveThinking: Anthropic thinking is adaptive, steered by an
	// effort level instead of a token budget.
	ForceAdaptiveThinking bool `json:"forceAdaptiveThinking,omitempty"`
	// SupportsTemperature: Anthropic accepts a temperature. Unset or null
	// means true, as in pi.
	SupportsTemperature Nullable[bool] `json:"supportsTemperature,omitzero"`
}

// acceptsImages reports whether the model takes image input; any other model
// gets pi-ai's placeholder text instead of images.
func (m Model) acceptsImages() bool { return slices.Contains(m.Input, ModalityImage) }

// Catalog is a versioned model catalog and price snapshot. It is never
// updated online. Version names a content: a catalog whose models or prices
// change gets a new Version, and Hash tells two contents apart even when a
// host reuses one.
type Catalog struct {
	Version string  `json:"version"`
	Models  []Model `json:"models"`
}

// BuiltinCatalog returns a fresh copy of the built-in catalog.
//
// Source: the frozen pi-ai 0.87.1 model data (providers/data/openai.json,
// sha256 3c52c858…2835, and providers/data/anthropic.json, sha256
// 474a010c…4e75) from the differential oracle's rebuilt copy, see
// internal/testkit/pioracle/node/PROVENANCE.md. Every listed field was
// re-checked against it when the oracle landed (compat when tools landed,
// gpt-4 when image placeholders landed, the reasoning models and their level
// maps when reasoning mapping landed, prices and gpt-5.5-pro when cost
// estimation landed, the Anthropic models when the Messages adapter landed).
// Models whose compat needs behavior barness-ai does not implement are not
// listed, so nothing here promises behavior the adapter does not have:
// OpenAI models with additional_tools or tool search, and Anthropic models
// with native mid-conversation tool changes, managed mid-conversation
// effort or server-side fallback models (claude-fable-5, claude-fable-5-1,
// claude-opus-4-8, claude-opus-5, claude-opus-5-5). pi's
// supportsOpenAIGrammarTools and Anthropic supportsStrictTools only affect
// constrained-sampling tools, which cannot be declared here, so they are not
// carried. gpt-4 and o3-mini are the text-only models: images reach them as
// placeholders; every Anthropic model takes images.
func BuiltinCatalog() Catalog {
	return Catalog{
		Version: "2026-10-02.2",
		Models:  append(builtinOpenAIModels(), builtinAnthropicModels()...),
	}
}

func builtinOpenAIModels() []Model {
	textOnly := func() []Modality { return []Modality{ModalityText} }
	textAndImage := func() []Modality { return []Modality{ModalityText, ModalityImage} }
	strict := ModelCompat{SupportsStrictMode: true}
	model := func(id, name string, input []Modality, contextWindow, maxTokens int, cost ModelCost) Model {
		return Model{Provider: ProviderOpenAI, API: APIOpenAIResponses, ID: id, Name: name, Input: input,
			ContextWindow: contextWindow, MaxTokens: maxTokens, Compat: strict, Cost: cost}
	}
	// price is pi's cost entry without tiers; OpenAI's data prices no
	// cache writes.
	price := func(input, output, cacheRead float64) ModelCost {
		return ModelCost{CostRates: CostRates{Input: input, Output: output, CacheRead: cacheRead}}
	}
	reasoning := func(m Model, levels ThinkingLevelMap) Model {
		m.Reasoning, m.ThinkingLevelMap = true, levels
		return m
	}
	return []Model{
		model("gpt-4", "GPT-4", textOnly(), 8192, 8192, price(30, 60, 0)),
		model("gpt-4.1", "GPT-4.1", textAndImage(), 1047576, 32768, price(2, 8, 0.5)),
		model("gpt-4.1-mini", "GPT-4.1 mini", textAndImage(), 1047576, 32768, price(0.4, 1.6, 0.1)),
		model("gpt-4o-mini", "GPT-4o mini", textAndImage(), 128000, 16384, price(0.15, 0.6, 0.075)),
		reasoning(model("gpt-5", "GPT-5", textAndImage(), 400000, 128000, price(1.25, 10, 0.125)), levelMap("", ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh)),
		reasoning(model("gpt-5-mini", "GPT-5 Mini", textAndImage(), 400000, 128000, price(0.25, 2, 0.025)), levelMap("", ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh)),
		reasoning(model("gpt-5-nano", "GPT-5 Nano", textAndImage(), 400000, 128000, price(0.05, 0.4, 0.005)), levelMap("", ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh)),
		reasoning(model("gpt-5-pro", "GPT-5 Pro", textAndImage(), 400000, 128000, price(15, 120, 0)), levelMap("", ThinkingHigh)),
		reasoning(model("gpt-5.1", "GPT-5.1", textAndImage(), 400000, 128000, price(1.25, 10, 0.125)), levelMap("none", ThinkingLow, ThinkingMedium, ThinkingHigh)),
		reasoning(model("gpt-5.2", "GPT-5.2", textAndImage(), 400000, 128000, price(1.75, 14, 0.175)), levelMap("none", ThinkingLow, ThinkingMedium, ThinkingHigh, ThinkingXHigh)),
		// The one listed model with tiered prices: the others pi tiers
		// need additional_tools or tool search.
		reasoning(model("gpt-5.5-pro", "GPT-5.5 Pro", textAndImage(), 1050000, 128000,
			ModelCost{CostRates: CostRates{Input: 30, Output: 180},
				Tiers: []CostTier{{InputTokensAbove: 272000, CostRates: CostRates{Input: 60, Output: 270}}}}),
			levelMap("", ThinkingMedium, ThinkingHigh, ThinkingXHigh)),
		reasoning(model("o3", "o3", textAndImage(), 200000, 100000, price(2, 8, 0.5)), levelMap("", ThinkingLow, ThinkingMedium, ThinkingHigh)),
		reasoning(model("o3-mini", "o3-mini", textOnly(), 200000, 100000, price(1.1, 4.4, 0.55)), levelMap("", ThinkingLow, ThinkingMedium, ThinkingHigh)),
		reasoning(model("o4-mini", "o4-mini", textAndImage(), 200000, 100000, price(1.1, 4.4, 0.275)), levelMap("", ThinkingLow, ThinkingMedium, ThinkingHigh)),
	}
}

// builtinAnthropicModels are pi's Anthropic Messages models that barness-ai
// can serve. All are reasoning models taking text and images. A level map
// lists only the entries pi's data sets; the others stay unset.
func builtinAnthropicModels() []Model {
	model := func(id, name string, contextWindow, maxTokens int, input, output, cacheRead, cacheWrite float64) Model {
		return Model{Provider: ProviderAnthropic, API: APIAnthropicMessages, ID: id, Name: name, Reasoning: true,
			Input: []Modality{ModalityText, ModalityImage}, ContextWindow: contextWindow, MaxTokens: maxTokens,
			Cost: ModelCost{CostRates: CostRates{Input: input, Output: output, CacheRead: cacheRead, CacheWrite: cacheWrite}}}
	}
	// adaptive marks a model with adaptive thinking and its level map.
	adaptive := func(m Model, levels ThinkingLevelMap) Model {
		m.Compat.ForceAdaptiveThinking, m.ThinkingLevelMap = true, levels
		return m
	}
	maxOnly := func() ThinkingLevelMap { return ThinkingLevelMap{Max: Value("max")} }
	xhighAndMax := func() ThinkingLevelMap { return ThinkingLevelMap{XHigh: Value("xhigh"), Max: Value("max")} }
	opus47 := adaptive(model("claude-opus-4-7", "Claude Opus 4.7", 1000000, 128000, 5, 25, 0.5, 6.25), xhighAndMax())
	opus47.Compat.SupportsTemperature = Value(false)
	return []Model{
		model("claude-haiku-4-5", "Claude Haiku 4.5 (latest)", 200000, 64000, 1, 5, 0.1, 1.25),
		model("claude-haiku-4-5-20251001", "Claude Haiku 4.5", 200000, 64000, 1, 5, 0.1, 1.25),
		model("claude-opus-4-5", "Claude Opus 4.5 (latest)", 200000, 64000, 5, 25, 0.5, 6.25),
		model("claude-opus-4-5-20251101", "Claude Opus 4.5", 200000, 64000, 5, 25, 0.5, 6.25),
		adaptive(model("claude-opus-4-6", "Claude Opus 4.6", 1000000, 128000, 5, 25, 0.5, 6.25), maxOnly()),
		opus47,
		model("claude-sonnet-4-5", "Claude Sonnet 4.5 (latest)", 1000000, 64000, 3, 15, 0.3, 3.75),
		model("claude-sonnet-4-5-20250929", "Claude Sonnet 4.5", 1000000, 64000, 3, 15, 0.3, 3.75),
		adaptive(model("claude-sonnet-4-6", "Claude Sonnet 4.6", 1000000, 128000, 3, 15, 0.3, 3.75), maxOnly()),
		adaptive(model("claude-sonnet-5", "Claude Sonnet 5", 1000000, 128000, 2, 10, 0.2, 2.5), xhighAndMax()),
	}
}

// levelMap is a complete level map as pi's OpenAI data writes it: off maps
// to off ("" for null), each supported level to itself, every other level
// to null.
func levelMap(off string, supported ...ThinkingLevel) ThinkingLevelMap {
	entry := func(l ThinkingLevel) Nullable[string] {
		if slices.Contains(supported, l) {
			return Value(string(l))
		}
		return Null[string]()
	}
	m := ThinkingLevelMap{Off: Null[string](), Minimal: entry(ThinkingMinimal), Low: entry(ThinkingLow), Medium: entry(ThinkingMedium),
		High: entry(ThinkingHigh), XHigh: entry(ThinkingXHigh), Max: entry(ThinkingMax)}
	if off != "" {
		m.Off = Value(off)
	}
	return m
}

func (c Catalog) clone() Catalog {
	out := Catalog{Version: c.Version, Models: make([]Model, len(c.Models))}
	for i, m := range c.Models {
		m.Input = append([]Modality(nil), m.Input...)
		m.SamplingParams = cloneJSONValues(m.SamplingParams)
		m.Cost.Tiers = slices.Clone(m.Cost.Tiers)
		out.Models[i] = m
	}
	return out
}

// Hash is "sha256:" and the hex SHA-256 of c's JSON encoding, which is
// deterministic. CallMetadata.CatalogHash carries it. It fails for a
// catalog NewClient would refuse: a non-finite price cannot be encoded.
func (c Catalog) Hash() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// validate refuses a catalog whose prices could not yield a finite,
// non-negative cost.
func (c Catalog) validate() error {
	for _, m := range c.Models {
		if problem := m.Cost.problem(); problem != "" {
			return &ConfigError{Field: "Catalog", Problem: "model " + m.ID + ": " + problem}
		}
	}
	return nil
}

func (c Catalog) lookup(provider ProviderID, api API, id string) (Model, bool) {
	for _, m := range c.Models {
		if m.Provider == provider && m.API == api && m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}
