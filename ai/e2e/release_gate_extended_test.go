package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func TestReleaseRaceArtifactContract(t *testing.T) {
	ev := run.Case(t, "OD57-release-race-artifact-contract")
	r := releaseProcess(t, ev)
	for _, name := range []string{"race-offline", "race-differential", "race-mixed-pressure-local", "race-mixed-pressure-cloud", "race-authorization-revocation"} {
		e := releaseEvidence(t, r, name)
		ev.Check("race needs its own "+name, e.Status != release.Pass, "missing race artifact passed")
	}
}
func TestReleaseAllRefusalImport(t *testing.T) {
	ev := run.Case(t, "OD57-release-all-refusal-null-metrics")
	out := t.TempDir()
	_, err := evaluationCommand(t, "-route", "luna", "-fixture", "all-refusal", "-out", out)
	if err != nil {
		t.Fatal(err)
	}
	dir, _ := evaluationBundle(t, out, ev)
	r := releaseProcess(t, ev, "-evaluation-luna", dir)
	e := releaseEvidence(t, r, "chinese-effect-luna")
	ev.Check("complete fixture imported without claiming live", e.Status == release.NotRun && strings.Contains(e.Detail, "fixture/COMPLETE"), "got %s", e.Detail)
	ev.Check("zero and null quality retained without threshold", strings.Contains(e.Detail, "accuracy 0;") && strings.Contains(e.Detail, "conditional accuracy null;") && strings.Contains(e.Detail, "ECE null;") && strings.Contains(e.Detail, "Brier null;"), "got %s", e.Detail)
}
func TestReleaseExecutionSourceDrift(t *testing.T) {
	ev := run.Case(t, "OD57-release-current-source-hash")
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS("../../.scratch/barness-ai-openai-decisions/chinese-live-evidence/live")); err != nil {
		t.Fatal(err)
	}
	read := func(name string) map[string]any {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err = json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	code := read("code-snapshot.json")
	code["ai/client.go"] = "sha256:" + strings.Repeat("0", 64)
	writeReleaseJSON(t, dir, "code-snapshot.json", code)
	raw, _ := os.ReadFile(filepath.Join(dir, "code-snapshot.json"))
	hash := chineseeval.Hash(raw)
	report := read("evaluation-report.json")
	report["codeHash"] = hash
	writeReleaseJSON(t, dir, "evaluation-report.json", report)
	manifest := read("manifest.json")
	manifest["versions"].(map[string]any)["code_hash"] = hash
	writeReleaseJSON(t, dir, "manifest.json", manifest)
	for _, name := range []string{"code-snapshot.json", "evaluation-report.json", "manifest.json"} {
		rehashReleaseFile(t, dir, name)
	}
	r := releaseProcess(t, ev, "-evaluation-luna", dir)
	e := releaseEvidence(t, r, "chinese-effect-luna")
	ev.Check("coherently rehashed stale source rejected", e.Status == release.Fail && strings.Contains(e.Detail, "source"), "got %s", e.Detail)
}
