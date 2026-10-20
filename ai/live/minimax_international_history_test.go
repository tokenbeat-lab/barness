//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func TestMiniMaxInternationalHistoricalAccounting(t *testing.T) {
	cs := run.Case(t, "MM11-global-research-candidate-history")
	raw, e := os.ReadFile("../e2e/testdata/minimax-international/cn-candidate-prefix.jsonl")
	if e != nil {
		t.Fatal(e)
	}
	base := t.TempDir()
	path := filepath.Join(base, "candidate.jsonl")
	if e = os.WriteFile(path, raw, 0600); e != nil {
		t.Fatal(e)
	}
	exe, _ := os.Executable()
	env := []string{"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=chat", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-minimax", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-global-account", "BARNESS_MINIMAX_REGION_ALIAS=global-test", "BARNESS_MINIMAX_CONTROLLED=success-chat", "BARNESS_MINIMAX_LEDGER=" + path, "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "research"), "BARNESS_MINIMAX_RESEARCH=1"}
	execute := func(entry string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=35s")
		cmd.Env = env
		out, e := cmd.CombinedOutput()
		cs.Record(entry, map[string]any{"success": e == nil, "output": string(out)})
		if !cs.Check("current global process succeeds "+entry, e == nil, "%v %s", e, out) {
			t.FailNow()
		}
	}
	execute("^TestMiniMaxResearch$")
	env[10] = "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "candidate")
	env[11] = "BARNESS_MINIMAX_RESEARCH=0"
	env = append(env, "BARNESS_MINIMAX_PREDECESSOR_ROOT="+filepath.Join(base, "research"))
	execute("^TestMiniMaxCandidate$")
	reports, _ := filepath.Glob(filepath.Join(base, "candidate", "*", minimaxlive.ReportFile))
	if len(reports) != 1 {
		t.Fatal("current own report absent")
	}
	dir := filepath.Dir(reports[0])
	r, e := minimaxlive.Verify(dir, false)
	cs.Check("own inventory and shared slot are distinct", e == nil && r.Identity.Endpoint == "https://api.minimax.io/v1" && r.Budget.Calls == 13 && r.Budget.Attempts == 12 && r.Cumulative.Calls == 18 && r.Cumulative.Attempts == 17, "%v own=%+v cumulative=%+v", e, r.Budget, r.Cumulative)
	evidenceRaw, e := os.ReadFile(filepath.Join(dir, "minimax-ledger.json"))
	if e != nil {
		t.Fatal(e)
	}
	var ledger minimaxlive.LedgerEvidence
	json.Unmarshal(evidenceRaw, &ledger)
	cs.Check("only current own research qualifies as predecessor", len(ledger.Predecessors) == 1, "predecessors=%d", len(ledger.Predecessors))
	for _, captures := range ledger.Predecessors {
		var p minimaxlive.Report
		json.Unmarshal(captures[minimaxlive.ReportFile], &p)
		cs.Check("CN consumption cannot supply native proof", p.Identity == r.Identity && p.Budget.Calls == 3, "%+v", p.Identity)
	}
	current, _ := os.ReadFile(path)
	cs.Check("old receipt bytes remain unchanged", bytes.HasPrefix(current, raw), "historical prefix changed")
	archive, e := archiveMiniMaxProcess(dir)
	if cs.Check("complete current and predecessor captures retained", e == nil, "%v", e) {
		cs.Fixture("global-with-cn-accounting.tar.gz", archive)
	}
	cs.Record("slot-budget", r.Cumulative)
}
