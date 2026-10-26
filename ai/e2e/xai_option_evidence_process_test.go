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

// Failure inventory: the actual option producer must emit a bundle accepted
// by the native read-only gate, including omitted effort between path segments.
// Pending paid/human stages must still refuse release after source/E2E pass.
func TestXAIOptionEvidenceConsumption(t *testing.T) {
	ev := run.Case(t, "XAI-E04-option-evidence-consumer")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	exitCode := func(err error) int {
		if err == nil {
			return 0
		}
		if ex, ok := err.(*exec.ExitError); ok {
			return ex.ExitCode()
		}
		return -1
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	producer := exec.CommandContext(ctx, "python3", "tools/verification/run.py", "--feature", "barness-ai-grok", "--stage", "e2e", "--", "go", "test", "-json", "./ai/e2e", "-run", "^TestXAIReasoningAndBudget$", "-count=1", "-timeout=2m")
	producer.Dir, producer.Env = root, xaiProcessEnv(t)
	output, err := producer.CombinedOutput()
	ev.Record("producer-process", map[string]any{"argv": producer.Args, "exitCode": exitCode(err), "output": string(output)})
	if !ev.Check("bounded real option producer passes", err == nil, "%v", err) {
		return
	}
	var collected struct {
		Attempt  string `json:"attempt"`
		Status   string `json:"status"`
		ExitCode int    `json:"exitCode"`
	}
	if err := json.Unmarshal(output, &collected); err != nil || collected.Status != "PASS" || collected.ExitCode != 0 {
		t.Fatal("producer receipt unavailable")
	}
	receiptPath := filepath.Join(root, collected.Attempt, "receipt.json")
	receiptRaw, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var receipt release.StageReceipt
	if err := json.Unmarshal(receiptRaw, &receipt); err != nil || len(receipt.Bundles) != 1 {
		t.Fatal("one native producer bundle required")
	}
	ev.Record("producer-receipt", map[string]any{"path": filepath.ToSlash(filepath.Join(collected.Attempt, "receipt.json")), "sha256": evidence.Digest(receiptRaw)})
	bundle := filepath.Join(root, collected.Attempt, receipt.Bundles[0].Path)
	manifestRaw, err := os.ReadFile(filepath.Join(bundle, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("producer-manifest.json", manifestRaw)
	binary := filepath.Join(t.TempDir(), "releasegate")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	output, err = build.CombinedOutput()
	ev.Record("consumer-build", map[string]any{"argv": build.Args, "exitCode": exitCode(err), "output": string(output)})
	if err != nil {
		t.Fatal("native consumer build failed")
	}
	binaryRaw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("consumer-binary", map[string]any{"sha256": evidence.Digest(binaryRaw), "bytes": len(binaryRaw)})
	state := []byte("Controlled option producer evidence; real live, race and pressure remain pending.\n")
	statePath := filepath.Join(run.Dir(), "xai-option-consumer-state.md")
	if err := os.WriteFile(statePath, state, 0600); err != nil {
		t.Fatal(err)
	}
	relativeState, err := filepath.Rel(root, statePath)
	if err != nil {
		t.Fatal(err)
	}
	pressureRequired := true
	record := release.Record{Schema: 1, Feature: "barness-ai-grok", CurrentState: release.FileRef{Path: filepath.ToSlash(relativeState), SHA256: evidence.Digest(state)}, Source: receipt.Source, SelectedInputs: receipt.SelectedInputs,
		Impact: release.Impact{LiveRoutes: []string{"xai-responses"}, PressureRequired: &pressureRequired, Reason: "Controlled producer/consumer regression; pending feature stages remain required"},
		E2E:    release.StageRecord{Status: "PASS", Receipt: release.FileRef{Path: filepath.ToSlash(filepath.Join(collected.Attempt, "receipt.json")), SHA256: evidence.Digest(receiptRaw)}, RequiredCases: []string{"XAI-E04-mapped-options-full/default/0", "XAI-E04-mapped-options-simple/default/128000"}},
		Live:   release.StageRecord{Status: "NOT_RUN"}, Race: release.StageRecord{Status: "NOT_RUN"}, Pressure: release.StageRecord{Status: "NOT_RUN"}}
	writeReleaseJSON(t, run.Dir(), "xai-option-consumer-record.json", record)
	out := filepath.Join(run.Dir(), "xai-option-native-consumer")
	consume := exec.CommandContext(ctx, binary, "-record", filepath.Join(run.Dir(), "xai-option-consumer-record.json"), "-out", out)
	consume.Dir, consume.Env = root, xaiProcessEnv(t)
	output, err = consume.CombinedOutput()
	ev.Record("native-consumer-process", map[string]any{"argv": consume.Args, "exitCode": exitCode(err), "output": string(output)})
	reportRaw, readErr := os.ReadFile(filepath.Join(out, "release-report.json"))
	if readErr != nil {
		t.Fatal("native report missing")
	}
	ev.Fixture("native-release-report.json", reportRaw)
	var report release.RecordReport
	if err := json.Unmarshal(reportRaw, &report); err != nil {
		t.Fatal(err)
	}
	checks := map[string]bool{}
	for _, check := range report.Checks {
		checks[check.ID] = check.Pass
	}
	ev.Check("native gate accepts actual source and E2E producer bundle", checks["source"] && checks["e2e"], "%s", reportRaw)
	ev.Check("unpaid and human stages still prevent release", exitCode(err) == 1 && !report.Pass && len(checks) == 5 && !checks["live"] && !checks["race"] && !checks["pressure"], "%s", reportRaw)
}
