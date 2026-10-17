package e2e

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Actual immutable evidence is the public import boundary. Each negative case
// edits its own extracted copy; original captures and prior rows stay intact.
func TestGLMFormalReports(t *testing.T) {
	var index struct {
		Routes map[string]struct{ Archive, ArchiveSHA256, ReportPath string }
	}
	raw, e := os.ReadFile("../live/testdata/glm-formal/live-captures.json")
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("formal captures not yet available")
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GLM12-formal-import-"+route)
			entry := index.Routes[route]
			raw, e := os.ReadFile("../../" + entry.Archive)
			if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != entry.ArchiveSHA256 {
				t.Fatal("immutable archive unavailable")
			}
			dir := t.TempDir()
			archive, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
			if e != nil {
				t.Fatal(e)
			}
			for _, f := range archive.File {
				if f.FileInfo().IsDir() {
					continue
				}
				if strings.Contains(f.Name, "..") || filepath.IsAbs(f.Name) || f.UncompressedSize64 > 16<<20 {
					t.Fatal("invalid capture path")
				}
				p := filepath.Join(dir, f.Name)
				os.MkdirAll(filepath.Dir(p), 0755)
				r, e := f.Open()
				if e != nil {
					t.Fatal(e)
				}
				b, e := io.ReadAll(io.LimitReader(r, 16<<20+1))
				r.Close()
				if e != nil || len(b) > 16<<20 {
					t.Fatal("capture read")
				}
				os.WriteFile(p, b, 0600)
			}
			reportPath := filepath.Join(dir, filepath.FromSlash(entry.ReportPath))
			r, e := supportmatrix.LoadReport(reportPath)
			if e != nil {
				t.Fatal(e)
			}
			m := supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{{Combo: r.Combo, Operation: "chat", Provider: "zhipu", API: r.API}, {Combo: "openai-chat", Operation: "chat", Provider: "openai", API: "openai-completions"}}}
			merged, e := supportmatrix.Merge(m, r)
			ev.Check("own full current-hash report accepted", e == nil, "%v", e)
			if e != nil {
				return
			}
			ev.Check("own22 model capabilities", merged.Rows[0].AllPassedAt != nil && len(merged.Rows[0].Capabilities) == 22, "incomplete row")
			ev.Check("other row and caller untouched", reflect.DeepEqual(merged.Rows[1], m.Rows[1]) && len(m.Rows[0].Capabilities) == 0, "mutation")
			audit := release.AuditBundle(filepath.Dir(reportPath), nil)
			ev.Check("release import verifies same evidence", audit.Err == "" && audit.Live != nil, "%s", audit.Err)
			for _, name := range []string{"stale-hash", "candidate", "controlled", "replay", "wrong-api", "wrong-provider", "wrong-model", "narrow", "unsupported", "fake-counters", "fake-provider-id"} {
				bad := r
				bad.Expected = append([]string{}, r.Expected...)
				bad.Scenarios = append([]supportmatrix.ScenarioResult{}, r.Scenarios...)
				proof := *r.GLM
				bad.GLM = &proof
				switch name {
				case "stale-hash":
					bad.CatalogHash = "sha256:old"
				case "candidate":
					bad.GLM = nil
				case "controlled", "replay":
					bad.GLM.Mode = name
				case "wrong-api":
					bad.API = "openai-responses"
					if r.API == bad.API {
						bad.API = "anthropic-messages"
					}
				case "wrong-provider":
					bad.Provider = "openai"
				case "wrong-model":
					bad.Scenarios[0].Model = "glm-foreign"
				case "narrow":
					bad.Expected = bad.Expected[:11]
					bad.Scenarios = bad.Scenarios[:11]
				case "unsupported":
					bad.Scenarios[0].Outcome = supportmatrix.Unsupported
					bad.Scenarios[0].Note = "synthetic unsupported"
				case "fake-counters":
					bad.Budget.HTTPAttempts = 0
				case "fake-provider-id":
					bad.Scenarios[0].ProviderRequestIDs = []string{"invented"}
				}
				before, _ := json.Marshal(m)
				_, err := supportmatrix.Merge(m, bad)
				after, _ := json.Marshal(m)
				ev.Check("refuses "+name, err != nil && bytes.Equal(before, after), "accepted or mutated: %v", err)
			}
			for _, name := range []string{"audit.json", "glm-integrity.json", "glm-execution-source.json", "formal-glm-report.json", "GLM11-glm-5.3-full-stream/request.json", "GLM11-glm-5.3-full-stream/response.sse", "GLM11-glm-5.3-full-stream/attempt-1-result.json", "GLM11-glm-5.3-full-stream/observations-1.json", "GLM11-glm-5.3-full-stream/assertions.json"} {
				p := filepath.Join(filepath.Dir(reportPath), name)
				old, e := os.ReadFile(p)
				if e != nil {
					t.Fatal(e)
				}
				os.Remove(p)
				_, err := supportmatrix.Merge(m, r)
				ev.Check("missing own capture "+name, err != nil, "accepted missing capture")
				os.WriteFile(p, old, 0600)
			}
			// A later actual failure retains the last complete pass. It is tested by
			// Merge's shared public history suite; this own report must retain history
			// when a rejected newer substituted report is submitted.
			newer := r
			newer.FinishedAt = r.FinishedAt.Add(time.Second)
			_, err := supportmatrix.Merge(merged, newer)
			ev.Check("history cannot be replaced by fabricated newer metadata", err != nil && merged.Rows[0].AllPassedAt.Equal(r.FinishedAt), "history lost")
			cat := ai.BuiltinCatalog()
			hash, _ := cat.Hash()
			ev.Check("formal price snapshot exact", r.CatalogHash == hash && r.CatalogVersion == cat.Version, "stale")
			ev.Record("verified-report", r)
			ev.Record("merged-own-row", merged.Rows[0])
		})
	}
}
