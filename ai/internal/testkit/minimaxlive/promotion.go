package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
// Its output is data only; frozen runtime code cannot manufacture a live PASS.
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

// Activate fsyncs an exclusive candidate-lineage receipt before the data-only
// catalog replacement. A leftover receipt records an interrupted activation;
// neither the receipt nor activation may silently overwrite previous history.
func Activate(dirs []string, receiptPath string) error {
	if receiptPath == "" {
		return fmt.Errorf("exclusive promotion receipt is required")
	}
	root, e := Root()
	if e != nil {
		return e
	}
	path := filepath.Join(root, "ai/catalog_minimax.json")
	lock := path + ".activation.lock"
	if e := os.Mkdir(lock, 0700); e != nil {
		return fmt.Errorf("MiniMax activation locked")
	}
	defer os.Remove(lock)
	unchanged := func() error {
		raw, e := ReadFile(path, 2<<20)
		if e != nil || Hash(raw) != "7ba8d4b2fb192e384b10d3c2aa42e3fdcf115fc8434cc98dfa5a934631a59981" {
			return fmt.Errorf("MiniMax inactive catalog target changed")
		}
		return nil
	}
	if e := unchanged(); e != nil {
		return e
	}
	proposal, e := PromotionProposal(dirs)
	if e != nil {
		return e
	}
	receiptPath, e = filepath.Abs(receiptPath)
	if e != nil {
		return e
	}
	if receiptPath == path {
		return fmt.Errorf("receipt cannot replace catalog")
	}
	for _, dir := range dirs {
		base, e := filepath.Abs(dir)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(base, receiptPath)
		if e != nil || rel == "." || !filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("receipt cannot mutate a candidate bundle")
		}
	}
	raw, e := json.MarshalIndent(proposal, "", "  ")
	if e != nil {
		return e
	}
	receipt, e := os.OpenFile(receiptPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	_, e = receipt.Write(append(raw, '\n'))
	if e == nil {
		e = receipt.Sync()
	}
	closeErr := receipt.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return e
	}
	temp, e := os.CreateTemp(filepath.Dir(path), ".minimax-activation-")
	if e != nil {
		return e
	}
	defer os.Remove(temp.Name())
	if _, e = temp.Write(proposal.Catalog); e == nil {
		e = temp.Sync()
	}
	closeErr = temp.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return e
	}
	if e := unchanged(); e != nil {
		return e
	}
	return os.Rename(temp.Name(), path)
}
