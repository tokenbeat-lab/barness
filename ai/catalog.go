package ai

import "slices"

// Modality is an input kind a model accepts.
type Modality string

const (
	ModalityText  Modality = "text"
	ModalityImage Modality = "image"
)

// Model is secret-free, versioned model metadata located by (Provider, API, ID).
type Model struct {
	Provider      ProviderID  `json:"provider"`
	API           API         `json:"api"`
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Reasoning     bool        `json:"reasoning"`
	Input         []Modality  `json:"input"`
	ContextWindow int         `json:"contextWindow"`
	MaxTokens     int         `json:"maxTokens"`
	Compat        ModelCompat `json:"compat,omitzero"`
}

// ModelCompat holds the protocol compatibility flags of pi-ai's model data
// that barness-ai implements. An absent flag has pi's default.
type ModelCompat struct {
	// SupportsStrictMode: Responses function tools carry a "strict" field.
	SupportsStrictMode bool `json:"supportsStrictMode,omitempty"`
	// SupportsMidConvoSystemMessages: later system messages are sent in place
	// as instruction updates instead of folding into the leading one. No
	// built-in model sets it yet; pi sets it only on reasoning models, which
	// arrive with ticket 09.
	SupportsMidConvoSystemMessages bool `json:"supportsMidConvoSystemMessages,omitempty"`
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
// gpt-4 when image placeholders landed). Only non-reasoning OpenAI ×
// Responses models are listed until reasoning mapping (ticket 09) and pricing
// (ticket 15) land, so nothing here promises behavior the adapter does not yet
// implement. gpt-4 is the text-only model: images reach it as placeholders.
func BuiltinCatalog() Catalog {
	textOnly := func() []Modality { return []Modality{ModalityText} }
	textAndImage := func() []Modality { return []Modality{ModalityText, ModalityImage} }
	strict := ModelCompat{SupportsStrictMode: true}
	return Catalog{
		Version: "2026-10-01.3",
		Models: []Model{
			{Provider: ProviderOpenAI, API: APIOpenAIResponses, ID: "gpt-4", Name: "GPT-4", Input: textOnly(), ContextWindow: 8192, MaxTokens: 8192, Compat: strict},
			{Provider: ProviderOpenAI, API: APIOpenAIResponses, ID: "gpt-4.1", Name: "GPT-4.1", Input: textAndImage(), ContextWindow: 1047576, MaxTokens: 32768, Compat: strict},
			{Provider: ProviderOpenAI, API: APIOpenAIResponses, ID: "gpt-4.1-mini", Name: "GPT-4.1 mini", Input: textAndImage(), ContextWindow: 1047576, MaxTokens: 32768, Compat: strict},
			{Provider: ProviderOpenAI, API: APIOpenAIResponses, ID: "gpt-4o-mini", Name: "GPT-4o mini", Input: textAndImage(), ContextWindow: 128000, MaxTokens: 16384, Compat: strict},
		},
	}
}

func (c Catalog) clone() Catalog {
	out := Catalog{Version: c.Version, Models: make([]Model, len(c.Models))}
	for i, m := range c.Models {
		m.Input = append([]Modality(nil), m.Input...)
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
