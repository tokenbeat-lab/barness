package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Keep the same six historical revocation cases selected by the former ZIP
// import. Original artifacts and the derived manifest are pinned beside the
// bundle; provenance never promotes this fixture to current live support.
func releaseRollbackBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS("testdata/retained/decisions-runtime/bundle")); err != nil {
		t.Fatal(err)
	}
	return dir
}
func TestReleaseRollbackArtifacts(t *testing.T) {
	for _, mode := range []string{"intact", "missing-case", "missing-cost", "missing-captures", "missing-observer", "missing-assertions", "missing-resources", "omit-inventory"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD58-release-import-"+mode)
			dir := releaseRollbackBundle(t)
			id := "P10-OD58-runtime-disabled-binding-initial-images"
			raw, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			for i, v := range m["cases"].([]any) {
				c := v.(map[string]any)
				if c["id"] != id {
					continue
				}
				if mode == "missing-case" {
					m["cases"] = append(m["cases"].([]any)[:i], m["cases"].([]any)[i+1:]...)
					break
				}
				file := map[string]string{"missing-cost": "active-result.json", "missing-captures": "requests.json", "missing-observer": "observations-req-active.json", "missing-assertions": "assertions.json", "missing-resources": "resources.json", "omit-inventory": "active-result.json"}[mode]
				if file != "" {
					delete(c["artifacts"].(map[string]any), file)
					if mode != "omit-inventory" {
						if err := os.Remove(filepath.Join(dir, id, file)); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			writeReleaseJSON(t, dir, "manifest.json", m)
			b, err := release.LoadBundle(dir)
			if err != nil {
				t.Fatal(err)
			}
			got := release.RevocationEvidence(b)
			ev.Record("import", map[string]any{"mode": mode, "status": got.Status, "detail": got.Detail})
			ev.Check("complete revocation evidence required", (got.Status == release.Pass) == (mode == "intact"), "status=%s", got.Status)
		})
	}
}
func TestReleaseOfflineInventoryCannotOmitFile(t *testing.T) {
	ev := run.Case(t, "OD57-release-offline-complete-inventory")
	dir := t.TempDir()
	id := "P10-synthetic"
	os.Mkdir(filepath.Join(dir, id), 0700)
	writeReleaseJSON(t, filepath.Join(dir, id), "result.json", map[string]any{"synthetic": true})
	writeReleaseJSON(t, filepath.Join(dir, id), "resources.json", map[string]any{"calls": 0})
	raw, _ := os.ReadFile(filepath.Join(dir, id, "result.json"))
	writeReleaseJSON(t, dir, "manifest.json", map[string]any{"package": "./ai/e2e", "cases": []any{map[string]any{"id": id, "dir": id, "status": "PASS", "artifacts": map[string]string{"result.json": evidence.Digest(raw)}}}})
	b, err := release.LoadBundle(dir)
	ev.Check("unlisted artifact refused", err != nil || len(b.Problems) > 0, "unhashed artifact accepted")
}
