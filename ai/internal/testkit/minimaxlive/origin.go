package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// Origin is a trusted launcher fact, never accepted merely because it appears
// inside an imported bundle. It pins the unchanged child report/seal after
// fixed live argv, exact-source approval, successful exit and credential audit.
type Origin struct {
	Schema                 int      `json:"schema"`
	Stage                  string   `json:"stage"`
	Identity               Identity `json:"identity"`
	Session                string   `json:"session"`
	ReportSHA256           string   `json:"reportSha256"`
	SealSHA256             string   `json:"sealSha256"`
	LedgerSHA256           string   `json:"ledgerSha256"`
	SourceCommit           string   `json:"sourceCommit"`
	SourceManifestSHA256   string   `json:"sourceManifestSha256"`
	BinariesManifestSHA256 string   `json:"binariesManifestSha256"`
	PlanSHA256             string   `json:"planSha256"`
}

// Verify adds the external actual-origin boundary to structural capture checks.
// Child self-checks use VerifyCapture; promotion and public formal consumers
// must use this function. Controlled/replay may prove mechanisms, never actual.
func Verify(dir string, formal bool) (Report, error) {
	r, e := Load(dir)
	if e != nil {
		return r, e
	}
	if r.Mode == "live" {
		if e := verifyOrigin(dir, r); e != nil {
			return r, e
		}
	}
	return VerifyCapture(dir, formal)
}

func verifyOrigin(dir string, r Report) error {
	id, e := Select(r.Identity.Route, r.Identity.Model, r.Identity.Purpose)
	if e != nil || id != r.Identity || r.Stage != "candidate" && r.Stage != "formal" {
		return fmt.Errorf("MiniMax actual origin identity invalid")
	}
	raw, e := ReadFile(filepath.Join(dir, ReportFile), 2<<20)
	if e != nil {
		return e
	}
	reportHash := Hash(raw)
	stage, e := StageLedgerPath(r.Stage)
	if e != nil {
		return fmt.Errorf("MiniMax actual origin root unavailable")
	}
	short := "m3"
	if id.Model == M31 {
		short = "m31"
	}
	route := "mm-global-" + r.Stage + "-" + short + "-" + id.Route
	receiptPath := filepath.Join(filepath.Dir(filepath.Dir(stage)), "minimax-launches", route+".json")
	receiptRaw, receiptErr := ReadFile(receiptPath, 2<<20)
	var origin Origin
	if receiptErr == nil {
		var receipt struct {
			State           string `json:"state"`
			Route           string `json:"route"`
			SourceCommit    string `json:"sourceCommit"`
			ExitCode        int    `json:"exitCode"`
			CredentialsRead bool   `json:"credentialsRead"`
			Audit           struct {
				Passed bool `json:"passed"`
			} `json:"launcherAudit"`
			Origin Origin `json:"origin"`
		}
		if json.Unmarshal(receiptRaw, &receipt) != nil || receipt.State != "COMPLETED" || receipt.Route != route || receipt.SourceCommit != r.GitCommit || receipt.ExitCode != 0 || !receipt.CredentialsRead || !receipt.Audit.Passed {
			return fmt.Errorf("MiniMax actual origin launcher receipt invalid")
		}
		origin = receipt.Origin
		// Canonical append-only history may have later identities; this process's
		// complete snapshot must be an exact prefix, not a copied synthetic chain.
		captured, e := ReadFile(filepath.Join(dir, "minimax-ledger.json"), 16<<20)
		var ledger LedgerEvidence
		current, err := ReadLedger(stage)
		if e != nil || StrictJSON(captured, &ledger) != nil || err != nil || len(ledger.Receipts) > len(current) || !reflect.DeepEqual(ledger.Receipts, current[:len(ledger.Receipts)]) {
			return fmt.Errorf("MiniMax actual origin canonical ledger differs")
		}
	} else {
		if !os.IsNotExist(receiptErr) {
			return fmt.Errorf("MiniMax actual origin receipt unavailable")
		}
		// Portable archives use only this curated repository verifier index, never
		// an index supplied by the bundle. Generation does not read this data.
		root, e := Root()
		if e != nil {
			return e
		}
		indexRaw, e := ReadFile(filepath.Join(root, "ai/live/testdata/minimax-formal/origins.json"), 1<<20)
		var index struct {
			Schema  int `json:"schema"`
			Entries map[string]struct {
				Origin        Origin `json:"origin"`
				ReceiptSHA256 string `json:"launcherReceiptSha256"`
			} `json:"entries"`
		}
		if e != nil || StrictJSON(indexRaw, &index) != nil || index.Schema != 1 {
			return fmt.Errorf("MiniMax actual origin trusted index unavailable")
		}
		entry, ok := index.Entries[reportHash]
		if !ok || len(entry.ReceiptSHA256) != 64 {
			return fmt.Errorf("MiniMax actual origin has no trusted launcher pin")
		}
		origin = entry.Origin
	}
	seal, e := ReadFile(filepath.Join(dir, SealFile), 2<<20)
	if e != nil {
		return e
	}
	ledger, e := ReadFile(filepath.Join(dir, "minimax-ledger.json"), 16<<20)
	if e != nil {
		return e
	}
	if origin.Schema != 1 || origin.Stage != r.Stage || origin.Identity != id || origin.Session != r.Session || origin.ReportSHA256 != reportHash || origin.SealSHA256 != Hash(seal) || origin.LedgerSHA256 != Hash(ledger) || origin.SourceCommit != r.GitCommit || len(origin.SourceManifestSHA256) != 64 || len(origin.BinariesManifestSHA256) != 64 || len(origin.PlanSHA256) != 64 {
		return fmt.Errorf("MiniMax actual origin pins differ")
	}
	return nil
}
