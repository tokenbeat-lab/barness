package minimaxlive

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

type PromotionProof struct {
	Route                string `json:"route"`
	ReportSHA256         string `json:"reportSha256"`
	SealSHA256           string `json:"sealSha256"`
	CandidateCatalogHash string `json:"candidateCatalogHash"`
	SourceHash           string `json:"sourceHash"`
	PlanHash             string `json:"planHash"`
	SDK                  string `json:"sdk"`
	GitCommit            string `json:"gitCommit"`
}
type Promotion struct {
	Schema        int              `json:"schema"`
	Kind          string           `json:"kind"`
	CatalogSHA256 string           `json:"catalogSha256"`
	Candidates    []PromotionProof `json:"candidates"`
	Catalog       json.RawMessage  `json:"catalog"`
}

// PromotionProposal requires actual current-source own candidate acceptance.
// Its output preserves historical lineage only. It cannot activate the current
// maintenance catalog or manufacture a live PASS.
func PromotionProposal(dirs []string) (Promotion, error) {
	out := Promotion{Schema: 1, Kind: "verified-candidates-for-catalog-activation"}
	if len(dirs) != 3 {
		return out, fmt.Errorf("exactly three own MiniMax candidate bundles required")
	}
	seen := map[string]bool{}
	for _, dir := range dirs {
		r, e := Verify(dir, false)
		if e != nil {
			return out, e
		}
		if r.Identity.Model != M3 || r.Identity.Purpose != "standard" || r.Stage != "candidate" || r.Kind != "minimax-candidate-only" || r.Mode != "live" || seen[r.Identity.Route] {
			return out, fmt.Errorf("only three distinct own actual M3 standard candidates can promote")
		}
		seen[r.Identity.Route] = true
		raw, e := ReadFile(filepath.Join(dir, ReportFile), 2<<20)
		if e != nil {
			return out, e
		}
		seal, e := ReadFile(filepath.Join(dir, SealFile), 2<<20)
		if e != nil {
			return out, e
		}
		out.Candidates = append(out.Candidates, PromotionProof{r.Identity.Route, Hash(raw), Hash(seal), r.CatalogHash, r.SourceHash, r.PlanHash, r.SDK, r.GitCommit})
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		if !seen[route] {
			return out, fmt.Errorf("own M3 route missing")
		}
	}
	slices.SortFunc(out.Candidates, func(a, b PromotionProof) int {
		if a.Route < b.Route {
			return -1
		}
		if a.Route > b.Route {
			return 1
		}
		return 0
	})
	catalog := minimax.PricedM3Catalog()
	catalog.Version = "2026-10-18.1-minimax"
	raw, e := json.MarshalIndent(catalog, "", "  ")
	if e != nil {
		return out, e
	}
	out.Catalog = append(raw, '\n')
	out.CatalogSHA256 = Hash(out.Catalog)
	return out, nil
}
