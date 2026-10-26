package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMiniMaxReasoningProbeProcess(t *testing.T) {
	ev := run.Case(t, "MM18-reasoning-probe-public-process")
	// Preserve the native process cases (and their complete child archives) as
	// an auxiliary bundle, rather than flattening them into one bounded case.
	base := filepath.Dir(run.Dir())
	before, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, entry := range before {
		seen[entry.Name()] = true
	}
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "-tags", "live", "./ai/live", "-run", "^TestMiniMaxThinkingProbeProcess$", "-count=1", "-timeout=180s")
	cmd.Dir, cmd.Env = "../..", append(xaiProcessEnv(t), "BARNESS_AI_EVIDENCE_DIR="+base)
	output, processErr := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"argv": cmd.Args, "output": regionalSafeOutput(output), "success": processErr == nil, "vendorCalls": 0, "grantSupplied": false})
	ev.Check("all exact native controlled probes pass", processErr == nil, "%s", output)
	after, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range after {
		if seen[entry.Name()] || !entry.IsDir() {
			continue
		}
		dir := filepath.Join(base, entry.Name())
		raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		if err != nil {
			continue
		}
		var manifest struct {
			Package string
			Cases   []struct{ ID, Status string }
		}
		if json.Unmarshal(raw, &manifest) != nil || manifest.Package != "-tags live ./ai/live" {
			continue
		}
		count++
		ev.Record("native-bundle", map[string]any{"sibling": entry.Name(), "cases": len(manifest.Cases), "manifest": json.RawMessage(raw)})
		ev.Check("six route and empty-reasoning processes retained", len(manifest.Cases) == 6, "cases=%d", len(manifest.Cases))
		for _, c := range manifest.Cases {
			ev.Check("native process assertion "+c.ID, c.Status == "PASS", "status=%s", c.Status)
		}
	}
	ev.Check("one complete native auxiliary bundle", count == 1, "bundles=%d", count)
}
