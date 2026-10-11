package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeArtifact(t *testing.T, dir, name string, v any) {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), raw, 0644); err != nil {
		t.Fatal(err)
	}
}

// The frozen record boundary must check values, not just field presence.
func TestImportedDifferentialRejectsWrongOracle(t *testing.T) {
	for _, field := range []string{"pi_version", "pi_commit", "pi_model_data_sha256", "request_sha256", "sdk_versions", "gate", "case_id", "model_catalog_hash"} {
		t.Run(field, func(t *testing.T) {
			dir := t.TempDir()
			id := "PIDIFF-P01-known"
			r := map[string]any{"case_id": id, "pi_commit": "abe508e1b89912adde45528136c3221eb69acdd7", "pi_version": "1.1.0", "pi_model_data_sha256": "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "sdk_versions": map[string]string{"@earendil-works/pi-ai": "1.1.0", "go": "go1.26.2"}, "model_catalog_hash": "sha256:8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "request_sha256": map[string]string{"pi": "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "barness": "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8"}, "frame_sha256": "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "findings": []any{}, "pending": 0, "pass": true, "gate": "PASS"}
			r[field] = "wrong"
			writeArtifact(t, dir, "manifest.json", map[string]any{"package": OfflinePackage, "cases": []map[string]string{{"id": id, "dir": id, "status": "PASS"}}})
			writeArtifact(t, dir, id+"/pidiff.json", r)
			b, err := LoadBundle(dir)
			if err == nil && (len(b.Diffs) != 1 || len(b.Diffs[0].Missing) == 0) {
				t.Fatal("invalid frozen differential accepted")
			}
		})
	}
}

func TestDesignLoadReportCannotBeMissingOrIncomplete(t *testing.T) {
	for _, name := range []string{"missing", "empty", "excessive-memory", "missing-headroom", "resource-leak", "concurrency"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			id := "E08-mixed-pressure-local-design-load"
			writeArtifact(t, dir, "manifest.json", map[string]any{"package": OfflinePackage, "cases": []map[string]string{{"id": id, "dir": id, "status": "PASS"}}})
			if name != "missing" {
				raw, err := os.ReadFile("../../../../.scratch/barness-ai-pi-1.0/mixed-operations-evidence/pressure.json")
				if err != nil {
					t.Fatal(err)
				}
				var frozen struct{ Cases []map[string]any }
				if err := json.Unmarshal(raw, &frozen); err != nil {
					t.Fatal(err)
				}
				r := frozen.Cases[1]
				switch name {
				case "empty":
					r = map[string]any{}
				case "excessive-memory":
					r["totalAllocDuringCalls"] = 999999999
				case "missing-headroom":
					r["headroom"] = map[string]any{}
				case "concurrency":
					r["peakPermits"] = 1
				}
				writeArtifact(t, dir, id+"/mixed-pressure.json", r)
				gauges := map[string]int{"calls": 0, "permits": 0, "waiters": 0, "events": 0, "bodies": 0, "hostPermits": 0}
				if name == "resource-leak" {
					gauges["bodies"] = 1
				}
				writeArtifact(t, dir, id+"/resources.json", gauges)
			}
			b, err := LoadBundle(dir)
			if err != nil {
				t.Fatal(err)
			}
			checks := DesignLoadEvidence(b)
			if checks[0].Status == Pass {
				t.Fatal("incomplete design-load artifact passed")
			}
		})
	}
}

func TestChineseEffectNeedsActualCompleteEvidence(t *testing.T) {
	base := "../../../../.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/"
	// This is complete historical evidence for the previous catalog. It must
	// not certify the current catalog after the pi 1.1.0 migration.
	if e := ChineseEffectEvidence(base + "live"); e.Status != Fail || e.Detail != "dataset/config/catalog drift" {
		t.Fatalf("historical catalog evidence was not rejected for drift: %+v", e)
	}
	for _, dir := range []string{t.TempDir(), base + "not-run"} {
		if e := ChineseEffectEvidence(dir); e.Status == Pass {
			t.Fatal("missing/unexecuted business evaluation passed")
		}
	}
}
