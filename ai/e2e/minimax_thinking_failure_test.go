package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// An initial transport error, 429, stream interruption, missing terminal or
// positive cache writes may not acquire a retry or a second reservation.
func TestMiniMaxThinkingSingleFailure(t *testing.T) {
	kimiHistoryPrivateStore(t)
	cli := miniMaxAuthorityBuild(t, "processgrant", "./ai/live/cmd/processgrant", false)
	binary := kimiRetainedBinary(t, miniMaxAuthorityBuild(t, "minimax.test", "./ai/live", true), "minimax-failures")
	for _, script := range []string{"success-transient-transport", "permanent-429", "thinking-on-read-error", "missing-thinking", "missing-terminal", "known-input-overrun", "extra-attempt", "unknown-cache-write", "missing-usage", "null-cache-usage", "missing-final-usage"} {
		t.Run(script, func(t *testing.T) {
			ev := run.Case(t, "MMTS-failure-"+script)
			g := mmThinkingTemplate(t, ev, cli, binary, strings.Repeat("a", 64))
			dir, r, cmd := mmThinkingProcess(t, ev, binary, g, script, "")
			if !ev.Check("failure cannot become a qualified single", cmd.ProcessState.ExitCode() == 1 && !r.Complete && r.Outcome == "FAIL" && r.Budget.Calls == 1 && r.Budget.Attempts <= 1 && len(r.Calls) == 1, "exit=%d outcome=%s calls=%d attempts=%d", cmd.ProcessState.ExitCode(), r.Outcome, r.Budget.Calls, r.Budget.Attempts) {
				return
			}
			_, e := minimaxlive.VerifyThinkingSupplement(dir, false)
			ev.Check("native single consumer rejects failure", e != nil, "accepted")
			_, raw, claim, e := livegrant.ReadEvidence(dir)
			if e != nil {
				t.Fatal(e)
			}
			decoded, e := livegrant.Decode(raw)
			if e != nil || livegrant.Verify(decoded, raw, claim) != nil {
				t.Fatal("consumed synthetic authority missing")
			}
			path, e := livegrant.LedgerPath(claim)
			if e != nil {
				t.Fatal(e)
			}
			before, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			ledger := minimaxlive.Ledger{Path: path, Stage: r.Stage, Session: r.Session, Identity: r.Identity, ProcessGrantSHA256: r.ProcessGrantSHA256, Scope: &decoded.Scope}
			_, e = ledger.Reserve("thinking-on", 2048, 2048)
			ev.Check("second reservation refused", e != nil, "accepted")
			_, e = ledger.Attempt("thinking-on")
			ev.Check("second actual attempt refused", e != nil, "accepted")
			after, _ := os.ReadFile(path)
			ev.Check("refused append leaves complete ledger unchanged", string(before) == string(after), "ledger changed")
			ev.Record("consumed-private-ledger", map[string]any{"path": filepath.ToSlash(path), "sha256": livegrant.Hash(after), "grantId": decoded.ID, "vendorCalls": 0})
		})
	}
}
