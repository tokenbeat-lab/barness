//go:build live

package live

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func TestMiniMaxThinkingProbeProcess(t *testing.T) {
	for _, route := range []string{"chat", "responses", "messages"} {
		for _, mode := range []string{"adaptive-probe", "adaptive-probe-empty"} {
			t.Run(route+"/"+mode, func(t *testing.T) {
				cs := run.Case(t, "MM18-thinking-probe-"+route+"-"+mode)
				base := t.TempDir()
				exe, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, exe, "-test.run=^TestMiniMaxCandidate$", "-test.timeout=40s")
				for _, name := range []string{"PATH", "TMPDIR", "GOCACHE", "GOPATH"} {
					if value, ok := os.LookupEnv(name); ok {
						cmd.Env = append(cmd.Env, name+"="+value)
					}
				}
				// No process grant is supplied. The controlled legacy ledger lives
				// entirely under this case's TempDir; no common-Git claim is made.
				cmd.Env = append(cmd.Env, "GOENV=off", "GOFLAGS=", "HOME="+base,
					"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE="+route,
					"BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard",
					"BARNESS_MINIMAX_KEY=synthetic-adaptive-probe", "BARNESS_MINIMAX_ENTITLEMENT=confirmed",
					"BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-probe", "BARNESS_MINIMAX_REGION_ALIAS=synthetic-region",
					"BARNESS_MINIMAX_CONTROLLED="+mode, "BARNESS_MINIMAX_LEDGER="+filepath.Join(base, "ledger.jsonl"),
					"BARNESS_AI_EVIDENCE_DIR="+filepath.Join(base, "bundles"))
				output, processErr := cmd.CombinedOutput()
				cs.Record("process", map[string]any{"argv": cmd.Args, "exitCode": miniMaxProbeExit(processErr), "output": string(output), "vendorCalls": 0, "grantSupplied": false})
				paths, _ := filepath.Glob(filepath.Join(base, "bundles", "*", minimaxlive.ReportFile))
				if !cs.Check("one original native report", len(paths) == 1, "reports=%d", len(paths)) {
					return
				}
				dir := filepath.Dir(paths[0])
				archive, err := archiveMiniMaxProcess(dir)
				if !cs.Check("complete child original retained", err == nil, "%v", err) {
					return
				}
				cs.Fixture("native-child.tar.gz", archive)
				raw, _ := os.ReadFile(paths[0])
				var report minimaxlive.Report
				if json.Unmarshal(raw, &report) != nil {
					t.Fatal("invalid native report")
				}
				cs.Check("no formal or grant identity", report.Mode == "controlled" && report.Kind != "minimax-formal" && report.ProcessGrantSHA256 == "", "controlled proof gained authority")
				cs.Check("original fourteen-scene bounds unchanged", len(report.Expected) == 14 && len(report.Calls) == 14 && report.Budget.Calls == 14 && report.Budget.Attempts == 13, "calls=%d attempts=%d", report.Budget.Calls, report.Budget.Attempts)
				var probe minimaxlive.Call
				for _, call := range report.Calls {
					if call.ID == "thinking-on" {
						probe = call
					} else {
						cs.Check("other scene keeps its own result "+call.ID, call.Outcome == "PASS", "%s", call.Outcome)
					}
				}
				if mode == "adaptive-probe" {
					cs.Check("reasoning task makes own positive probe succeed", processErr == nil && report.Complete && report.Outcome == "PASS" && probe.Outcome == "PASS", "exit=%d own=%s", miniMaxProbeExit(processErr), probe.Outcome)
				} else {
					cs.Check("empty reasoning still fails its own positive probe", miniMaxProbeExit(processErr) == 1 && !report.Complete && report.Outcome == "FAIL" && probe.Outcome == "FAIL" && strings.Contains(string(output), "own reasoning actually observed"), "exit=%d own=%s", miniMaxProbeExit(processErr), probe.Outcome)
				}
				miniMaxProbeRequests(t, cs, dir, route)
			})
		}
	}
}

func miniMaxProbeExit(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}
