package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	raw, e := os.ReadFile("../live/testdata/glm-formal/support-views.json")
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("formal captures not yet available")
	}
	for _, route := range []string{"chat", "responses", "messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GLM12-formal-import-"+route)
			entry := index.Routes[route]
			dir := extractGLMFormalArchive(t, entry.Archive, entry.ArchiveSHA256)
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
			_, overwriteError := supportmatrix.CreateGLMSupportView(route, filepath.Dir(reportPath))
			ev.Check("existing support view is never overwritten", overwriteError != nil, "existing destination changed")
			ev.Check("own22 model capabilities", merged.Rows[0].AllPassedAt != nil && len(merged.Rows[0].Capabilities) == 22, "incomplete row")
			ev.Check("other row and caller untouched", reflect.DeepEqual(merged.Rows[1], m.Rows[1]) && len(m.Rows[0].Capabilities) == 0, "mutation")
			audit := release.AuditBundle(filepath.Dir(reportPath), nil)
			ev.Check("release import verifies same evidence", audit.Err == "" && audit.Live != nil, "%s", audit.Err)
			for _, name := range []string{"stale-hash", "candidate", "controlled", "replay", "wrong-api", "wrong-provider", "wrong-model", "narrow", "unsupported", "fake-counters", "fake-provider-id", "fake-known-usage", "fake-unknown-usage", "billing-claim"} {
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
				case "fake-known-usage":
					bad.Budget.InputTokensUsed++
				case "fake-unknown-usage":
					bad.GLM.UnknownUsageAttempts = 0
				case "billing-claim":
					bad.GLM.UsageBasis = "full-vendor-invoice"
				case "fake-provider-id":
					bad.Scenarios[0].ProviderRequestIDs = []string{"invented"}
				}
				before, _ := json.Marshal(m)
				_, err := supportmatrix.Merge(m, bad)
				after, _ := json.Marshal(m)
				ev.Check("refuses "+name, err != nil && bytes.Equal(before, after), "accepted or mutated: %v", err)
			}
			for _, name := range []string{"audit.json", "glm-support-integrity.json", "glm-support-derivation.json", "original/glm-integrity.json", "original/glm-execution-source.json", "original/formal-glm-report.json", "original/GLM11-glm-5.3-full-stream/request.json", "original/GLM11-glm-5.3-full-stream/response.sse", "original/GLM11-glm-5.3-full-stream/attempt-1-result.json", "original/GLM11-glm-5.3-full-stream/observations-1.json", "original/GLM11-glm-5.3-full-stream/assertions.json"} {
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
			for _, name := range []string{"raw-index", "raw-archive", "original-report", "importer-commit", "importer-source"} {
				proofPath := filepath.Join(filepath.Dir(reportPath), "glm-support-derivation.json")
				inventoryPath := filepath.Join(filepath.Dir(reportPath), "glm-support-integrity.json")
				proofBytes, _ := os.ReadFile(proofPath)
				reportBytes, _ := os.ReadFile(reportPath)
				inventoryBytes, _ := os.ReadFile(inventoryPath)
				manifestPath := filepath.Join(filepath.Dir(reportPath), "manifest.json")
				manifestBytes, _ := os.ReadFile(manifestPath)
				bad := r
				glmProof := *r.GLM
				proof := *r.GLM.Derivation
				glmProof.Derivation = &proof
				bad.GLM = &glmProof
				switch name {
				case "raw-index":
					proof.RawCaptureIndexSHA256 = "changed"
				case "raw-archive":
					proof.RawArchiveSHA256 = "changed"
				case "original-report":
					proof.OriginalSupportReportSHA256 = "changed"
				case "importer-commit":
					proof.ImporterGitCommit = r.GitCommit
				case "importer-source":
					proof.ImporterSourceSHA256 = "changed"
				}
				encoded, _ := supportmatrix.Encode(proof)
				os.WriteFile(proofPath, encoded, 0600)
				var manifest map[string]any
				json.Unmarshal(manifestBytes, &manifest)
				manifest["derivation"] = proof
				encoded, _ = supportmatrix.Encode(manifest)
				os.WriteFile(manifestPath, encoded, 0600)
				encoded, _ = supportmatrix.Encode(bad)
				os.WriteFile(reportPath, encoded, 0600)
				files := map[string]string{}
				filepath.WalkDir(filepath.Dir(reportPath), func(p string, d os.DirEntry, e error) error {
					if e != nil || d.IsDir() {
						return e
					}
					rel, _ := filepath.Rel(filepath.Dir(reportPath), p)
					if rel != "glm-support-integrity.json" {
						b, _ := os.ReadFile(p)
						files[filepath.ToSlash(rel)] = fmt.Sprintf("%x", sha256.Sum256(b))
					}
					return nil
				})
				encoded, _ = supportmatrix.Encode(map[string]any{"schema": 1, "files": files})
				os.WriteFile(inventoryPath, encoded, 0600)
				loaded, e := supportmatrix.LoadReport(reportPath)
				if e == nil {
					_, e = supportmatrix.Merge(m, loaded)
				}
				ev.Check("coherent local rehash cannot substitute "+name, e != nil, "accepted forged derivation")
				os.WriteFile(proofPath, proofBytes, 0600)
				os.WriteFile(reportPath, reportBytes, 0600)
				os.WriteFile(inventoryPath, inventoryBytes, 0600)
				os.WriteFile(manifestPath, manifestBytes, 0600)
			}
			oversizedPath := filepath.Join(filepath.Dir(reportPath), "original/GLM11-glm-5.3-full-stream/request.json")
			oldRequest, _ := os.ReadFile(oversizedPath)
			os.WriteFile(oversizedPath, make([]byte, 16<<20+1), 0600)
			_, oversizedError := supportmatrix.Merge(m, r)
			ev.Check("oversized captured artifact refused within read ceiling", oversizedError != nil, "oversized read accepted")
			os.WriteFile(oversizedPath, oldRequest, 0600)
			originalReport, e := supportmatrix.LoadReport(filepath.Join(filepath.Dir(reportPath), "original/live-report.json"))
			if e == nil {
				_, e = supportmatrix.Merge(m, originalReport)
			}
			ev.Check("immutable false-zero original cannot substitute corrected support view", e != nil, "old raw support report imported")
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
