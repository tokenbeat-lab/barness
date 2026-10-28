package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Presence is observed from native wire, never inferred from the public zero
// value. Complete and omitted controlled responses are distinct real processes.
func TestMiniMaxThinkingBillingApplicability(t *testing.T) {
	kimiHistoryPrivateStore(t)
	cli := miniMaxAuthorityBuild(t, "processgrant", "./ai/live/cmd/processgrant", false)
	binary := kimiRetainedBinary(t, miniMaxAuthorityBuild(t, "minimax.test", "./ai/live", true), "minimax-billing")
	consumer := miniMaxAuthorityBuild(t, "releasegate", "./ai/release/cmd/releasegate", false)
	for _, script := range []string{"omit-cache-creation", "success-messages", "null-cache-usage", "unknown-cache-write", "cache-creation-ttl", "invalid-cache-creation", "missing-cache-read", "null-input-usage", "missing-final-usage", "missing-terminal"} {
		t.Run(script, func(t *testing.T) {
			ev := run.Case(t, "MMCF-billing-"+script)
			g := mmThinkingTemplate(t, ev, cli, binary, strings.Repeat("a", 64))
			dir, r, cmd := mmThinkingProcess(t, ev, binary, g, script, "")
			positive := script == "omit-cache-creation" || script == "success-messages"
			if !positive {
				ev.Check("out of fixed missing-only scope remains rejected", cmd.ProcessState.ExitCode() == 1 && r.Outcome == "FAIL", "exit=%d outcome=%s failure=%s", cmd.ProcessState.ExitCode(), r.Outcome, r.Failure)
				var seal minimaxlive.Seal
				raw, err := os.ReadFile(filepath.Join(dir, minimaxlive.SealFile))
				if err != nil {
					t.Fatal(err)
				}
				mustUnmarshal(t, raw, &seal)
				files, err := minimaxlive.Inventory(dir)
				ev.Check("negative final report remains sealed", err == nil && reflect.DeepEqual(files, seal.Files), "%v", err)
				return
			}
			if !ev.Check("automatic-cache single succeeds without invented counter", cmd.ProcessState.ExitCode() == 0 && r.Outcome == "PASS" && !r.Complete && r.Budget.Calls == 1 && r.Budget.Attempts == 1, "exit=%d outcome=%s failure=%s", cmd.ProcessState.ExitCode(), r.Outcome, r.Failure) {
				return
			}
			hasCreation := false
			for _, u := range r.Calls[0].UsagePresence {
				var fields map[string]json.RawMessage
				mustUnmarshal(t, u, &fields)
				if _, ok := fields["cache_creation_input_tokens"]; ok {
					hasCreation = true
				}
			}
			var captured minimaxlive.Capture
			raw, err := os.ReadFile(filepath.Join(dir, r.Calls[0].CaseID, "result.json"))
			if err != nil {
				t.Fatal(err)
			}
			mustUnmarshal(t, raw, &captured)
			expectedCounter := "REPORTED_ZERO"
			if script == "omit-cache-creation" {
				expectedCounter = "UNREPORTED"
				ev.Check("raw omission is retained", !hasCreation, "counter fabricated")
				ev.Check("partial remains partial", captured.Result.Metadata.Attempts[0].UsageReporting == ai.UsagePartial, "%+v", captured.Result.Metadata.Attempts[0])
			} else {
				ev.Check("explicit zero remains reported", hasCreation, "counter lost")
			}
			inspect := exec.CommandContext(ctxFor(t), consumer, "-inspect-bundle", dir, "-mode", "controlled", "-stage", "candidate")
			inspect.Dir, inspect.Env = "../..", regionalChildEnvironment()
			output, err := inspect.CombinedOutput()
			ev.Record("inspection", map[string]any{"argv": inspect.Args, "error": errString(err), "output": regionalSafeOutput(output), "vendorCalls": 0})
			if !ev.Check("native single inspection succeeds", err == nil, "%v %s", err, output) {
				return
			}
			var q struct {
				Billing struct {
					IndependentCacheWriteFee, CacheCreationCounter string
					UsageReporting                                 ai.UsageReporting
				}
			}
			mustUnmarshal(t, output, &q)
			ev.Check("write fee applicability does not rewrite measurement", q.Billing.IndependentCacheWriteFee == "NOT_APPLICABLE" && q.Billing.CacheCreationCounter == expectedCounter && q.Billing.UsageReporting == captured.Result.Metadata.Attempts[0].UsageReporting, "%s", output)
			ev.Fixture("inspection.json", output)
			// Even a newly sealed altered synthetic packet cannot introduce explicit
			// cache control into this automatic-cache contract.
			request := filepath.Join(dir, r.Calls[0].CaseID, "request.json")
			prior, err := os.ReadFile(request)
			if err != nil {
				t.Fatal(err)
			}
			var bad map[string]any
			mustUnmarshal(t, prior, &bad)
			bad["cache_control"] = map[string]any{"type": "ephemeral"}
			if err := os.WriteFile(request, mustMarshal(t, bad), 0600); err != nil {
				t.Fatal(err)
			}
			_, policyErr := minimaxlive.AssessThinkingBilling(dir, r)
			ev.Check("billing rule rejects explicit cache control", policyErr != nil && strings.Contains(policyErr.Error(), "cache_control"), "%v", policyErr)
			negative := exec.CommandContext(ctxFor(t), consumer, inspect.Args[1:]...)
			negative.Dir, negative.Env = "../..", regionalChildEnvironment()
			out, err := negative.CombinedOutput()
			ev.Record("cache-control-refusal", map[string]any{"argv": negative.Args, "error": errString(err), "output": regionalSafeOutput(out)})
			ev.Check("explicit cache control never gains automatic fee exemption", err != nil, "%s", out)
			if err := os.WriteFile(request, prior, 0600); err != nil {
				t.Fatal(err)
			}
		})
	}
}
