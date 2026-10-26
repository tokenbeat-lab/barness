package e2e

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestGrokHistoryMiniMaxFreshLedger(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "minimax-live")
	source, e := minimaxlive.ExecutionSource()
	if e != nil {
		t.Fatal(e)
	}
	rawSource := mustMarshal(t, source)
	build := exec.Command("go", "test", "-c", "-tags", "live", "-ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource="+livegrant.Hash(rawSource), "-o", binary, "./ai/live")
	build.Dir, build.Env = "../..", xaiProcessEnv(t)
	if raw, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build fresh producer: %v %s", e, raw)
	}
	binaryHash, e := livegrant.FileHash(binary)
	if e != nil {
		t.Fatal(e)
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-MINIMAX-FRESH-"+route)
			scope, e := supportmatrix.ProcessScope("minimax-"+route, "controlled", "candidate", binaryHash)
			if e != nil {
				t.Fatal(e)
			}
			grant := livegrant.Grant{Schema: 1, ID: "synthetic-minimax-" + route + "-" + evidence.Digest([]byte(run.Dir()))[:16], AccountAlias: "synthetic-account", RegionAlias: "synthetic-region", ExpiresAt: time.Now().UTC().Add(time.Hour), Scope: scope}
			base := filepath.Join(filepath.Dir(run.Dir()), "minimax-fresh-native", filepath.Base(run.Dir()), route)
			approval := filepath.Join(t.TempDir(), "grant.json")
			if e = os.WriteFile(approval, mustMarshal(t, grant), 0600); e != nil {
				t.Fatal(e)
			}
			cmd := miniMaxFreshLedgerCommand(t, binary, route, approval, base)
			raw, e := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "path": cmd.Path, "dir": cmd.Dir, "environmentSHA256": livegrant.Hash([]byte(strings.Join(cmd.Env, "\x00"))), "output": regionalSafeOutput(raw), "success": e == nil, "vendorCalls": 0, "origin": "controlled"})
			if !ev.Check("fresh process executes full fourteen-call native plan", e == nil, "%s", raw) {
				return
			}
			paths, _ := filepath.Glob(filepath.Join(base, "*", minimaxlive.ReportFile))
			if len(paths) != 1 {
				t.Fatalf("native originals=%d", len(paths))
			}
			dir := filepath.Dir(paths[0])
			r, e := minimaxlive.Verify(dir, false)
			ev.Check("independent native consumer verifies new one-process ledger", e == nil && r.Complete && r.Budget.Calls == 14 && r.ProcessGrantSHA256 != "" && r.Cumulative.Calls == 14, "%v", e)
			_, formal := minimaxlive.Verify(dir, true)
			ev.Check("controlled never formal", formal != nil, "formal accepted")
			claimedStore := miniMaxRecordPrivateStore(t, ev, cmd, dir, "claimed")
			originalPID := cmd.ProcessState.Pid()
			// exec.Cmd cannot be reused. Preserve the original private authority
			// namespace rather than allocating another fixture for this restart.
			restart := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
			restart.Args = append([]string(nil), cmd.Args...)
			restart.Dir, restart.Env = cmd.Dir, append([]string(nil), cmd.Env...)
			ev.Check("restart preserves exact command and private environment", restart.Path == cmd.Path && slices.Equal(restart.Args, cmd.Args) && restart.Dir == cmd.Dir && slices.Equal(restart.Env, cmd.Env), "restart command drifted")
			raw, e = restart.CombinedOutput()
			miniMaxCheckRefusedRestart(t, ev, restart, base, dir, originalPID, raw, e)
			restartedStore := miniMaxRecordPrivateStore(t, ev, restart, dir, "refused-restart")
			ev.Check("same private namespace keeps its original one-use guards", maps.Equal(claimedStore, restartedStore), "private store changed during refusal")
			ev.Record("original-native-pointer", dir)
		})
	}
}
