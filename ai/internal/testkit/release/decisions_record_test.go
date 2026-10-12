package release

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// The imported P10 record is the public evidence boundary. A forged source,
// absent result hash, or modified complete capture must not certify parity.
func TestDecisionsEvidenceRejectsSourceAndCaptureDrift(t *testing.T) {
	for _, name := range []string{"valid", "source", "results", "protocol", "capture", "sdk-missing", "sdk-wrong"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			id := "PIDIFF-P10-OD48-object-plain-classify"
			request := pioracle.Request{Body: json.RawMessage(`{"model":"gpt-6-luna"}`)}
			result := json.RawMessage(`{"answers":{},"stopReason":"stop"}`)
			requestJSON, _ := json.Marshal(request)
			r := map[string]any{"case_id": id, "protocol": "openai-decisions", "pi_commit": oracleCommit, "pi_version": "1.1.0", "pi_model_data_sha256": oracleModelData, "pi_source_sha256": pioracle.SourceSHA256, "sdk_versions": map[string]string{"@earendil-works/pi-ai": "1.1.0", "go": "go1.26.2"}, "model_catalog_hash": "sha256:" + oracleModelData, "request_sha256": map[string]string{"pi": pioracle.SHA256(requestJSON), "barness": pioracle.SHA256(requestJSON)}, "result_sha256": map[string]string{"pi": pioracle.SHA256(result), "barness": pioracle.SHA256(result)}, "frame_sha256": oracleModelData, "findings": []any{}, "pending": 0, "pass": true, "gate": "PASS"}
			r["sdk_lock_sha256"] = "e0b4084c372bc764b73faa25f2a311a3649e5b7ffb341352f25c13a891b7de65"
			switch name {
			case "sdk-missing":
				delete(r, "sdk_lock_sha256")
			case "sdk-wrong":
				r["sdk_lock_sha256"] = oracleModelData
			case "source":
				r["pi_source_sha256"] = oracleModelData
			case "results":
				delete(r, "result_sha256")
			case "protocol":
				r["protocol"] = "openai-responses"
			}
			writeArtifact(t, dir, filepath.Join(id, "pidiff.json"), r)
			obs := pioracle.Observation{Request: request, Result: result}
			writeArtifact(t, dir, filepath.Join(id, "pi-observation.json"), obs)
			if name == "capture" {
				obs.Result = json.RawMessage(`{"answers":{},"stopReason":"error"}`)
			}
			writeArtifact(t, dir, filepath.Join(id, "barness-observation.json"), obs)
			writeArtifact(t, dir, "manifest.json", map[string]any{"package": OfflinePackage, "cases": []map[string]string{{"id": id, "dir": id, "status": "PASS"}}})
			b, err := LoadBundle(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(b.Diffs) != 1 {
				t.Fatal("missing record")
			}
			valid := len(b.Diffs[0].Missing) == 0
			if valid != (name == "valid") {
				t.Fatalf("%s accepted=%v problems=%v", name, valid, b.Diffs[0].Missing)
			}
		})
	}
}
