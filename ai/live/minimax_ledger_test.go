//go:build live

package live

import (
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func captureMiniMaxLedger() error {
	r := minimaxReport
	receipts, e := minimaxlive.ReadLedger(minimaxLedger.Path)
	if os.IsNotExist(e) {
		receipts = nil
		e = nil
	}
	if e != nil {
		return e
	}
	r.Cumulative = minimaxlive.Summarize(receipts, r.Identity)
	ledger := minimaxlive.LedgerEvidence{Receipts: receipts, Predecessors: map[string]map[string][]byte{}}
	for _, receipt := range receipts {
		if receipt.Identity != r.Identity || receipt.Session == r.Session {
			continue
		}
		if _, ok := ledger.Predecessors[receipt.Session]; ok {
			continue
		}
		if receipt.Session == "" || filepath.Base(receipt.Session) != receipt.Session {
			return os.ErrInvalid
		}
		base := filepath.Join(filepath.Dir(run.Dir()), receipt.Session)
		files, e := minimaxlive.Inventory(base)
		if e != nil {
			return e
		}
		files[minimaxlive.SealFile] = ""
		captures := map[string][]byte{}
		for name := range files {
			raw, e := minimaxlive.ReadFile(filepath.Join(base, filepath.FromSlash(name)), 16<<20)
			if e != nil {
				return e
			}
			captures[name] = raw
		}
		ledger.Predecessors[receipt.Session] = captures
	}
	return run.Record("minimax-ledger", ledger)
}
