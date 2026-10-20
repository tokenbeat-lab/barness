package minimaxlive

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Budget records reservation, transport consumption and known usage separately.
// Requested tokens are conservative Host bounds, not a vendor billing ceiling.
type Budget struct {
	Calls            int `json:"calls"`
	ReservedAttempts int `json:"reservedAttempts"`
	Attempts         int `json:"attempts"`
	RequestedOutput  int `json:"requestedOutput"`
	ReservedInput    int `json:"reservedInput"`
}
type Receipt struct {
	RenewalID          string    `json:"renewalId,omitempty"`
	RenewalGrantSHA256 string    `json:"renewalGrantSha256,omitempty"`
	Session            string    `json:"session"`
	Sequence           int       `json:"sequence"`
	Stage              string    `json:"stage"`
	Identity           Identity  `json:"identity"`
	Case               string    `json:"case"`
	Kind               string    `json:"kind"`
	Output             int       `json:"output"`
	Input              int       `json:"input"`
	At                 time.Time `json:"at"`
	Previous           string    `json:"previous"`
	Hash               string    `json:"hash"`
}
type Ledger struct {
	Path, Stage, Session string
	Identity             Identity
}

func Hash(raw []byte) string       { s := sha256.Sum256(raw); return hex.EncodeToString(s[:]) }
func receiptHash(r Receipt) string { r.Hash = ""; b, _ := json.Marshal(r); return Hash(b) }
func (l Ledger) Reserve(caseID string, output, input int) (Receipt, error) {
	if output < 1 || output > 4096 || input < 1 || input > 120000 {
		return Receipt{}, fmt.Errorf("invalid MiniMax per-call reservation")
	}
	return l.append(caseID, "reserve", output, input)
}
func (l Ledger) Attempt(caseID string) (Receipt, error) { return l.append(caseID, "attempt", 0, 0) }
func (l Ledger) append(caseID, kind string, output, input int) (Receipt, error) {
	if l.Stage != "candidate" && l.Stage != "formal" {
		return Receipt{}, fmt.Errorf("invalid MiniMax stage")
	}
	selected, e := Select(l.Identity.Route, l.Identity.Model, l.Identity.Purpose)
	if e != nil || selected != l.Identity || l.Stage == "formal" && selected.Model != M3 {
		return Receipt{}, fmt.Errorf("invalid stage identity")
	}
	renewal, grantPin, e := l.renewalAuthority(caseID)
	if e != nil {
		return Receipt{}, e
	}
	if caseID == "" {
		return Receipt{}, fmt.Errorf("empty case")
	}
	if e = os.MkdirAll(filepath.Dir(l.Path), 0700); e != nil {
		return Receipt{}, e
	}
	// A failed/crashed writer leaves the lock in place. Fail closed; manual
	// recovery requires inspecting existing receipts, never starting a fresh file.
	lock := l.Path + ".lock"
	if e = os.Mkdir(lock, 0700); e != nil {
		return Receipt{}, fmt.Errorf("MiniMax stage ledger locked")
	}
	if renewal != "" {
		if e = syncLedgerDirectory(filepath.Dir(l.Path)); e != nil {
			return Receipt{}, e
		}
	}
	defer os.Remove(lock)
	var receipts []Receipt
	if _, e = os.Lstat(l.Path); e == nil {
		receipts, e = ReadLedger(l.Path)
	} else if os.IsNotExist(e) {
		e = nil
	}
	if e != nil {
		return Receipt{}, e
	}
	stage, own := Budget{}, Budget{}
	seen := map[string]bool{}
	perCase := map[string]int{}
	reserved := 0
	attempted := 0
	for _, r := range receipts {
		if r.Stage != l.Stage || r.RenewalID != renewal || r.RenewalGrantSHA256 != grantPin {
			return Receipt{}, fmt.Errorf("mixed stage ledger")
		}
		seen[r.Identity.BudgetSlot()] = true
		addBudget(&stage, r)
		if r.Identity.BudgetSlot() == l.Identity.BudgetSlot() {
			addBudget(&own, r)
			if r.Identity == l.Identity && r.Case == caseID {
				if r.Kind == "reserve" {
					perCase[caseID]++
					reserved++
				} else {
					attempted++
				}
			}
		}
	}
	maxCalls, maxAttempts, maxIDs := 120, 144, 6
	if l.Stage == "formal" {
		maxCalls, maxAttempts, maxIDs = 60, 72, 3
	}
	seen[l.Identity.BudgetSlot()] = true
	if len(seen) > maxIDs {
		return Receipt{}, fmt.Errorf("stage identity budget exhausted")
	}
	r := Receipt{RenewalID: renewal, RenewalGrantSHA256: grantPin, Session: l.Session, Sequence: len(receipts) + 1, Stage: l.Stage, Identity: l.Identity, Case: caseID, Kind: kind, Output: output, Input: input, At: time.Now().UTC()}
	if len(receipts) > 0 {
		r.Previous = receipts[len(receipts)-1].Hash
	}
	addBudget(&stage, r)
	addBudget(&own, r)
	if stage.Calls > maxCalls || stage.Attempts > maxAttempts || stage.ReservedAttempts > maxAttempts || own.Calls > 20 || own.Attempts > 24 || own.Attempts > own.ReservedAttempts || own.ReservedAttempts > 24 || own.RequestedOutput > 40000 || own.ReservedInput > 120000 {
		return Receipt{}, fmt.Errorf("MiniMax cumulative stage/identity budget exhausted")
	}
	if kind == "reserve" && (perCase[caseID] > 0 && (!RerunnableCase(caseID) || perCase[caseID] > 1)) {
		return Receipt{}, fmt.Errorf("unregistered repeated scenario")
	}
	if kind == "attempt" && attempted >= reserved {
		return Receipt{}, fmt.Errorf("unreserved actual attempt")
	}
	if renewal != "" {
		if e = l.claimRenewal(renewal, grantPin, receipts); e != nil {
			return Receipt{}, e
		}
	}
	r.Hash = receiptHash(r)
	f, e := os.OpenFile(l.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return Receipt{}, e
	}
	raw, _ := json.Marshal(r)
	raw = append(raw, '\n')
	_, e = f.Write(raw)
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e == nil {
		e = closeErr
	}
	if e == nil && renewal != "" {
		e = syncLedgerDirectory(filepath.Dir(l.Path))
	}
	return r, e
}
func addBudget(b *Budget, r Receipt) {
	if r.Kind == "reserve" {
		b.Calls++
		b.ReservedAttempts++
		b.RequestedOutput += r.Output
		b.ReservedInput += r.Input
	} else {
		b.Attempts++
	}
}
func ReadLedger(path string) ([]Receipt, error) {
	raw, e := ReadFile(path, 4<<20)
	if e != nil {
		return nil, e
	}
	var out []Receipt
	previous := ""
	for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		if line == "" {
			continue
		}
		var r Receipt
		if e = StrictJSON([]byte(line), &r); e != nil || r.Sequence != len(out)+1 || r.Previous != previous || r.Hash != receiptHash(r) || (r.Kind != "reserve" && r.Kind != "attempt") {
			return nil, fmt.Errorf("invalid MiniMax ledger receipt chain")
		}
		if !accountingIdentity(r) {
			return nil, fmt.Errorf("invalid ledger identity")
		}
		previous = r.Hash
		out = append(out, r)
	}
	return out, nil
}

// ReadFile admits and reads the same descriptor, bounded even if the file
// grows after admission. Lstat/SameFile preserve the no-symlink contract.
func ReadFile(path string, max int64) ([]byte, error) {
	if max < 0 || max > 256<<20 {
		return nil, fmt.Errorf("invalid artifact read budget")
	}
	before, e := os.Lstat(path)
	if e != nil {
		return nil, e
	}
	if !before.Mode().IsRegular() {
		return nil, fmt.Errorf("unsafe artifact")
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !info.Mode().IsRegular() || !os.SameFile(before, info) || info.Size() > max {
		return nil, fmt.Errorf("unsafe or oversized artifact")
	}
	raw, e := io.ReadAll(io.LimitReader(f, max+1))
	if e != nil {
		return nil, e
	}
	if int64(len(raw)) > max {
		return nil, fmt.Errorf("artifact grew beyond read budget")
	}
	return raw, nil
}
