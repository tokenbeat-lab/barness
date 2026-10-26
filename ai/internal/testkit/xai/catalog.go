// Package xai holds the independently configured Grok Responses candidate.
// Its availability confers no live qualification or built-in activation.
package xai

import (
	_ "embed"
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

// candidateData is the reviewed 2026-10-26 subset, not a model-weight lock.
// Provenance: .scratch/barness-ai-grok/research.md and design.md §§4,12;
// https://docs.x.ai/developers/grok-4-7 and /developers/pricing. Limits are
// conservative integration bounds. The >=200000 threshold is represented by
// the public Catalog's strict-above threshold 199999, including cache input.
//
//go:embed candidate.json
var candidateData []byte

// CandidateCatalog returns an independent copy for controlled E2E and an
// explicitly authorized candidate live run. Hosts load the same JSON through
// their existing Catalog configuration; BuiltinCatalog is never modified.
func CandidateCatalog() ai.Catalog {
	var cat ai.Catalog
	if err := json.Unmarshal(candidateData, &cat); err != nil {
		panic("invalid xAI candidate JSON")
	}
	if _, err := cat.Hash(); err != nil {
		panic("invalid xAI candidate metadata")
	}
	return cat
}
