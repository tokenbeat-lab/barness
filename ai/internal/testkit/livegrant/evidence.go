package livegrant

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func ReadEvidence(dir string) (Grant, []byte, Receipt, error) {
	var r Receipt
	root, e := os.OpenRoot(dir)
	if e != nil {
		return Grant{}, nil, r, e
	}
	defer root.Close()
	raw, e := evidenceFile(root, "process-grant.json")
	if e != nil || len(raw) > 1<<20 {
		return Grant{}, nil, r, fmt.Errorf("fresh process approval evidence absent")
	}
	g, e := Decode(raw)
	if e != nil {
		return g, raw, r, e
	}
	var manifest struct{ Versions map[string]string }
	mraw, me := evidenceFile(root, "manifest.json")
	if me != nil || json.Unmarshal(mraw, &manifest) != nil || manifest.Versions["process_build_source"] != g.Scope.SourceSHA256 {
		return g, raw, r, fmt.Errorf("producer binary source attestation absent or differs")
	}
	claim, e := evidenceFile(root, "process-authorization-receipt.json")
	if e != nil || len(claim) > 1<<20 || json.Unmarshal(claim, &r) != nil {
		return g, raw, r, fmt.Errorf("fresh process claim evidence absent")
	}
	return g, raw, r, nil
}

func evidenceFile(root *os.Root, name string) ([]byte, error) {
	info, e := root.Lstat(name)
	if e != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return nil, fmt.Errorf("process evidence must be regular and bounded")
	}
	f, e := root.OpenFile(name, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	opened, e := f.Stat()
	if e != nil || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("process evidence changed during open")
	}
	raw, e := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if e != nil || len(raw) > 1<<20 {
		return nil, fmt.Errorf("process evidence read exceeds bound")
	}
	return raw, nil
}

func VerifyEvidence(g Grant, raw []byte, r Receipt, want Scope, account, region string, started, finished time.Time) error {
	if e := g.Check(want, started); e != nil {
		return e
	}
	if g.AccountAlias != account || g.RegionAlias != region || r.StartedAt.Before(started) || r.StartedAt.After(finished) || finished.Sub(started) > 1800*time.Second || !finished.Before(g.ExpiresAt) {
		return fmt.Errorf("fresh process claim account, region or time differs")
	}
	return Verify(g, raw, r)
}

// LedgerPath is a fresh MiniMax process's own ledger under its already consumed
// claim. It neither selects nor resets any historical stage/renewal ledger.
func LedgerPath(r Receipt) (string, error) {
	p, e := store()
	if e != nil {
		return "", e
	}
	if len(r.ScopeSHA256) != 64 {
		return "", fmt.Errorf("process claim scope missing")
	}
	return filepath.Join(p, "ledger-"+r.ScopeSHA256+".jsonl"), nil
}
