package xai

import (
	_ "embed"
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

// imagesCandidateData fixes the agreed global Imagine 2.0 subset and the
// 2026-10-29 USD/image price snapshot. Official capability/price research is
// retained in .scratch/barness-ai-xai-imagine/research.md. This candidate
// neither enters BuiltinCatalog nor confers live qualification.
//
//go:embed images-candidate.json
var imagesCandidateData []byte

// ImagesCandidateCatalog returns an independent host-configured candidate
// snapshot. Controlled checks and explicitly authorized live requests use
// the same input; all actual authorization still belongs to the Host.
func ImagesCandidateCatalog() ai.Catalog {
	var cat ai.Catalog
	if err := json.Unmarshal(imagesCandidateData, &cat); err != nil {
		panic("invalid xAI images candidate JSON")
	}
	if _, err := cat.Hash(); err != nil {
		panic("invalid xAI images candidate metadata")
	}
	return cat
}
