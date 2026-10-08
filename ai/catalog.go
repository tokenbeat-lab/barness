package ai

import "slices"

// Catalog is a versioned model catalog and price snapshot. It is never
// updated online. Version names a content: a catalog whose models or prices
// change gets a new Version, and Hash tells two contents apart even when a
// host reuses one.
type Catalog struct {
	Version          string            `json:"version"`
	Models           []Model           `json:"models"`
	ImageModels      []ImageModel      `json:"imageModels,omitempty"`
	ClassifierModels []ClassifierModel `json:"classifierModels,omitempty"`
}

// BuiltinCatalog returns a fresh copy of the built-in catalog.
//
// OpenAI image facts and own live acceptance: ADR-0022, fixed date snapshot,
// sourced token rates and conservatively verified capabilities on 2026-10-08.
//
// TypeSafe classifier facts: https://docs.typesafe.ai/models and /api,
// read 2026-10-08; own live evidence and fixed version inclusion: ADR-0021.
//
// Source: frozen pi-ai 1.0.0 providers/data/{openai,anthropic,google,deepseek}.json.
// Per-file hashes and the complete model-data hash are pinned in
// internal/testkit/pioracle/node/PROVENANCE.md. Included fields are checked
// against that data by TestCatalogInclusion, including the 1.0 additions
// evaluated under ADR-0018 and the managed effort implemented in ADR-0019.
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
		Version: "2026-10-08.4",
		Models: slices.Concat(builtinOpenAIModels(), builtinOpenAIChatModels(), builtinAnthropicModels(),
			builtinGoogleModels(), builtinDeepSeekResponsesModels(), builtinDeepSeekChatModels()),
		ImageModels:      builtinOpenAIImageModels(),
		ClassifierModels: builtinTypeSafeModels(),
	}
}
