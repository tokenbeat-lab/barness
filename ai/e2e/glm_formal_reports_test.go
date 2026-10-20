package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
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
	var known struct {
		SourceCommit string
		Routes       map[string]struct {
			ArchiveSHA256, ReportSHA256 string
			InputSide, Output           int
		}
	}
	literal, err := os.ReadFile("../live/testdata/glm-formal/known-usage.json")
	if err != nil || json.Unmarshal(literal, &known) != nil {
		t.Fatal("independent current-capture expectation unavailable")
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
			// The existing matrix may retain this immutable historical pass;
			// refusing current applicability must neither clear nor refresh it.
			m.Rows[0].AllPassedAt = &r.FinishedAt
			beforeMatrix, _ := json.Marshal(m)
			_, e = supportmatrix.Merge(m, r)
			ev.Check("historical own capture refuses changed current execution source", glmExecutionSourceDrift(e), "%v", e)
			afterMatrix, _ := json.Marshal(m)
			ev.Check("stale execution source cannot update caller matrix", bytes.Equal(beforeMatrix, afterMatrix), "matrix mutated")
			_, overwriteError := supportmatrix.CreateGLMSupportView(route, filepath.Dir(reportPath))
			ev.Check("existing support view is never overwritten", overwriteError != nil, "existing destination changed")
			ev.Check("historical own22 model-qualified cases retained", len(r.Expected) == 22 && len(r.Scenarios) == 22 && r.Provider == "zhipu", "incomplete historical report")
			ev.Check("other row and caller untouched", reflect.DeepEqual(m.Rows[1], supportmatrix.Row{Combo: "openai-chat", Operation: "chat", Provider: "openai", API: "openai-completions"}) && len(m.Rows[0].Capabilities) == 0, "mutation")
			audit := release.AuditBundle(filepath.Dir(reportPath), nil)
			ev.Check("release audit refuses same stale execution source", audit.Err == "live artifact integrity verification failed" && audit.Live != nil && glmExecutionSourceDrift(e), "%s", audit.Err)
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
					proof.ImporterGitCommit = "f5cf5362610a8bdb8f3e170f408f3858c149e540"
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
			want := known.Routes[route]
			ev.Check("historical raw totals survive while current execution acceptance refuses", glmExecutionSourceDrift(e) && originalReport.GitCommit == known.SourceCommit && originalReport.Budget.InputTokensUsed == want.InputSide && originalReport.Budget.OutputTokensUsed == want.Output && r.GLM.Derivation.RawArchiveSHA256 == want.ArchiveSHA256 && r.GLM.Derivation.OriginalFormalReportSHA256 == want.ReportSHA256, "err=%v budget=%+v expected=%+v", e, originalReport.Budget, want)
			zero := originalReport
			zero.Budget.InputTokensUsed, zero.Budget.OutputTokensUsed = 0, 0
			_, zeroError := supportmatrix.Merge(m, zero)
			ev.Check("false-zero raw budget contradicting captured Attempts refused", zeroError != nil, "accepted false-zero raw budget")
			// A later actual failure retains the last complete pass. It is tested by
			// Merge's shared public history suite; this own report must retain history
			// when a rejected newer substituted report is submitted.
			newer := r
			newer.FinishedAt = r.FinishedAt.Add(time.Second)
			_, err := supportmatrix.Merge(m, newer)
			ev.Check("history cannot be replaced by fabricated newer metadata", err != nil && len(m.Rows[0].Capabilities) == 0 && m.Rows[0].AllPassedAt.Equal(r.FinishedAt), "history lost")
			cat := ai.BuiltinCatalog()
			hash, _ := cat.Hash()
			ev.Check("historical catalog snapshot still exact despite changed execution source", r.CatalogHash == hash && r.CatalogVersion == cat.Version, "stale")
			ev.Record("historical-report-current-source-refused", r)
			ev.Record("unchanged-own-row", m.Rows[0])
		})
	}
}

// Shared implementation/cache changes invalidate old own live execution bytes.
// Exact classification distinguishes source staleness from missing archives,
// a malformed report, or any unrelated consumer failure.
func glmExecutionSourceDrift(err error) bool {
	return errors.Is(err, supportmatrix.ErrInvalidReport) && err.Error() == "supportmatrix: invalid report: GLM execution source drift"
}
