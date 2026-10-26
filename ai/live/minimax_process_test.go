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
	for _, name := range []string{"not-run", "switch-off", "missing-entitlement", "narrowed", "foreign-key", "mixed-route", "endpoint-override", "wrong-model", "formal-absent", "success-chat", "success-responses", "success-messages", "success-m31-chat", "success-m31-responses", "success-m31-messages", "missing-thinking", "wrong-actual-model", "missing-terminal", "extra-attempt", "output-budget", "input-budget", "known-input-overrun", "inventory-failure", "research-chat", "success-rerun-tool", "success-transient-request", "transient-exhausted", "transient-stream", "permanent-429", "unknown-429", "typed-503", "success-transient-transport", "success-one-hour-zero", "tool-auth401", "tool-permanent429", "tool-stream-error"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "MM07-process-"+name)
			base := t.TempDir()
			route, model, purpose := "chat", minimaxlive.M3, "standard"
			if strings.HasPrefix(name, "success-") {
				route = strings.TrimPrefix(name, "success-")
				if name == "success-rerun-tool" || strings.HasPrefix(name, "success-transient-") {
					route = "chat"
				}
			}
			if strings.HasPrefix(route, "m31-") {
				route = strings.TrimPrefix(route, "m31-")
				model, purpose = minimaxlive.M31, "mplan"
			}
			if name == "permanent-429" || name == "unknown-429" || name == "typed-503" || name == "success-one-hour-zero" {
				route = "messages"
			}
			env := []string{"BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE=" + route, "BARNESS_MINIMAX_MODEL=" + model, "BARNESS_MINIMAX_PURPOSE=" + purpose, "BARNESS_MINIMAX_KEY=synthetic-minimax", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=global-test", "BARNESS_MINIMAX_CONTROLLED=" + name, "BARNESS_MINIMAX_LEDGER=" + filepath.Join(base, "stage.jsonl"), "BARNESS_AI_EVIDENCE_DIR=" + filepath.Join(base, "captures")}
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
				Unknown             []string
				Budget              minimaxlive.Budget
				Calls               []minimaxlive.Call
			}
			json.Unmarshal(raw, &r)
			var reportFields map[string]json.RawMessage
			json.Unmarshal(raw, &reportFields)
			cs.Check("cumulative slot budget is labeled separately from proof identity", len(reportFields["cumulativeSlotBudget"]) > 0 && len(reportFields["cumulativeIdentityBudget"]) == 0, "missing slot budget label")
			if model == minimaxlive.M31 {
				cs.Check("raw usage is separate from unknown M Plan pricing", strings.Contains(strings.Join(r.Unknown, "\n"), "M3.1 price unknown; usage is raw provider evidence; Cost is synthetic mechanism output, not M Plan fees"), "missing M3.1 price warning")
				dir := filepath.Dir(reports[0])
				originalSeal, _ := os.ReadFile(filepath.Join(dir, minimaxlive.SealFile))
				var edited minimaxlive.Report
				json.Unmarshal(raw, &edited)
				edited.Unknown = nil
				changed, _ := json.Marshal(edited)
				os.WriteFile(reports[0], changed, 0600)
				files, inventoryErr := minimaxlive.Inventory(dir)
				if inventoryErr != nil {
					t.Fatal(inventoryErr)
				}
				changed, _ = json.Marshal(minimaxlive.Seal{Schema: 1, Files: files})
				os.WriteFile(filepath.Join(dir, minimaxlive.SealFile), changed, 0600)
				_, warningErr := minimaxlive.Verify(dir, false)
				cs.Check("unpriced actual usage cannot lose synthetic-cost context", warningErr != nil && strings.Contains(warningErr.Error(), "price"), "%v", warningErr)
				os.WriteFile(reports[0], raw, 0600)
				os.WriteFile(filepath.Join(dir, minimaxlive.SealFile), originalSeal, 0600)

			}
			if strings.HasPrefix(name, "tool-") {
				count := 0
				for _, call := range r.Calls {
					if call.ID == "tool-call" {
						count++
						cs.Check("error is not a capability-rerun reason", call.Failure != "tool-capability", "failure=%s", call.Failure)
					}
				}
				cs.Check("failed tool scene has no capability rerun", count == 1, "tool calls=%d", count)
				captured, e := os.ReadFile(filepath.Join(filepath.Dir(reports[0]), "MM07-tool-call", "result.json"))
				var failure minimaxlive.Capture
				json.Unmarshal(captured, &failure)
				phase := "request"
				if name == "tool-stream-error" {
					phase = "stream"
					cs.Check("failed stream retains partial thinking", len(failure.Result.Message.Content) > 0, "partial content absent")
				}
				cs.Check("original public error retained", e == nil && failure.ErrorCode != "" && string(failure.ErrorPhase) == phase, "code=%s phase=%s", failure.ErrorCode, failure.ErrorPhase)
			}
			if strings.HasPrefix(name, "success-") {
				_, promotionErr := minimaxlive.PromotionProposal([]string{filepath.Dir(reports[0]), filepath.Dir(reports[0]), filepath.Dir(reports[0])})
				cs.Check("controlled cannot promote Builtin", promotionErr != nil, "controlled proof activated")
			}
			cs.Check("controlled cannot formal certify", r.Kind != "minimax-formal" || r.Mode != "live", "kind=%s mode=%s", r.Kind, r.Mode)
			formal, _ := filepath.Glob(filepath.Join(filepath.Dir(reports[0]), "live-report.json"))
			cs.Check("no formal report", len(formal) == 0, "formal report was published")
			if strings.HasPrefix(name, "success-") {
				selected, _ := minimaxlive.Select(route, model, purpose)
				planned := len(minimaxlive.Plan(selected))
				cs.Check("complete own identity", r.Outcome == "PASS" && r.Complete && ((name != "success-rerun-tool" && !strings.HasPrefix(name, "success-transient-") && r.Budget.Calls == planned && r.Budget.Attempts == planned-1) || ((name == "success-rerun-tool" || strings.HasPrefix(name, "success-transient-")) && r.Budget.Calls == planned+1 && r.Budget.Attempts == planned)) && r.Budget.RequestedOutput < 40000, "%s", raw)
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
