package release

import (
	"encoding/json"
	"os"
	"testing"
)

// Failure first at the actual filesystem import boundary: a historical Jev
// report has no classifier-image capacity, even when its old case was PASS.
func TestDesignLoadRejectsHistoricalTextOnlyReport(t *testing.T) {
	dir := t.TempDir()
	id := "E08-mixed-pressure-local-design-load"
	raw, err := os.ReadFile("../../../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/pressure.json")
	if err != nil {
		t.Fatal(err)
	}
	var frozen struct{ Cases []map[string]any }
	if err := json.Unmarshal(raw, &frozen); err != nil {
		t.Fatal(err)
	}
	writeArtifact(t, dir, "manifest.json", map[string]any{"package": OfflinePackage, "cases": []map[string]string{{"id": id, "dir": id, "status": "PASS"}}})
	writeArtifact(t, dir, id+"/mixed-pressure.json", frozen.Cases[1])
	writeArtifact(t, dir, id+"/resources.json", map[string]int{"calls": 0, "permits": 0, "waiters": 0, "events": 0, "bodies": 0, "hostPermits": 0})
	b, err := LoadBundle(dir)
	if err != nil {
		t.Fatal(err)
	}
	if checks := DesignLoadEvidence(b); checks[0].Status == Pass {
		t.Fatal("historical text-only report certifies classifier images")
	}
}
