package minimaxlive

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"path/filepath"
	"reflect"
)

func verifyProcessAuthorization(dir string, r Report) error {
	if r.ProcessGrantSHA256 == "" {
		return fmt.Errorf("fresh process approval missing")
	}
	g, raw, claim, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return e
	}
	if livegrant.Hash(raw) != r.ProcessGrantSHA256 || claim.Session != r.Session {
		return fmt.Errorf("fresh M3 approval attribution differs")
	}
	want, e := ProcessScope(r.Identity, r.Mode, r.Stage, g.Scope.BinarySHA256)
	if r.Kind == ThinkingKind {
		want, e = ThinkingOnScope(r.Identity, r.Mode, g.Scope.BinarySHA256, g.Scope.BaselineManifestSHA256)
	}
	if e != nil {
		return e
	}
	original, _ := json.Marshal(r.ExecutionSource)
	if Hash(original) != g.Scope.SourceSHA256 {
		return fmt.Errorf("original process source differs")
	}
	want.SourceSHA256, want.CatalogHash, want.CatalogVersion = g.Scope.SourceSHA256, r.CatalogHash, r.CatalogVersion
	if e = livegrant.VerifyEvidence(g, raw, claim, want, r.AccountAlias, r.RegionAlias, r.StartedAt, r.FinishedAt); e != nil {
		return e
	}
	path, e := livegrant.LedgerPath(claim)
	if e != nil {
		return e
	}
	rows, e := ReadLedger(path)
	if e != nil {
		return e
	}
	var captured LedgerEvidence
	bytes, e := ReadFile(filepath.Join(dir, "minimax-ledger.json"), 16<<20)
	if e != nil || json.Unmarshal(bytes, &captured) != nil || !reflect.DeepEqual(rows, captured.Receipts) {
		return fmt.Errorf("fresh M3 authoritative process ledger differs")
	}
	return nil
}
