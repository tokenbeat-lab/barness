package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, "python3", "tools/verification/run.py", "--feature", "barness-ai-grok-glm-offline", "--stage", "e2e", "--", "go", "test", "-json", "./ai/e2e", "-run", "^TestGrokHistoryGLMControlledFullCapture$", "-count=1", "-timeout=3m")
	cmd.Dir, cmd.Env = root, xaiProcessEnv(t)
	out, err := cmd.CombinedOutput()
	ev.Record("producer-collector", map[string]any{"argv": cmd.Args, "output": string(out), "success": err == nil, "vendorCalls": 0})
	if !ev.Check("actual controlled producer collector completes", err == nil, "%v %s", err, out) {
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
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if output, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, output)
	}
	reportDir := filepath.Join(base, "consumer")
	consume := exec.CommandContext(ctx, binary, "-record", rel(recordPath), "-out", reportDir)
	consume.Dir, consume.Env = root, xaiProcessEnv(t)
	output, consumeErr := consume.CombinedOutput()
	exit := -1
	if consume.ProcessState != nil {
		exit = consume.ProcessState.ExitCode()
	}
	reportRaw, err := os.ReadFile(filepath.Join(reportDir, "release-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report release.RecordReport
	if err = json.Unmarshal(reportRaw, &report); err != nil {
		t.Fatal(err)
	}
	ev.Record("offline-consumer", map[string]any{"argv": consume.Args, "exit": exit, "output": string(output), "report": report})
	checks := map[string]bool{}
	for _, check := range report.Checks {
		checks[check.ID] = check.Pass
	}
	ev.Check("actual release consumer accepts complete offline native inventories", checks["source"] && checks["e2e"], "checks=%+v", report.Checks)
	ev.Check("controlled evidence never supplies missing manual stages", consumeErr != nil && exit == 1 && !report.Pass && !checks["race"], "exit=%d report=%+v", exit, report)
	ev.Check("collector independently registers the original GLM auxiliary bundle", len(receipt.Bundles) == 2, "bundles=%v", receipt.Bundles)
}
