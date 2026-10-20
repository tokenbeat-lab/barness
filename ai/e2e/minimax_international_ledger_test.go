package e2e

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Auxiliary public ledger boundary; failure inventory is in ticket-11 before
// implementation. The fixed historical prefix is accounting, never live proof.
func TestMiniMaxInternationalHistoricalBudget(t *testing.T) {
	ev := run.Case(t, "MM11-historical-budget")
	raw, e := os.ReadFile("testdata/minimax-international/cn-candidate-prefix.jsonl")
	if e != nil {
		t.Fatal(e)
	}
	ev.Check("literal historical prefix retained", len(raw) == 1804 && minimaxlive.Hash(raw) == "222c68c07ce06819b6c84f6aa739b0a08824f1b70893bae7accd014bcae2b4f8", "unexpected historical input")
	path := filepath.Join(t.TempDir(), "candidate.jsonl")
	if e = os.WriteFile(path, raw, 0600); e != nil {
		t.Fatal(e)
	}
	id, e := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	if e != nil {
		t.Fatal(e)
	}
	ev.Check("current selection is international only", id.Endpoint == "https://api.minimax.io/v1", "%+v", id)
	prior, e := minimaxlive.ReadLedger(path)
	if e != nil {
		t.Fatal(e)
	}
	b := minimaxlive.Summarize(prior, id)
	ev.Check("old region still consumes current finite slot", b.Calls == 2 && b.Attempts == 2 && b.RequestedOutput == 5120 && b.ReservedInput == 4096, "%+v", b)
	ledger := minimaxlive.Ledger{Path: path, Stage: "candidate", Session: "global-research", Identity: id}
	_, e = ledger.Reserve("research-full-stream", 1024, 2048)
	ev.Check("same scene name in new exact target is not duplicate", e == nil, "%v", e)
	if e != nil {
		return
	}
	_, e = ledger.Attempt("research-full-stream")
	ev.Check("new target attempt pairs only its reservation", e == nil, "%v", e)
	for n := 0; n < 17; n++ {
		_, e = ledger.Reserve(fmt.Sprintf("global-%d", n), 1024, 2048)
		if e != nil {
			t.Fatal(e)
		}
	}
	_, e = ledger.Reserve("region-reset-bypass", 1024, 2048)
	ev.Check("historical two leave only eighteen current calls", e != nil, "new region reset20-call slot")
	after, e := minimaxlive.ReadLedger(path)
	if e != nil {
		t.Fatal(e)
	}
	b = minimaxlive.Summarize(after, id)
	ev.Check("all consumption counted once", b.Calls == 20 && b.Attempts == 3, "%+v", b)
	retained, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	ev.Check("historical bytes not relabeled or rewritten", bytes.HasPrefix(retained, raw), "prefix changed")
	old := ledger
	old.Identity = prior[0].Identity
	_, e = old.Reserve("forbidden-current-cn", 1, 1)
	ev.Check("historical decoder grants no current request authority", e != nil, "CN identity was callable")
	ev.Record("ledger", after)
	ev.Record("budget", b)
}

func TestMiniMaxInternationalHistoricalLedgerSchema(t *testing.T) {
	original, e := os.ReadFile("testdata/minimax-international/cn-candidate-prefix.jsonl")
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"extra-field", "duplicate-field", "noncanonical-field"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "MM11-historical-schema-"+mode)
			replacement := `"sequence":1,"unregistered":true`
			if mode == "duplicate-field" {
				replacement = `"sequence":1,"sequence":1`
			}
			if mode == "noncanonical-field" {
				replacement = `"Sequence":1`
			}
			raw := bytes.Replace(original, []byte(`"sequence":1`), []byte(replacement), 1)
			path := filepath.Join(t.TempDir(), "candidate.jsonl")
			if e := os.WriteFile(path, raw, 0600); e != nil {
				t.Fatal(e)
			}
			ev.Fixture("untrusted-ledger.jsonl", raw)
			_, e := minimaxlive.ReadLedger(path)
			ev.Check("archival accounting accepts no editable schema variants", e != nil, "accepted %s", mode)
			ev.Record("rejected", e != nil)
		})
	}
}
