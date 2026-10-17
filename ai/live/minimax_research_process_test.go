//go:build live

package live

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func TestMiniMaxResearchPredecessor(t *testing.T) {
	cs := run.Case(t, "MM07-research-predecessor")
	base := t.TempDir()
	exe, _ := os.Executable()
	env := []string{"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=chat", "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-minimax", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=cn-test", "BARNESS_MINIMAX_CONTROLLED=success-chat", "BARNESS_MINIMAX_LEDGER=" + filepath.Join(base, "stage.jsonl"), "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "research"), "BARNESS_MINIMAX_RESEARCH=1"}
	execute := func(entry string, wantOK bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=35s")
		cmd.Env = env
		output, e := cmd.CombinedOutput()
		cs.Check("child stage behavior "+entry, (e == nil) == wantOK, "err=%v output=%s", e, output)
	}
	execute("^TestMiniMaxResearch$", true)
	env[11] = "BARNESS_MINIMAX_RESEARCH=0"
	env[10] = "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "candidate")
	env = append(env, "BARNESS_MINIMAX_PREDECESSOR_ROOT="+filepath.Join(base, "research"))
	execute("^TestMiniMaxCandidate$", true)
	reports, _ := filepath.Glob(filepath.Join(base, "*", "*", minimaxlive.ReportFile))
	if !cs.Check("separate immutable predecessors", len(reports) == 2, "reports=%d", len(reports)) {
		return
	}
	dir := ""
	for _, path := range reports {
		if filepath.Base(filepath.Dir(filepath.Dir(path))) == "candidate" {
			dir = filepath.Dir(path)
		}
	}
	r, e := minimaxlive.Verify(dir, false)
	cs.Check("full current report accounts prior research", e == nil && r.Budget.Calls == 13 && r.Cumulative.Calls == 16 && r.Cumulative.Attempts == 15, "err=%v budget=%+v cumulative=%+v", e, r.Budget, r.Cumulative)
	raw, _ := os.ReadFile(filepath.Join(dir, "minimax-ledger.json"))
	var ledger minimaxlive.LedgerEvidence
	json.Unmarshal(raw, &ledger)
	cs.Check("own prior captures retained", len(ledger.Predecessors) == 1, "predecessors=%d", len(ledger.Predecessors))
	archive, archiveErr := archiveMiniMaxProcess(dir)
	if cs.Check("full research predecessor captures retained", archiveErr == nil, "%v", archiveErr) {
		cs.Fixture("full-with-research.tar.gz", archive)
	}
	execute("^TestMiniMaxCandidate$", false)
}
