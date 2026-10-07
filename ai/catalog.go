package ai

import "slices"

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
// 474a010c…4e75, providers/data/google.json, sha256 328822707e…708c, and
// providers/data/deepseek.json, sha256 549a7ddb…4d0d)
// from the differential oracle's rebuilt copy, see
// internal/testkit/pioracle/node/PROVENANCE.md. Every listed field was
// re-checked against it when the oracle landed (compat when tools landed,
// gpt-4 when image placeholders landed, the reasoning models and their level
// maps when reasoning mapping landed, prices and gpt-5.5-pro when cost
// estimation landed, the Anthropic models when the Messages adapter landed,
// the Google models when the Gemini adapter landed).
// (ADR-0018 listed the models below whose compat only turns on optional
// features, when issue 34 landed, and the Anthropic models with managed
// mid-conversation effort when issue 27 implemented it.)
//
// A model is listed when its Provider × API is implemented and the adapter
// keeps all of its hard constraints, the compat flags without which the
// vendor rejects a reachable request (ADR-0018). Optional features barness-ai
// does not implement do not keep a model out: it is served as pi serves it
// with the feature off, and the difference is a ledger extension. Those are
// Anthropic's native mid-conversation tool changes and server-side fallback
// models, and OpenAI's additional_tools and tool search, so their flags
// (supportsMidConvoToolChanges, allowedFallbackModels,
// supportsAdditionalTools, supportsToolSearch) are not carried. pi's
// supportsOpenAIGrammarTools and Anthropic supportsStrictTools only affect
// constrained-sampling tools, which cannot be declared here, so they are not
// carried either. Not listed: Google models that generateContent alone
// cannot serve (see builtinGoogleModels).
// gpt-4 and o3-mini are the text-only models: images reach them as
// placeholders; every Anthropic and Google model takes images.
//
// pi's data lists OpenAI's models for Responses only; the OpenAI Chat
// Completions models are the same entries on that API (see
// builtinOpenAIChatModels), as a pi user configures a custom model.
// DeepSeek's Chat models are pi's DeepSeek data as is (see
// builtinDeepSeekChatModels); its Responses model is the same data on the
// Responses API, a route pi itself does not have (see
// builtinDeepSeekResponsesModels).
func BuiltinCatalog() Catalog {
	return Catalog{
		Version: "2026-10-03.2",
		Models: slices.Concat(builtinOpenAIModels(), builtinOpenAIChatModels(), builtinAnthropicModels(),
			builtinGoogleModels(), builtinDeepSeekResponsesModels(), builtinDeepSeekChatModels()),
	}
}

func (c Catalog) lookup(provider ProviderID, api API, id string) (Model, bool) {
	for _, m := range c.Models {
		if m.Provider == provider && m.API == api && m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}
