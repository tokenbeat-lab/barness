//go:build live

package live

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"os"
)

// The current M3 process has one fresh approval and its own finite ledger.
// Closed stage ledgers/renewals remain historical inputs, never fresh authority.
func initializeFreshMiniMax() error {
	r := minimaxReport
	if os.Getenv("BARNESS_MINIMAX_FORMAL_RENEWAL") != "" || os.Getenv("BARNESS_MINIMAX_LEDGER") != "" {
		return fmt.Errorf("fresh M3 process forbids historical renewal or ledger selection")
	}
	if os.Getenv("BARNESS_MINIMAX_KEY") == "" {
		return nil
	}
	if minimaxConfigError != "" {
		return nil
	}
	a, e := loadProcessAuthorization("minimax-"+r.Identity.Route, r.Mode, r.Stage, r.AccountAlias, r.RegionAlias, os.Getenv("BARNESS_MINIMAX_GRANT"))
	if e != nil {
		return e
	}
	if e = selectMiniMaxProcess(a); e != nil {
		return e
	}
	if e = a.consume(); e != nil {
		return e
	}
	path, e := livegrant.LedgerPath(a.claim)
	if e != nil {
		return e
	}
	r.ProcessGrantSHA256 = livegrant.Hash(a.raw)
	minimaxLedger = minimaxlive.Ledger{Path: path, Stage: r.Stage, Session: r.Session, Identity: r.Identity, ProcessGrantSHA256: r.ProcessGrantSHA256, Scope: &a.grant.Scope}
	if a.grant.ExpiresAt.Before(r.Deadline) {
		r.Deadline = a.grant.ExpiresAt
	}
	return nil
}
