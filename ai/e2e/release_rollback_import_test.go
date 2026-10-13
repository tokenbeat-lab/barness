package e2e

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Import the retained public runtime run, keeping each original artifact byte.
func releaseRollbackBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	z, err := zip.OpenReader("../../.scratch/barness-ai-openai-decisions/runtime-evidence/runtime.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var prefix string
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "runs/full/") && strings.HasSuffix(f.Name, "/manifest.json") {
			prefix = strings.TrimSuffix(f.Name, "manifest.json")
			break
		}
	}
	if prefix == "" {
		t.Fatal("original runtime manifest absent")
	}
	for _, f := range z.File {
		name, ok := strings.CutPrefix(f.Name, prefix)
		if !ok || (!strings.HasPrefix(name, "P10-OD58-runtime-") && name != "manifest.json" && name != "audit.json") {
			continue
		}
		if f.FileInfo().IsDir() {
			continue
		}
		if !filepath.IsLocal(name) {
			t.Fatal("unsafe archive member")
		}
		dst := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			t.Fatal(err)
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(io.LimitReader(r, 8<<20))
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	var cases []any
	for _, v := range m["cases"].([]any) {
		if strings.HasPrefix(v.(map[string]any)["id"].(string), "P10-OD58-runtime-") {
			cases = append(cases, v)
		}
	}
	m["cases"] = cases
	writeReleaseJSON(t, dir, "manifest.json", m)
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
	writeReleaseJSON(t, dir, "manifest.json", map[string]any{"package": "./ai/e2e", "cases": []any{map[string]any{"id": id, "dir": id, "status": "PASS", "artifacts": map[string]string{"result.json": pioracle.SHA256(raw)}}}})
	b, err := release.LoadBundle(dir)
	ev.Check("unlisted artifact refused", err != nil || len(b.Problems) > 0, "unhashed artifact accepted")
}
