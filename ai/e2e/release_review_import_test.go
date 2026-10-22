package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// These coherently rehashed mutations exercise semantic validation after the
// inventory layer accepts the artifacts. A recorded PASS is not sufficient.
func TestReleaseRevocationSemanticIntegrity(t *testing.T) {
	for _, file := range []string{"requests.json", "observations-req-active.json", "observations-req-prior.json", "observations-req-after-revoke.json", "revoked-binding.json", "active-result.json", "resources.json"} {
		for _, mutation := range []string{"null", "{}", "[]", "mismatch"} {
			t.Run(file+"/"+mutation, func(t *testing.T) {
				ev := run.Case(t, "OD58-release-semantic-"+strings.TrimSuffix(file, ".json")+"-"+strings.NewReplacer("{}", "object", "[]", "array").Replace(mutation))
				dir := releaseRollbackBundle(t)
				id := "P10-OD58-runtime-disabled-binding-initial-images"
				path := filepath.Join(dir, id, file)
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if mutation == "mismatch" {
					var v any
					mustUnmarshal(t, raw, &v)
					switch file {
					case "requests.json":
						for _, x := range v.([]any) {
							c := x.(map[string]any)
							if c["header"].(map[string]any)["X-Barness-Test-Call"] == "active" {
								c["path"] = "/tenant-b/luna/decisions"
								break
							}
						}
					case "revoked-binding.json":
						v.(map[string]any)["Enabled"] = true
					case "active-result.json":
						v.(map[string]any)["usage"].(map[string]any)["cost"].(map[string]any)["total"] = 0
					case "resources.json":
						v.(map[string]any)["calls"] = 1
					default:
						a := v.([]any)
						a[len(a)-1].(map[string]any)["call"].(map[string]any)["tenantId"] = "tenant-b"
					}
					raw, err = json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					raw = []byte(mutation)
				}
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
				updateReleaseInventory(t, dir, id, file, evidence.Digest(raw))
				b, err := release.LoadBundle(dir)
				if err != nil {
					t.Fatal(err)
				}
				ev.Check("mutation passes integrity layer", len(b.Problems) == 0, "unexpected integrity failure")
				got := release.RevocationEvidence(b)
				ev.Record("import", map[string]any{"status": got.Status, "detail": got.Detail})
				ev.Check("invalid semantic proof rejected", got.Status == release.Fail, "accepted %s", file)
				ev.Check("safe diagnostic identifies case and file", strings.Contains(got.Detail, id) && strings.Contains(got.Detail, file), "got %s", got.Detail)
			})
		}
	}
}

func updateReleaseInventory(t *testing.T, dir, id, file, hash string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	mustUnmarshal(t, raw, &m)
	for _, x := range m["cases"].([]any) {
		c := x.(map[string]any)
		if c["id"] == id {
			a := c["artifacts"].(map[string]any)
			if hash == "" {
				delete(a, file)
			} else {
				a[file] = hash
			}
		}
	}
	writeReleaseJSON(t, dir, "manifest.json", m)
}

func TestReleaseLiveInventoryCannotOmitProof(t *testing.T) {
	for _, file := range []string{"exchanges.json", "result.json", "observations.json"} {
		t.Run(file, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-live-inventory-"+strings.TrimSuffix(file, ".json"))
			dir := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS("../../.scratch/barness-ai-openai-decisions/formal-support-evidence/live")); err != nil {
				t.Fatal(err)
			}
			before := release.AuditBundle(dir, nil)
			ev.Check("complete original bundle accepted", before.Err == "", "got %s", before.Err)
			id := "LIVE-P10-image-png"
			updateReleaseInventory(t, dir, id, file, "")
			if err := os.WriteFile(filepath.Join(dir, id, file), []byte("null\n"), 0600); err != nil {
				t.Fatal(err)
			}
			after := release.AuditBundle(dir, nil)
			ev.Record("import", map[string]any{"file": file, "error": after.Err})
			ev.Check("omitted proof is rejected despite existing file", after.Err != "", "unhashed proof accepted")
		})
	}
}
