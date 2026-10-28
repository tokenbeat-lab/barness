package supportmatrix

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

const kimiActivationVersion = "kimi-k3-formal-2026-10-22.1"
const kimiInactiveSHA256 = "c4daa0ee0c4219762157a4a506768e40f034eb4c2cc014ffb5f8a0ba42ab4109"

// KimiPromotionProof keeps a fixed pair and its entire transitive inventory.
// It cannot be represented by the regional single-measurement proof: original
// source, commit, time, grant and budget belong to each measurement separately.
type KimiPromotionProof struct {
	Combo         string            `json:"combo"`
	Composition   string            `json:"composition"`
	Qualification KimiQualification `json:"qualification"`
	Inputs        map[string]string `json:"inputs"`
}

type KimiPromotion struct {
	Schema                  int                  `json:"schema"`
	Kind                    string               `json:"kind"`
	Mode                    string               `json:"mode"`
	CandidateCatalogVersion string               `json:"candidateCatalogVersion"`
	CandidateCatalogHash    string               `json:"candidateCatalogHash"`
	CatalogSHA256           string               `json:"catalogSha256"`
	PreviousCatalogSHA256   string               `json:"previousCatalogSha256"`
	Target                  string               `json:"target"`
	Candidates              []KimiPromotionProof `json:"candidates"`
	Catalog                 json.RawMessage      `json:"catalog"`
}

// This dedicated seam consumes six fixed nine-plus-one compositions. Native
// verification checks each original against current route facts; equality of
// unrelated measurements' source/build/binary/commit is not applicability.
func KimiPromotionProposal(dirs []string) (KimiPromotion, error) {
	return kimiPromotionProposal(dirs, "live")
}

func kimiPromotionProposal(dirs []string, mode string) (KimiPromotion, error) {
	out := KimiPromotion{Schema: 2, Kind: "verified-kimi-catalog-activation-intent", Mode: mode, PreviousCatalogSHA256: kimiInactiveSHA256, Target: "ai/catalog_kimi.json"}
	if mode == "controlled" {
		out.Kind = "controlled-kimi-catalog-proposal"
	}
	if len(dirs) != 6 {
		return out, refuse("Kimi promotion requires exactly six fixed nine-plus-one compositions")
	}
	root, err := kimiRepositoryRoot()
	if err != nil {
		return out, err
	}
	seen := map[string]bool{}
	for _, dir := range dirs {
		absolute, err := filepath.Abs(dir)
		if err != nil {
			return out, err
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return out, err
		}
		relative = filepath.ToSlash(relative)
		checked, err := kimiLocalPath(root, relative)
		if err != nil {
			return out, err
		}
		q, err := VerifyKimiComposition(checked, false)
		if err != nil {
			return out, fmt.Errorf("Kimi paired candidate proof rejected: %w", err)
		}
		if q.Mode != mode || seen[q.Combo] {
			return out, refuse("Kimi promotion requires six distinct own compositions in the requested mode")
		}
		inputs, err := KimiCompositionInputs(checked)
		if err != nil {
			return out, err
		}
		again, err := VerifyKimiComposition(checked, false)
		if err != nil || !reflect.DeepEqual(q, again) {
			return out, refuse("Kimi composition changed during promotion")
		}
		pinned, err := KimiCompositionInputs(checked)
		if err != nil || !reflect.DeepEqual(inputs, pinned) {
			return out, refuse("Kimi transitive input pins changed during promotion")
		}
		seen[q.Combo] = true
		out.Candidates = append(out.Candidates, KimiPromotionProof{Combo: q.Combo, Composition: relative, Qualification: q, Inputs: inputs})
	}
	for _, combo := range []string{"kimi-cn-chat", "kimi-cn-responses", "kimi-cn-messages", "kimi-intl-chat", "kimi-intl-responses", "kimi-intl-messages"} {
		if !seen[combo] {
			return out, refuse("Kimi own composition route missing")
		}
	}
	slices.SortFunc(out.Candidates, func(a, b KimiPromotionProof) int { return strings.Compare(a.Combo, b.Combo) })
	cat := kimi.CandidateCatalog()
	out.CandidateCatalogVersion = cat.Version
	out.CandidateCatalogHash, err = cat.Hash()
	if err != nil {
		return out, err
	}
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
	if err := kimiPromotionProtectOriginals(receiptPath, proposal); err != nil {
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
