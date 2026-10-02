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

// supportsLongCacheRetention is the model's compat flag; unset or null is
// true, as in pi.
func (m Model) supportsLongCacheRetention() bool {
	if v, ok := m.Compat.SupportsLongCacheRetention.Get(); ok {
		return v
	}
	return true
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
// sha256 3c52c858…2835, providers/data/anthropic.json, sha256
// 474a010c…4e75, and providers/data/google.json, sha256 328822707e…708c)
// from the differential oracle's rebuilt copy, see
// internal/testkit/pioracle/node/PROVENANCE.md. Every listed field was
// re-checked against it when the oracle landed (compat when tools landed,
// gpt-4 when image placeholders landed, the reasoning models and their level
// maps when reasoning mapping landed, prices and gpt-5.5-pro when cost
// estimation landed, the Anthropic models when the Messages adapter landed,
// the Google models when the Gemini adapter landed).
// Models whose compat needs behavior barness-ai does not implement are not
// listed, so nothing here promises behavior the adapter does not have:
// OpenAI models with additional_tools or tool search, and Anthropic models
// with native mid-conversation tool changes, managed mid-conversation
// effort or server-side fallback models (claude-fable-5, claude-fable-5-1,
// claude-opus-4-8, claude-opus-5, claude-opus-5-5), and Google models that
// generateContent alone cannot serve (see builtinGoogleModels). pi's
// supportsOpenAIGrammarTools and Anthropic supportsStrictTools only affect
// constrained-sampling tools, which cannot be declared here, so they are not
// carried. gpt-4 and o3-mini are the text-only models: images reach them as
// placeholders; every Anthropic and Google model takes images.
//
// pi's data lists OpenAI's models for Responses only; the OpenAI Chat
// Completions models are the same entries on that API (see
// builtinOpenAIChatModels), as a pi user configures a custom model.
func BuiltinCatalog() Catalog {
	return Catalog{
		Version: "2026-10-02.4",
		Models: slices.Concat(builtinOpenAIModels(), builtinOpenAIChatModels(), builtinAnthropicModels(),
			builtinGoogleModels()),
	}
}

// builtinOpenAIChatModels are the listed OpenAI models that Chat Completions
// also serves, with pi's OpenAI data on the openai-completions API
// (ADR-0013): every field, compat included, as for Responses. The pro
// models answer only on Responses, so they are left out.
func builtinOpenAIChatModels() []Model {
	var out []Model
	for _, m := range builtinOpenAIModels() {
		if m.ID == "gpt-5-pro" || m.ID == "gpt-5.5-pro" {
			continue
		}
		m.API = APIOpenAICompletions
		out = append(out, m)
	}
	return out
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

// builtinGoogleModels are pi's google-generative-ai models that barness-ai
// can serve. All are reasoning models taking text and images, without
// tiered prices or cache writes. Gemini 2.5 models have no level map (they
// take a thinking budget); the others steer thinking by level. Left out are
// the models generateContent alone cannot serve (ADR-0012): the Deep
// Research agents (deep-research-preview-04-2026,
// deep-research-max-preview-04-2026), which run through the Interactions
// API; gemini-2.5-computer-use-preview-10-2025, which needs the hosted
// computer use tool; gemini-3.1-flash-live-preview, a Live API model; and
// gemini-3.1-flash-lite-image, whose image output the adapter drops.
func builtinGoogleModels() []Model {
	model := func(id, name string, contextWindow, maxTokens int, input, output, cacheRead float64) Model {
		return Model{Provider: ProviderGoogle, API: APIGoogleGenerativeAI, ID: id, Name: name, Reasoning: true,
			Input: []Modality{ModalityText, ModalityImage}, ContextWindow: contextWindow, MaxTokens: maxTokens,
			Cost: ModelCost{CostRates: CostRates{Input: input, Output: output, CacheRead: cacheRead}}}
	}
	// levels is a Gemini 3 level map as pi's data writes it: off, xhigh
	// and max null, the four Gemini levels mapped to themselves unless
	// listed as unsupported.
	levels := func(m Model, unsupported ...ThinkingLevel) Model {
		entry := func(l ThinkingLevel) Nullable[string] {
			if slices.Contains(unsupported, l) {
				return Null[string]()
			}
			return Value(string(l))
		}
		m.ThinkingLevelMap = ThinkingLevelMap{Off: Null[string](), Minimal: entry(ThinkingMinimal), Low: entry(ThinkingLow),
			Medium: entry(ThinkingMedium), High: entry(ThinkingHigh), XHigh: Null[string](), Max: Null[string]()}
		return m
	}
	// gemma is a Gemma 4 model: free, its level map pi's uppercase one
	// without xhigh and max entries.
	gemma := func(id, name string) Model {
		m := model(id, name, 262144, 32768, 0, 0, 0)
		m.ThinkingLevelMap = ThinkingLevelMap{Off: Null[string](), Minimal: Value("MINIMAL"), Low: Null[string](),
			Medium: Null[string](), High: Value("HIGH")}
		return m
	}
	const window, output = 1048576, 65536
	return []Model{
		model("gemini-2.5-flash", "Gemini 2.5 Flash", window, output, 0.3, 2.5, 0.03),
		model("gemini-2.5-flash-lite", "Gemini 2.5 Flash-Lite", window, output, 0.1, 0.4, 0.01),
		model("gemini-2.5-pro", "Gemini 2.5 Pro", window, output, 1.25, 10, 0.125),
		levels(model("gemini-3-flash-preview", "Gemini 3 Flash Preview", window, output, 0.5, 3, 0.05)),
		levels(model("gemini-3.1-flash-lite", "Gemini 3.1 Flash Lite", window, output, 0.25, 1.5, 0.025)),
		levels(model("gemini-3.1-flash-lite-preview", "Gemini 3.1 Flash Lite Preview", window, output, 0.25, 1.5, 0.025)),
		levels(model("gemini-3.1-pro-preview", "Gemini 3.1 Pro Preview", window, output, 2, 12, 0.2), ThinkingMinimal),
		levels(model("gemini-3.1-pro-preview-customtools", "Gemini 3.1 Pro Preview Custom Tools", window, output, 2, 12, 0.2), ThinkingMinimal),
		levels(model("gemini-3.5-flash", "Gemini 3.5 Flash", window, output, 1.5, 9, 0.15)),
		levels(model("gemini-3.5-flash-lite", "Gemini 3.5 Flash Lite", window, output, 0.3, 2.5, 0.03)),
		levels(model("gemini-3.6-flash", "Gemini 3.6 Flash", window, output, 0.75, 3.75, 0.075)),
		levels(model("gemini-3.7-flash", "Gemini 3.7 Flash", window, output, 0.75, 3.75, 0.075), ThinkingMinimal),
		levels(model("gemini-3.8-flash", "Gemini 3.8 Flash", window, output, 0.75, 3.75, 0.075), ThinkingMinimal),
		levels(model("gemini-flash-latest", "Gemini Flash Latest", window, output, 1.5, 9, 0.15)),
		levels(model("gemini-flash-lite-latest", "Gemini Flash-Lite Latest", window, output, 0.25, 1.5, 0.025)),
		gemma("gemma-4-26b-a4b-it", "Gemma 4 26B A4B IT"),
		gemma("gemma-4-31b-it", "Gemma 4 31B IT"),
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
