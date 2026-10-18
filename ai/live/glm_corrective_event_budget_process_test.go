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

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMCorrectiveEventBudgetProcess(t *testing.T) {
	cs := run.Case(t, "GLM-corrective-1-event-budget-known-usage")
	base := t.TempDir()
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestGLMFormal$", "-test.timeout=80s")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=messages", "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_KEY=synthetic-glm", "BARNESS_GLM_CONTROLLED=corrective-eventcap", "BARNESS_AI_EVIDENCE_DIR=" + base}
	output, err := cmd.CombinedOutput()
	cs.Record("exit", map[string]any{"success": err == nil, "output": string(output)})
	cs.Check("event evidence overflow remains failure", err != nil, "unexpected success")
	matches, _ := filepath.Glob(filepath.Join(base, "*", "formal-glm-report.json"))
	if !cs.Check("failed report retained", len(matches) == 1, "reports=%d", len(matches)) {
		// A timed-out synthetic child may not reach final report publication.
		// Preserve the partial artifact directory before TempDir cleanup too.
		entries, _ := os.ReadDir(base)
		for _, entry := range entries {
			if entry.IsDir() {
				cs.Check("interrupted child evidence retained", glmRetainCorrectiveChild(cs, filepath.Join(base, entry.Name()), "interrupted-child") == nil, "partial copy failed")
			}
		}
		return
	}
	dir := filepath.Dir(matches[0])
	cs.Check("full bounded failed child retained", glmRetainCorrectiveChild(cs, dir, "failed-child") == nil, "copy failed")
	raw, _ := os.ReadFile(matches[0])
	var report glmReport
	json.Unmarshal(raw, &report)
	cs.Check("no support qualification", report.Outcome == "FAIL" && !report.Complete && report.Budget.CallsUsed <= 24 && report.Budget.HTTPAttempts <= 32, "budget=%+v", report.Budget)
	var result struct {
		Result struct {
			Message  struct{ Usage ai.Usage }
			Metadata ai.CallMetadata
		}
	}
	own := filepath.Join(dir, "GLM11-glm-5.3-full-stream")
	raw, _ = os.ReadFile(filepath.Join(own, "attempt-1-result.json"))
	json.Unmarshal(raw, &result)
	cs.Check("known initial input retained in public failure", result.Result.Message.Usage.Input == 20 && len(result.Result.Metadata.Attempts) == 1, "usage=%+v", result.Result.Message.Usage)
	if len(report.Scenarios) > 0 {
		first := report.Scenarios[0]
		cs.Check("known Attempt usage survives event-evidence refusal in report", first.Usage.Input == 20 && len(first.UsageReporting) == 1 && first.UsageReporting[0] == result.Result.Metadata.Attempts[0].UsageReporting, "usage=%+v reporting=%v", first.Usage, first.UsageReporting)
	}
	var events struct {
		Overflow bool
		Events   []json.RawMessage
	}
	raw, _ = os.ReadFile(filepath.Join(own, "events.json"))
	json.Unmarshal(raw, &events)
	cs.Check("finite event bound explicit", events.Overflow && len(events.Events) > 0 && len(events.Events) <= 8193 && len(raw) < 16<<20, "events=%d bytes=%d", len(events.Events), len(raw))
}
