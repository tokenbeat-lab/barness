package supportmatrix

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

const kimiActivationVersion = "kimi-k3-formal-2026-10-22.1"
const kimiInactiveSHA256 = "c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109"

// Kimi promotion has its own fixed six-route domain and data target. Repeated
// selection/proof checks are intentional: future Kimi reviews must not widen
// the old regional nine-model activation or inherit its approved receipts.
func KimiPromotionProposal(dirs []string) (RegionalPromotion, error) {
	return kimiPromotionProposal(dirs, "live")
}

func kimiPromotionProposal(dirs []string, mode string) (RegionalPromotion, error) {
	out := RegionalPromotion{Schema: 1, Kind: "verified-kimi-catalog-activation-intent", Mode: mode, PreviousCatalogSHA256: kimiInactiveSHA256, Target: "ai/catalog_kimi.json"}
	if mode == "controlled" {
		out.Kind = "controlled-kimi-catalog-proposal"
	}
	if len(dirs) != 6 {
		return out, refuse("Kimi promotion requires exactly six own candidates")
	}
	seen := map[string]bool{}
	var source, build, binary, commit string
	for _, dir := range dirs {
		r, err := regionalReport(dir, false, mode)
		if err != nil {
			return out, fmt.Errorf("Kimi own candidate proof rejected: %w", err)
		}
		e := r.Regional
		if !RegionalIsKimi(r.Combo) || e.Stage != "candidate" || e.Mode != mode || seen[r.Combo] {
			return out, refuse("Kimi promotion requires six distinct own candidate routes")
		}
		if source != "" && (source != e.SourceSHA256 || build != e.BuildSourceSHA256 || binary != e.BinarySHA256 || commit != r.GitCommit || out.CandidateCatalogVersion != r.CatalogVersion || out.CandidateCatalogHash != r.CatalogHash) {
			return out, refuse("Kimi current candidate source/build/catalog differs")
		}
		source, build, binary, commit = e.SourceSHA256, e.BuildSourceSHA256, e.BinarySHA256, r.GitCommit
		out.CandidateCatalogVersion, out.CandidateCatalogHash = r.CatalogVersion, r.CatalogHash
		seen[r.Combo] = true
		proof := RegionalPromotionProof{Combo: r.Combo, Session: filepath.Base(filepath.Clean(dir)), GitCommit: r.GitCommit, SDK: r.SDK, SourceSHA256: e.SourceSHA256, BuildSourceSHA256: e.BuildSourceSHA256, BinarySHA256: e.BinarySHA256, Files: map[string]string{}}
		root, err := os.OpenRoot(dir)
		if err != nil {
			return out, err
		}
		for _, name := range []string{RegionalReportFile, regionalSealFile, "manifest.json", "regional-grant.json", "regional-authorization-receipt.json", "regional-execution-source.json", "regional-catalog.json", "regional-candidate-facts.json"} {
			raw, err := regionalRead(root, name)
			if err != nil {
				root.Close()
				return out, err
			}
			proof.Files[name] = glmHash(raw)
		}
		if err := root.Close(); err != nil {
			return out, err
		}
		again, err := regionalReport(dir, false, mode)
		if err != nil || !reflect.DeepEqual(r, again) {
			return out, refuse("Kimi candidate changed during promotion")
		}
		root, err = os.OpenRoot(dir)
		if err != nil {
			return out, err
		}
		for name, pin := range proof.Files {
			raw, err := regionalRead(root, name)
			if err != nil || glmHash(raw) != pin {
				root.Close()
				return out, refuse("Kimi candidate pins changed during promotion")
			}
		}
		if err := root.Close(); err != nil {
			return out, err
		}
		out.Candidates = append(out.Candidates, proof)
	}
	for _, combo := range []string{"kimi-cn-chat", "kimi-cn-responses", "kimi-cn-messages", "kimi-intl-chat", "kimi-intl-responses", "kimi-intl-messages"} {
		if !seen[combo] {
			return out, refuse("Kimi own candidate route missing")
		}
	}
	slices.SortFunc(out.Candidates, func(a, b RegionalPromotionProof) int {
		if a.Combo < b.Combo {
			return -1
		}
		if a.Combo > b.Combo {
			return 1
		}
		return 0
	})
	cat := kimi.CandidateCatalog()
	payload := struct {
		Schema           int        `json:"schema"`
		Active           bool       `json:"active"`
		Version          string     `json:"version"`
		ProvenanceSHA256 string     `json:"provenanceSha256"`
		Models           []ai.Model `json:"models"`
	}{1, mode == "live", kimiActivationVersion, kimi.SnapshotHash(), cat.Models}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return out, err
	}
	out.Catalog = append(raw, '\n')
	out.CatalogSHA256 = glmHash(out.Catalog)
	return out, nil
}

func ActivateKimiCatalog(dirs []string, receiptPath string) error {
	if receiptPath == "" {
		return refuse("Kimi exclusive promotion receipt required")
	}
	root, err := minimaxlive.Root()
	if err != nil {
		return err
	}
	path := filepath.Join(root, "ai/catalog_kimi.json")
	lock := path + ".activation.lock"
	if err := os.Mkdir(lock, 0700); err != nil {
		return refuse("Kimi activation locked")
	}
	defer os.Remove(lock)
	unchanged := func() error {
		raw, err := minimaxlive.ReadFile(path, 2<<20)
		if err != nil || glmHash(raw) != kimiInactiveSHA256 {
			return refuse("Kimi inactive catalog target changed")
		}
		return nil
	}
	if err := unchanged(); err != nil {
		return err
	}
	receiptPath, err = kimiPromotionPath(root, receiptPath, dirs)
	if err != nil {
		return err
	}
	proposal, err := KimiPromotionProposal(dirs)
	if err != nil {
		return err
	}
	raw, err := Encode(proposal)
	if err != nil {
		return err
	}
	if err := regionalPromotionReceipt(receiptPath, raw); err != nil {
		return err
	}
	if err := unchanged(); err != nil {
		return err
	}
	return regionalPromotionReplace(path, proposal.Catalog)
}
