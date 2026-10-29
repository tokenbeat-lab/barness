package supportmatrix

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

// XAIImagesExecutionSource includes the candidate's actual per-image rates.
// The shared source inventory includes every native harness implementation.
func XAIImagesExecutionSource() (map[string]string, error) {
	files, err := RegionalExecutionSource()
	if err != nil {
		return nil, err
	}
	root, err := minimaxlive.Root()
	if err != nil {
		return nil, err
	}
	const path = "ai/internal/testkit/xai/images-candidate.json"
	raw, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return nil, err
	}
	files[path] = livegrant.Hash(raw)
	return files, nil
}

func xaiImagesProcessScope(mode, stage, binary string) (livegrant.Scope, error) {
	if stage != "candidate" || mode != "live" && mode != "controlled" {
		return livegrant.Scope{}, fmt.Errorf("Imagine supports only its own candidate stage")
	}
	cat := xai.ImagesCandidateCatalog()
	hash, err := cat.Hash()
	if err != nil {
		return livegrant.Scope{}, err
	}
	source, err := XAIImagesExecutionSource()
	if err != nil {
		return livegrant.Scope{}, err
	}
	return livegrant.Scope{Route: xaiimages.Route, Mode: mode, Stage: stage, Provider: "xai", API: "xai-images", Endpoint: xaiimages.Endpoint, SDK: xaiimages.SDK, CatalogVersion: cat.Version, CatalogHash: hash, SourceSHA256: RegionalSourceHash(source), BinarySHA256: binary, Models: []string{xaiimages.Model}, Scenes: xaiimages.Expected(), MaxCalls: 6, MaxAttempts: 6, CallSeconds: 180, ProcessSeconds: 1200, MaxUSD: xaiimages.MaxUSD}, nil
}
