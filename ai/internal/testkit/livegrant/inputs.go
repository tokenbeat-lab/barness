package livegrant

import (
	"fmt"
	"path/filepath"
)

// EvidenceInputs returns hashes from the already verified trusted common Git
// store. Callers never supply an alternate store or revive a consumed approval.
func EvidenceInputs(dir string) (map[string]string, error) {
	g, raw, r, e := ReadEvidence(dir)
	if e != nil {
		return nil, e
	}
	if e = Verify(g, raw, r); e != nil {
		return nil, e
	}
	ledger, e := LedgerPath(r)
	if e != nil {
		return nil, e
	}
	out := map[string]string{}
	for _, name := range []string{"id-" + Hash([]byte(g.ID)) + ".json", "scope-" + r.ScopeSHA256 + ".json", filepath.Base(ledger)} {
		p := filepath.Join(filepath.Dir(ledger), name)
		h, e := FileHash(p)
		if e != nil {
			return nil, fmt.Errorf("trusted process authority input missing")
		}
		out[name] = h
	}
	return out, nil
}

// ClaimEvidenceInputs closes over the two authoritative one-shot claim keys.
// Routes without a billing ledger use this after their own native validation;
// EvidenceInputs continues to require the MiniMax process ledger as before.
func ClaimEvidenceInputs(dir string) (map[string]string, error) {
	g, raw, r, e := ReadEvidence(dir)
	if e != nil {
		return nil, e
	}
	if e = Verify(g, raw, r); e != nil {
		return nil, e
	}
	base, e := store()
	if e != nil {
		return nil, e
	}
	id, scope := keys(g)
	out := map[string]string{}
	for _, name := range []string{id, scope} {
		h, e := FileHash(filepath.Join(base, name))
		if e != nil {
			return nil, fmt.Errorf("trusted process claim input missing")
		}
		out[name] = h
	}
	return out, nil
}
