package minimaxlive

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var pendingRenewalRoutes sync.Map

// RenewalPreflight is the launcher's read-only, keyless guard. The launcher
// still owns its exclusive intent before reading a credential; this check
// cannot consume or stand in for the child's process-bound BeginRenewal.
func RenewalPreflight(route string) (RenewalGrant, string, error) {
	return renewalPreflight(FinalAcceptanceRenewalID, route)
}

// RecoveryPreflight checks the fixed recovery without claiming a process nonce.
func RecoveryPreflight(route string) (RenewalGrant, string, error) {
	return renewalPreflight(RecoveryRenewalID, route)
}

func Recovery4Preflight(route string) (RenewalGrant, string, error) {
	return renewalPreflight(Recovery4RenewalID, route)
}

func renewalPreflight(renewal, route string) (RenewalGrant, string, error) {
	g, pin, e := approvedRenewal(renewal, true)
	if e != nil {
		return g, "", e
	}
	id, e := Select(route, M3, "standard")
	if e != nil {
		return g, "", e
	}
	path, e := RenewalLedgerPath(renewal)
	if e != nil {
		return g, "", e
	}
	if _, e = os.Lstat(path + ".lock"); !os.IsNotExist(e) {
		return g, "", fmt.Errorf("final renewal ledger locked or unavailable")
	}
	if _, e = unclaimedRenewalRoute(path, id, pin); e != nil {
		return g, "", e
	}
	return g, pin, nil
}

func unclaimedRenewalRoute(path string, id Identity, pin string) ([]Receipt, error) {
	claim := filepath.Join(filepath.Dir(path), id.Route+".json")
	if _, e := os.Lstat(claim); !os.IsNotExist(e) {
		return nil, fmt.Errorf("final renewal route already claimed or unavailable")
	}
	var rows []Receipt
	if _, e := os.Lstat(path); e == nil {
		var e error
		rows, e = ReadLedger(path)
		if e != nil {
			return nil, e
		}
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	stage := Budget{}
	identities := map[Identity]Budget{}
	for _, r := range rows {
		if r.RenewalID != filepath.Base(filepath.Dir(path)) || r.RenewalGrantSHA256 != pin || r.Stage != "formal" || r.Identity.Model != M3 || r.Identity.Purpose != "standard" || r.Identity == id {
			return nil, fmt.Errorf("final renewal ledger has changed authority or hidden route consumption")
		}
		addBudget(&stage, r)
		b := identities[r.Identity]
		addBudget(&b, r)
		identities[r.Identity] = b
		if b.Calls > 20 || b.Attempts > 24 || b.Attempts > b.ReservedAttempts || b.ReservedAttempts > 24 || b.RequestedOutput > 40000 || b.ReservedInput > 120000 {
			return nil, fmt.Errorf("final renewal identity budget exhausted")
		}
	}
	if stage.Calls >= 60 || stage.Attempts >= 72 || stage.ReservedAttempts >= 72 || len(identities) >= 3 {
		return nil, fmt.Errorf("final renewal total budget exhausted")
	}
	return rows, nil
}

// BeginRenewal claims one approved canonical route before the Host reads its
// credential. No Logical Call is invented. A failed or crashed begin remains
// consumed; only this process may write the first reservation into its ledger.
func (l Ledger) BeginRenewal() error {
	renewal := os.Getenv("BARNESS_MINIMAX_FORMAL_RENEWAL")
	if renewal != FinalAcceptanceRenewalID && renewal != RecoveryRenewalID && renewal != Recovery4RenewalID && renewal != Recovery5RenewalID && renewal != Recovery6RenewalID {
		return fmt.Errorf("only current fixed final renewal can begin")
	}
	id, e := Select(l.Identity.Route, M3, "standard")
	if e != nil || id != l.Identity {
		return fmt.Errorf("final renewal requires exact international standard M3 identity")
	}
	_, pin, e := l.renewalAuthority("full-stream")
	if e != nil {
		return e
	}
	lock := l.Path + ".lock"
	if e = os.Mkdir(lock, 0700); e != nil {
		return fmt.Errorf("final renewal ledger locked")
	}
	if e = syncLedgerDirectory(filepath.Dir(l.Path)); e != nil {
		return e // Leave the uncertain lock closed.
	}
	defer os.Remove(lock)
	rows, e := unclaimedRenewalRoute(l.Path, l.Identity, pin)
	if e != nil {
		return e
	}
	if e = l.claimRenewal(renewal, pin, rows); e != nil {
		return e
	}
	nonce, e := renewalProcessNonce()
	if e != nil {
		return e
	}
	pendingRenewalRoutes.Store(filepath.Join(filepath.Dir(l.Path), l.Identity.Route+".json"), nonce)
	return nil
}
