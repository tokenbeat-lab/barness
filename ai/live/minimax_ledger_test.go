//go:build live

package live

import (
	"os"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func captureMiniMaxLedger() error {
	receipts, err := minimaxlive.ReadLedger(minimaxLedger.Path)
	if os.IsNotExist(err) {
		receipts, err = nil, nil
	}
	if err != nil {
		return err
	}
	minimaxReport.Cumulative = minimaxlive.Summarize(receipts, minimaxReport.Identity)
	return run.Record("minimax-ledger", minimaxlive.LedgerEvidence{Receipts: receipts})
}
