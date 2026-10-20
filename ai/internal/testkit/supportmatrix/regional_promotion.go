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
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

const regionalActivationVersion = "2026-10-20.1-regional"
const regionalInactiveSHA256 = "7923e919d03e274d940a164da65d747ef06247a74fbc99d442f93996c004787a"

// RegionalPromotionProof pins the original, independently consumed candidate.
// The receipt preserves lineage; it does not rewrite a candidate as formal.
type RegionalPromotionProof struct {
	Combo             string            `json:"combo"`
	Session           string            `json:"session"`
	GitCommit         string            `json:"gitCommit"`
	SDK               string            `json:"sdk"`
	SourceSHA256      string            `json:"sourceSha256"`
	BuildSourceSHA256 string            `json:"buildSourceSha256"`
	BinarySHA256      string            `json:"binarySha256"`
	Files             map[string]string `json:"files"`
}

type RegionalPromotion struct {
	Schema                  int                      `json:"schema"`
	Kind                    string                   `json:"kind"`
	Mode                    string                   `json:"mode"`
	CandidateCatalogVersion string                   `json:"candidateCatalogVersion"`
	CandidateCatalogHash    string                   `json:"candidateCatalogHash"`
	CatalogSHA256           string                   `json:"catalogSha256"`
	PreviousCatalogSHA256   string                   `json:"previousCatalogSha256"`
	Target                  string                   `json:"target"`
	Candidates              []RegionalPromotionProof `json:"candidates"`
	Catalog                 json.RawMessage          `json:"catalog"`
}

// RegionalPromotionProposal accepts six distinct own live candidate originals.
// Offline controlled captures have a separate, explicitly nonproduction path.
func RegionalPromotionProposal(dirs []string) (RegionalPromotion, error) {
	return regionalPromotionProposal(dirs, "live")
}

func regionalPromotionProposal(dirs []string, mode string) (RegionalPromotion, error) {
	out := RegionalPromotion{Schema: 1, Kind: "verified-regional-catalog-activation-intent", Mode: mode, PreviousCatalogSHA256: regionalInactiveSHA256, Target: "ai/catalog_regional.json"}
	if mode == "controlled" {
		out.Kind = "controlled-regional-catalog-proposal"
	}
	if len(dirs) != 6 {
		return out, refuse("regional promotion requires exactly six own candidates")
	}
	seen := map[string]bool{}
	var source, build, binary, commit string
	for _, dir := range dirs {
		r, err := regionalReport(dir, false, mode)
		if err != nil {
			return out, err
		}
		e := r.Regional
		if e.Stage != "candidate" || e.Mode != mode || seen[r.Combo] {
			return out, refuse("regional promotion needs distinct own live candidate routes")
		}
		if source != "" && (source != e.SourceSHA256 || build != e.BuildSourceSHA256 || binary != e.BinarySHA256 || commit != r.GitCommit || out.CandidateCatalogVersion != r.CatalogVersion || out.CandidateCatalogHash != r.CatalogHash) {
			return out, refuse("regional candidate source, build or catalog differs")
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
		// A replacement between verification and pin collection cannot become a
		// different accepted original. The second full consumer rechecks its seal.
		again, err := regionalReport(dir, false, mode)
		if err != nil || !reflect.DeepEqual(r, again) {
			return out, refuse("regional candidate changed during promotion")
		}
		root, err = os.OpenRoot(dir)
		if err != nil {
			return out, err
		}
		for name, pin := range proof.Files {
			raw, err := regionalRead(root, name)
			if err != nil || glmHash(raw) != pin {
				root.Close()
				return out, refuse("regional candidate pins changed during promotion")
			}
		}
		if err := root.Close(); err != nil {
			return out, err
		}
		out.Candidates = append(out.Candidates, proof)
	}
	for _, combo := range []string{"zai-chat", "zai-responses", "zai-messages", "minimax-cn-chat", "minimax-cn-responses", "minimax-cn-messages"} {
		if !seen[combo] {
			return out, refuse("regional own candidate route missing")
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
	cat, err := regional.CandidateCatalog()
	if err != nil {
		return out, err
	}
	payload := struct {
		Schema  int        `json:"schema"`
		Active  bool       `json:"active"`
		Version string     `json:"version"`
		Models  []ai.Model `json:"models"`
	}{1, mode == "live", regionalActivationVersion, cat.Models}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return out, err
	}
	out.Catalog = append(raw, '\n')
	out.CatalogSHA256 = glmHash(out.Catalog)
	return out, nil
}

// ActivateRegionalCatalog durably records a new immutable activation intent
// before atomically replacing only the inactive data. An interrupted intent is
// retained as history and cannot be overwritten or treated as completed proof.
func ActivateRegionalCatalog(dirs []string, receiptPath string) error {
	if receiptPath == "" {
		return refuse("regional exclusive promotion receipt required")
	}
	root, err := minimaxlive.Root()
	if err != nil {
		return err
	}
	path := filepath.Join(root, "ai/catalog_regional.json")
	lock := path + ".activation.lock"
	if err := os.Mkdir(lock, 0700); err != nil {
		return fmt.Errorf("regional activation locked")
	}
	defer os.Remove(lock)
	unchanged := func() error {
		raw, err := minimaxlive.ReadFile(path, 2<<20)
		if err != nil || glmHash(raw) != regionalInactiveSHA256 {
			return refuse("regional inactive catalog target changed")
		}
		return nil
	}
	if err := unchanged(); err != nil {
		return err
	}
	receiptPath, err = regionalPromotionPath(root, receiptPath, dirs)
	if err != nil {
		return err
	}
	proposal, err := RegionalPromotionProposal(dirs)
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

// WriteRegionalControlledProposal only exercises the offline consumer. Its
// inactive payload is deliberately invalid as production activation data.
func WriteRegionalControlledProposal(dirs []string, path string) error {
	root, err := minimaxlive.Root()
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return err
	}
	path = filepath.Join(parent, filepath.Base(path))
	for _, forbidden := range append([]string{filepath.Join(root, "ai")}, dirs...) {
		base, err := filepath.EvalSymlinks(forbidden)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return refuse("controlled proposal cannot mutate production or original evidence")
		}
	}
	proposal, err := regionalPromotionProposal(dirs, "controlled")
	if err != nil {
		return err
	}
	raw, err := Encode(proposal)
	if err != nil {
		return err
	}
	return regionalPromotionReceipt(path, raw)
}
