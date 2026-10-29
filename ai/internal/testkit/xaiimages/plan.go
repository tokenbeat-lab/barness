// Package xaiimages holds the finite Imagine candidate plan. It grants no
// execution permission and makes no published-model support declaration.
package xaiimages

const (
	Route    = "xai-images"
	Model    = "grok-imagine-image-2.0"
	Endpoint = "https://api.x.ai/v1"
	Entry    = "^TestXAIImagesCandidate$"
	SDK      = "direct HTTP (xAI Images; ADR-0056)"
	MaxUSD   = 0.60
)

// Scene fixes every charge-bearing option before approval. No environment
// parameter can expand counts, choose another tier, or add a request rerun.
type Scene struct {
	ID          string  `json:"id"`
	References  int     `json:"references"`
	Outputs     int     `json:"outputs"`
	Quality     string  `json:"quality"`
	Resolution  string  `json:"resolution"`
	AspectRatio string  `json:"aspectRatio,omitempty"`
	MaximumUSD  float64 `json:"maximumUSD"`
}

// Plan returns independent values. The sum of the conservative image rates
// from the 2026-10-29 snapshot is USD 0.53, below the USD 0.60 stop threshold.
func Plan() []Scene {
	return []Scene{
		{"generation-low", 0, 1, "low", "1k", "1:1", 0.04},
		{"generation-medium", 0, 2, "medium", "1.5k", "16:9", 0.14},
		{"generation-auto", 0, 1, "auto", "2k", "1:1", 0.08},
		{"single-edit", 1, 1, "medium", "1k", "", 0.07},
		{"ordered-edit", 2, 1, "low", "1.5k", "16:9", 0.07},
		{"five-reference-edit", 5, 1, "auto", "2k", "1:1", 0.13},
	}
}

func Expected() []string {
	var out []string
	for _, scene := range Plan() {
		out = append(out, scene.ID)
	}
	return out
}
