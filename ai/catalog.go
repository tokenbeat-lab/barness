package ai

import (
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
	// SupportsLongCacheRetention: long cache retention is requested. Unset
	// or null means true, as in pi.
	SupportsLongCacheRetention Nullable[bool] `json:"supportsLongCacheRetention,omitzero"`
	// SupportsExplicitPromptCacheMode: cache retention is sent as
	// prompt_cache_options instead of prompt_cache_retention.
	SupportsExplicitPromptCacheMode bool `json:"supportsExplicitPromptCacheMode,omitempty"`
}

// acceptsImages reports whether the model takes image input; any other model
// gets pi-ai's placeholder text instead of images.
func (m Model) acceptsImages() bool { return slices.Contains(m.Input, ModalityImage) }

// Catalog is a versioned model catalog. It is never updated online.
type Catalog struct {
	Version string  `json:"version"`
	Models  []Model `json:"models"`
}

// BuiltinCatalog returns a fresh copy of the built-in catalog.
//
// Source: the frozen pi-ai 0.87.1 model data (providers/data/openai.json,
// sha256 3c52c858…2835) from the differential oracle's rebuilt copy, see
// internal/testkit/pioracle/node/PROVENANCE.md. Every listed field was
// re-checked against it when the oracle landed (compat when tools landed,
// gpt-4 when image placeholders landed, the reasoning models and their level
// maps when reasoning mapping landed). Pricing arrives with ticket 15.
// Models whose compat needs behavior barness-ai does not implement
// (additional_tools, tool search) are not listed, so nothing here promises
// behavior the adapter does not have. pi's supportsOpenAIGrammarTools only
// affects grammar-constrained tools, which cannot be declared here, so it is
// not carried. gpt-4 and o3-mini are the text-only models: images reach them
// as placeholders.
func BuiltinCatalog() Catalog {
	textOnly := func() []Modality { return []Modality{ModalityText} }
	textAndImage := func() []Modality { return []Modality{ModalityText, ModalityImage} }
	strict := ModelCompat{SupportsStrictMode: true}
	model := func(id, name string, input []Modality, contextWindow, maxTokens int) Model {
		return Model{Provider: ProviderOpenAI, API: APIOpenAIResponses, ID: id, Name: name, Input: input,
			ContextWindow: contextWindow, MaxTokens: maxTokens, Compat: strict}
	}
	reasoning := func(m Model, levels ThinkingLevelMap) Model {
		m.Reasoning, m.ThinkingLevelMap = true, levels
		return m
	}
	return Catalog{
		Version: "2026-10-01.4",
		Models: []Model{
			model("gpt-4", "GPT-4", textOnly(), 8192, 8192),
			model("gpt-4.1", "GPT-4.1", textAndImage(), 1047576, 32768),
			model("gpt-4.1-mini", "GPT-4.1 mini", textAndImage(), 1047576, 32768),
			model("gpt-4o-mini", "GPT-4o mini", textAndImage(), 128000, 16384),
			reasoning(model("gpt-5", "GPT-5", textAndImage(), 400000, 128000), levelMap("", ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh)),
			reasoning(model("gpt-5-mini", "GPT-5 Mini", textAndImage(), 400000, 128000), levelMap("", ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh)),
			reasoning(model("gpt-5-nano", "GPT-5 Nano", textAndImage(), 400000, 128000), levelMap("", ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh)),
			reasoning(model("gpt-5-pro", "GPT-5 Pro", textAndImage(), 400000, 128000), levelMap("", ThinkingHigh)),
			reasoning(model("gpt-5.1", "GPT-5.1", textAndImage(), 400000, 128000), levelMap("none", ThinkingLow, ThinkingMedium, ThinkingHigh)),
			reasoning(model("gpt-5.2", "GPT-5.2", textAndImage(), 400000, 128000), levelMap("none", ThinkingLow, ThinkingMedium, ThinkingHigh, ThinkingXHigh)),
			reasoning(model("o3", "o3", textAndImage(), 200000, 100000), levelMap("", ThinkingLow, ThinkingMedium, ThinkingHigh)),
			reasoning(model("o3-mini", "o3-mini", textOnly(), 200000, 100000), levelMap("", ThinkingLow, ThinkingMedium, ThinkingHigh)),
			reasoning(model("o4-mini", "o4-mini", textAndImage(), 200000, 100000), levelMap("", ThinkingLow, ThinkingMedium, ThinkingHigh)),
		},
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
		out.Models[i] = m
	}
	return out
}

func (c Catalog) lookup(provider ProviderID, api API, id string) (Model, bool) {
	for _, m := range c.Models {
		if m.Provider == provider && m.API == api && m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}
