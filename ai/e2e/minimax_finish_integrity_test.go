package e2e

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Reproduce post-verification failures through actual controlled producers.
// Inventory must describe the final FAIL report, for both single and full.
func TestMiniMaxFinalFailureSeal(t *testing.T) {
	kimiHistoryPrivateStore(t)
	cli := miniMaxAuthorityBuild(t, "processgrant", "./ai/live/cmd/processgrant", false)
	binary := kimiRetainedBinary(t, miniMaxAuthorityBuild(t, "minimax.test", "./ai/live", true), "minimax-final-seal")
	for _, kind := range []string{"single", "full"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "MMCF-final-seal-"+kind)
			g := mmThinkingTemplate(t, ev, cli, binary, strings.Repeat("a", 64))
			script := "missing-final-usage"
			if kind == "full" {
				id, _ := minimaxlive.Select("messages", minimaxlive.M3, "standard")
				h, _ := livegrant.FileHash(binary)
				g.Scope, _ = minimaxlive.ProcessScope(id, "controlled", "candidate", h)
				script = "known-input-overrun"
			}
			dir, r, cmd := mmThinkingProcess(t, ev, binary, g, script, "")
			ev.Check("post-verification refusal remains failure", cmd.ProcessState.ExitCode() == 1 && r.Outcome == "FAIL" && !r.Complete && r.Failure != "", "exit=%d outcome=%s failure=%s", cmd.ProcessState.ExitCode(), r.Outcome, r.Failure)
			var seal minimaxlive.Seal
			raw, err := os.ReadFile(filepath.Join(dir, minimaxlive.SealFile))
			if err != nil {
				t.Fatal(err)
			}
			mustUnmarshal(t, raw, &seal)
			files, err := minimaxlive.Inventory(dir)
			ev.Check("seal pins final failure bytes", err == nil && reflect.DeepEqual(files, seal.Files), "inventory error=%v report pin=%s final=%s", err, seal.Files[minimaxlive.ReportFile], files[minimaxlive.ReportFile])
			report, _ := os.ReadFile(filepath.Join(dir, minimaxlive.ReportFile))
			ev.Check("final failure report belongs to seal", seal.Files[minimaxlive.ReportFile] == minimaxlive.Hash(report), "report changed after seal")
			ev.Record("native", map[string]any{"bundle": repoRelative(t, dir), "reportSHA256": minimaxlive.Hash(report), "seal": seal, "exitCode": cmd.ProcessState.ExitCode(), "vendorCalls": 0})
		})
	}
}
