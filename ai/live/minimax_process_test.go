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

func TestMiniMaxProcess(t *testing.T) {
	for _, name := range []string{"not-run", "switch-off", "missing-entitlement", "narrowed", "foreign-key", "mixed-route", "endpoint-override", "wrong-model", "formal-absent", "success-chat", "success-responses", "success-messages", "success-m31-chat", "success-m31-responses", "success-m31-messages", "missing-thinking", "wrong-actual-model", "missing-terminal", "extra-attempt", "output-budget", "input-budget", "known-input-overrun", "inventory-failure", "research-chat", "success-rerun-tool"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "MM07-process-"+name)
			base := t.TempDir()
			route, model, purpose := "chat", minimaxlive.M3, "standard"
			if strings.HasPrefix(name, "success-") {
				route = strings.TrimPrefix(name, "success-")
				if name == "success-rerun-tool" {
					route = "chat"
				}
			}
			if strings.HasPrefix(route, "m31-") {
				route = strings.TrimPrefix(route, "m31-")
				model, purpose = minimaxlive.M31, "mplan"
			}
			env := []string{"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=" + route, "BARNESS_MINIMAX_MODEL=" + model, "BARNESS_MINIMAX_PURPOSE=" + purpose, "BARNESS_MINIMAX_KEY=synthetic-minimax", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=cn-test", "BARNESS_MINIMAX_CONTROLLED=" + name, "BARNESS_MINIMAX_LEDGER=" + filepath.Join(base, "stage.jsonl"), "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "captures")}
			switch name {
			case "not-run":
				env[4] = "BARNESS_MINIMAX_KEY="
			case "switch-off":
				env[0] = "BARNESS_AI_LIVE=0"
			case "missing-entitlement":
				env[5] = "BARNESS_MINIMAX_ENTITLEMENT="
			case "foreign-key":
				env = append(env, "OPENAI_API_KEY=foreign-synthetic")
			case "mixed-route":
				env = append(env, "BARNESS_GLM_ROUTE=chat")
			case "endpoint-override":
				env = append(env, "MINIMAX_BASE_URL=https://invalid.example")
			case "research-chat":
				env = append(env, "BARNESS_MINIMAX_RESEARCH=1")
			case "wrong-model":
				env[2] = "BARNESS_MINIMAX_MODEL=MiniMax-M2.7"
			case "formal-absent":
				env = append(env, "BARNESS_MINIMAX_FORMAL=1")
			}
			entry := "^TestMiniMaxCandidate$"
			if strings.HasPrefix(name, "research-") {
				entry = "^TestMiniMaxResearch$"
			}
			if name == "narrowed" {
				entry += "/full-stream"
			}
			if name == "formal-absent" {
				entry = "^TestMiniMaxFormal$"
			}
			exe, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=35s")
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			cs.Record("exit", map[string]any{"success": err == nil, "output": string(output)})
			ok := strings.HasPrefix(name, "research-") || strings.HasPrefix(name, "success-") || name == "not-run" || name == "switch-off" || name == "missing-entitlement"
			cs.Check("bounded process exit", (err == nil) == ok, "err=%v output=%s", err, output)
			reports, _ := filepath.Glob(filepath.Join(base, "captures", "*", minimaxlive.ReportFile))
			if !cs.Check("owned report exists", len(reports) == 1, "reports=%d", len(reports)) {
				return
			}
			archive, archiveErr := archiveMiniMaxProcess(filepath.Dir(reports[0]))
			if cs.Check("raw controlled child evidence retained", archiveErr == nil, "%v", archiveErr) {
				cs.Fixture("child-capture.tar.gz", archive)
			}
			raw, _ := os.ReadFile(reports[0])
			cs.Record("child-report", json.RawMessage(raw))
			var r struct {
				Outcome, Mode, Kind string
				Complete            bool
				Budget              minimaxlive.Budget
			}
			json.Unmarshal(raw, &r)
			cs.Check("controlled cannot formal certify", r.Kind != "minimax-formal" || r.Mode != "live", "kind=%s mode=%s", r.Kind, r.Mode)
			formal, _ := filepath.Glob(filepath.Join(filepath.Dir(reports[0]), "live-report.json"))
			cs.Check("no formal report", len(formal) == 0, "formal report was published")
			if strings.HasPrefix(name, "success-") {
				cs.Check("complete own identity", r.Outcome == "PASS" && r.Complete && ((name != "success-rerun-tool" && r.Budget.Calls == 13 && r.Budget.Attempts == 12) || (name == "success-rerun-tool" && r.Budget.Calls == 14 && r.Budget.Attempts == 13)) && r.Budget.RequestedOutput < 40000, "%s", raw)
			} else if strings.HasPrefix(name, "research-") {
				cs.Check("research cannot certify full candidate", r.Outcome == "PASS" && !r.Complete && r.Kind == "minimax-research-only" && r.Budget.Calls == 3 && r.Budget.Attempts == 3, "%s", raw)
			} else if ok {
				cs.Check("NOT_RUN spends zero", r.Outcome == "NOT_RUN" && !r.Complete && r.Budget.Calls == 0 && r.Budget.Attempts == 0, "%s", raw)
			} else {
				cs.Check("failure is retained", r.Outcome == "FAIL" && !r.Complete, "%s", raw)
			}
		})
	}
}
