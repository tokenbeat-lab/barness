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
			r := map[string]any{"case_id": id, "pi_commit": "a13d35a742c6ef8462812a28fbe1d8c8b7431c32", "pi_version": "1.0.0", "pi_model_data_sha256": "8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e", "sdk_versions": map[string]string{"@earendil-works/pi-ai": "1.0.0", "go": "go1.26.2"}, "model_catalog_hash": "sha256:8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e", "request_sha256": map[string]string{"pi": "8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e", "barness": "8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e"}, "frame_sha256": "8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e", "findings": []any{}, "pending": 0, "pass": true, "gate": "PASS"}
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
	if e := ChineseEffectEvidence(base + "live"); e.Status != Pass {
		t.Fatalf("actual evaluation refused: %+v", e)
	}
	for _, dir := range []string{t.TempDir(), base + "not-run"} {
		if e := ChineseEffectEvidence(dir); e.Status == Pass {
			t.Fatal("missing/unexecuted business evaluation passed")
		}
	}
}
