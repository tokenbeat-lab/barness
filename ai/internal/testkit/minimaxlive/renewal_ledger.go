package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (l Ledger) renewalAuthority(caseID string) (string, string, error) {
	selected := os.Getenv("BARNESS_MINIMAX_FORMAL_RENEWAL")
	if selected == "" {
		if filepath.Base(filepath.Dir(l.Path)) == FormalRenewalID && filepath.Base(l.Path) == "formal.jsonl" {
			return "", "", fmt.Errorf("fixed renewal ledger requires explicit selection")
		}
		return "", "", nil
	}
	canonical, e := RenewalLedgerPath(FormalRenewalID)
	if e != nil {
		return "", "", e
	}
	if selected != FormalRenewalID || l.Stage != "formal" || l.Identity.Model != M3 || l.Path != canonical {
		return "", "", fmt.Errorf("only fixed canonical formal renewal ledger permitted")
	}
	_, pin, e := approvedRenewal(selected, true)
	if e != nil {
		return "", "", e
	}
	registered := false
	for _, c := range Plan(l.Identity) {
		registered = registered || c.ID == caseID
	}
	if !registered {
		return "", "", fmt.Errorf("renewal case is not registered")
	}
	return selected, pin, nil
}

type renewalClaim struct {
	Schema      int      `json:"schema"`
	RenewalID   string   `json:"renewalId"`
	GrantSHA256 string   `json:"grantSha256"`
	Session     string   `json:"session"`
	Identity    Identity `json:"identity"`
}

// An exclusive route claim survives deleting the ledger or choosing another
// process/output directory. A crashed first append remains closed, not fresh.
func (l Ledger) claimRenewal(pin string, rows []Receipt) error {
	if l.Session == "" || filepath.Base(l.Session) != l.Session {
		return fmt.Errorf("renewal session invalid")
	}
	path := filepath.Join(filepath.Dir(l.Path), l.Identity.Route+".json")
	raw, e := ReadFile(path, 16384)
	if e == nil {
		var claim renewalClaim
		if StrictJSON(raw, &claim) != nil || claim != (renewalClaim{1, FormalRenewalID, pin, l.Session, l.Identity}) {
			return fmt.Errorf("renewal identity already claimed or changed")
		}
		found := false
		for _, r := range rows {
			if r.Identity == l.Identity {
				if r.Session != l.Session {
					return fmt.Errorf("renewal identity cannot restart")
				}
				found = true
			}
		}
		if !found {
			return fmt.Errorf("claimed renewal ledger missing")
		}
		return nil
	}
	if !os.IsNotExist(e) {
		return e
	}
	for _, r := range rows {
		if r.Identity == l.Identity {
			return fmt.Errorf("renewal route claim missing")
		}
	}
	raw, e = json.Marshal(renewalClaim{1, FormalRenewalID, pin, l.Session, l.Identity})
	if e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write(raw); e != nil {
		return e
	}
	return f.Sync()
}
