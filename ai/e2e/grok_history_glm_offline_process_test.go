package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Exercise the real producer/collector/compiled-record consumer seam. Only the
// offline check can pass; this controlled fixture supplies no race authority.
func TestGrokHistoryGLMOfflineRecordProcess(t *testing.T) {
	ev := run.Case(t, "GROK-HISTORY-GLM-OFFLINE-RECORD")
	ctx := glmProcessContext(t, 8*time.Minute)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	producerCtx, producerCancel := context.WithTimeout(ctx, 6*time.Minute)
	defer producerCancel()
	// Propagate a shorter caller deadline into the nested Go test as well;
	// otherwise an outer timeout could kill its capture owner mid-finalization.
	innerBudget := 5 * time.Minute
	deadline, _ := producerCtx.Deadline()
	if available := time.Until(deadline) - 2*glmProcessCleanup; available < innerBudget {
		innerBudget = available
	}
	if !ev.Check("parent deadline leaves bounded child and cleanup time", innerBudget > 0, "remaining=%s", innerBudget) {
		return
	}
	producer := glmCollect(t, producerCtx, root, xaiProcessEnv(t), "python3", "tools/verification/run.py", "--feature", "barness-ai-grok-glm-offline", "--stage", "e2e", "--timeout", "360", "--", "go", "test", "-json", "./ai/e2e", "-run", "^TestGrokHistoryGLMControlledFullCapture$", "-count=1", "-timeout="+innerBudget.String())
	out := []byte(producer.Stdout)
	ev.Record("producer-collector", producer)
	if !ev.Check("actual controlled producer collector completes", producer.success(), "%s %s", producer.Status, regionalSafeOutput(producer.output())) {
		return
	}
	var result struct {
		Attempt, Status string
		ExitCode        int
	}
	mustUnmarshal(t, out, &result)
	receiptFile := filepath.Join(root, result.Attempt, "receipt.json")
	raw, err := os.ReadFile(receiptFile)
	if err != nil {
		t.Fatal(err)
	}
	var receipt release.StageReceipt
	mustUnmarshal(t, raw, &receipt)
	base, err := os.MkdirTemp(run.Dir(), "glm-offline-record-")
	if err != nil {
		t.Fatal(err)
	}
	rel := func(path string) string {
		v, e := filepath.Rel(root, path)
		if e != nil {
			t.Fatal(e)
		}
		return filepath.ToSlash(v)
	}
	state := []byte("Synthetic controlled GLM inventory regression; no live or race qualification.\n")
	statePath := filepath.Join(base, "state.md")
	if err = os.WriteFile(statePath, state, 0600); err != nil {
		t.Fatal(err)
	}
	pressure := false
	record := release.Record{Schema: 1, Feature: receipt.Feature, CurrentState: release.FileRef{Path: rel(statePath), SHA256: evidence.Digest(state)}, Source: receipt.Source, SelectedInputs: receipt.SelectedInputs, Impact: release.Impact{LiveRoutes: []string{}, PressureRequired: &pressure, Reason: "Controlled producer and read-only offline import only."}, E2E: release.StageRecord{Status: "PASS", Receipt: release.FileRef{Path: rel(receiptFile), SHA256: evidence.Digest(raw)}, RequiredCases: []string{"GROK-HISTORY-GLM-CONTROLLED-FULL"}}, Live: release.StageRecord{Status: "NOT_APPLICABLE", Reason: "All child requests are controlled."}, Race: release.StageRecord{Status: "NOT_RUN"}, Pressure: release.StageRecord{Status: "NOT_APPLICABLE", Reason: "Read-only bounded inventory import."}}
	recordPath := filepath.Join(base, "record.json")
	if err = os.WriteFile(recordPath, mustMarshal(t, record), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "releasegate")
	buildCtx, buildCancel := context.WithTimeout(ctx, time.Minute)
	defer buildCancel()
	build := glmCapture(t, buildCtx, root, xaiProcessEnv(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	ev.Record("consumer-build", build)
	if !build.success() {
		t.Fatalf("build: %s %s", build.Status, regionalSafeOutput(build.output()))
	}
	reportDir := filepath.Join(base, "consumer")
	consumeCtx, consumeCancel := context.WithTimeout(ctx, 30*time.Second)
	defer consumeCancel()
	consume := glmCapture(t, consumeCtx, root, xaiProcessEnv(t), binary, "-record", rel(recordPath), "-out", reportDir)
	exit := -1
	if consume.ExitCode != nil {
		exit = *consume.ExitCode
	}
	reportRaw, err := os.ReadFile(filepath.Join(reportDir, "release-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report release.RecordReport
	if err = json.Unmarshal(reportRaw, &report); err != nil {
		t.Fatal(err)
	}
	ev.Record("offline-consumer", map[string]any{"process": consume, "report": report})
	checks := map[string]bool{}
	for _, check := range report.Checks {
		checks[check.ID] = check.Pass
	}
	ev.Check("actual release consumer accepts complete offline native inventories", checks["source"] && checks["e2e"], "checks=%+v", report.Checks)
	ev.Check("controlled evidence never supplies missing manual stages", consume.Status == "FAIL" && consume.ContextError == "" && exit == 1 && !report.Pass && !checks["race"], "exit=%d report=%+v", exit, report)
	ev.Check("collector independently registers the original GLM auxiliary bundle", len(receipt.Bundles) == 2, "bundles=%v", receipt.Bundles)
}
