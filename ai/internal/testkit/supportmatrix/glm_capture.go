package supportmatrix

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

func readGLMJSON(path string, v any) error {
	raw, err := glmRead(path)
	if err != nil {
		return err
	}
	guard := json.NewDecoder(bytes.NewReader(raw))
	guard.UseNumber()
	if err = uniqueFields(guard, 0); err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}
func glmRead(path string) ([]byte, error) {
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() > 16<<20 {
		return nil, glmBad("nonregular, missing or oversized capture")
	}
	return os.ReadFile(path)
}
func glmHash(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }

func verifyGLMBundle(dir string, r Report, detail glmDetail) error {
	var inventory struct {
		Schema int
		Files  map[string]string
	}
	if readGLMJSON(filepath.Join(dir, "glm-integrity.json"), &inventory) != nil || inventory.Schema != 1 || len(inventory.Files) < 300 {
		return glmBad("complete integrity inventory missing")
	}
	actual := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		if rel == ReportFile || rel == "glm-integrity.json" {
			return nil
		}
		raw, e := glmRead(path)
		if e != nil {
			return e
		}
		actual[rel] = glmHash(raw)
		return nil
	})
	if err != nil || !reflect.DeepEqual(actual, inventory.Files) {
		return glmBad("capture inventory missing, changed or extra file")
	}
	var a struct{ Findings []audit.Finding }
	if readGLMJSON(filepath.Join(dir, "audit.json"), &a) != nil || a.Findings == nil || len(a.Findings) > 0 {
		return glmBad("audit receipt missing or failed")
	}
	if findings, e := audit.Bundle(dir, nil); e != nil || len(findings) > 0 {
		return glmBad("independent audit failed")
	}
	var manifest struct {
		Package  string
		Commit   string    `json:"git_commit"`
		Started  time.Time `json:"started_at"`
		Finished time.Time `json:"finished_at"`
		Versions map[string]string
		Cases    []struct {
			ID, Dir, Status string
			Artifacts       map[string]string
		}
	}
	if readGLMJSON(filepath.Join(dir, "manifest.json"), &manifest) != nil || manifest.Package != "-tags live -timeout 1800s ./ai/live" || manifest.Commit != r.GitCommit || manifest.Started.IsZero() || manifest.Finished.Before(r.FinishedAt) || manifest.Versions["glm_kind"] != "glm-formal" || manifest.Versions["glm_mode"] != "live" || manifest.Versions["live_combo"] != r.Combo || manifest.Versions["live_sdk"] != r.SDK || manifest.Versions["model_catalog_hash"] != r.CatalogHash || manifest.Versions["model_catalog_version"] != r.CatalogVersion || manifest.Versions["glm_source_manifest_hash"] != GLMSourcePin {
		return glmBad("manifest provenance differs")
	}
	source, e := glmRead(filepath.Join(dir, "GLM11-sources", "sources.json"))
	if e != nil || glmHash(source) != GLMSourcePin {
		return glmBad("source or price manifest differs")
	}
	var sources struct {
		CandidateSHA256, PlanSHA256 string
		Sources                     []struct{ File, SHA256 string }
	}
	if json.Unmarshal(source, &sources) != nil {
		return glmBad("source manifest invalid")
	}
	snapshots := map[string]string{"candidate.json": sources.CandidateSHA256, "plan.json": sources.PlanSHA256}
	for _, s := range sources.Sources {
		if filepath.Base(s.File) != s.File || snapshots[s.File] != "" {
			return glmBad("source snapshot identity invalid")
		}
		snapshots[s.File] = s.SHA256
	}
	for name, hash := range snapshots {
		raw, e := glmRead(filepath.Join(dir, "GLM11-sources", name))
		if e != nil || glmHash(raw) != hash {
			return glmBad("referenced source/plan/candidate/price snapshot changed")
		}
	}
	var recorded map[string]string
	if readGLMJSON(filepath.Join(dir, "glm-execution-source.json"), &recorded) != nil {
		return glmBad("execution source snapshot missing")
	}
	current, e := GLMExecutionSource()
	if e != nil || !reflect.DeepEqual(recorded, current) {
		return glmBad("execution source drift")
	}
	for scenarioIndex, sc := range r.Scenarios {
		matches := 0
		for _, c := range manifest.Cases {
			if c.ID != sc.CaseID {
				continue
			}
			matches++
			if c.Dir != c.ID || filepath.Base(c.Dir) != c.Dir || c.Status != string(sc.Outcome) {
				return glmBad("case outcome differs")
			}
			base := filepath.Join(dir, c.Dir)
			entries, e := os.ReadDir(base)
			if e != nil || len(entries) != len(c.Artifacts) {
				return glmBad("case inventory incomplete")
			}
			for _, entry := range entries {
				raw, e := glmRead(filepath.Join(base, entry.Name()))
				if e != nil || c.Artifacts[entry.Name()] != glmHash(raw) {
					return glmBad("case artifact changed")
				}
			}
			var assertions struct {
				ID         string `json:"case_id"`
				Status     string
				Assertions []struct{ Pass bool }
			}
			if readGLMJSON(filepath.Join(base, "assertions.json"), &assertions) != nil || assertions.ID != sc.CaseID || assertions.Status != string(sc.Outcome) {
				return glmBad("assertions missing")
			}
			if sc.Outcome == Pass {
				if len(assertions.Assertions) == 0 {
					return glmBad("PASS lacks assertions")
				}
				for _, v := range assertions.Assertions {
					if !v.Pass {
						return glmBad("PASS contains failed assertion")
					}
				}
			}
			if sc.Attempts == 0 {
				continue
			}
			if e = verifyGLMCall(base, r, sc, detail.Scenarios[scenarioIndex].RequestIDSources); e != nil {
				return e
			}
		}
		if matches != 1 {
			return glmBad("scenario lacks unique own captured case")
		}
	}
	return nil
}

func verifyGLMCall(base string, r Report, s ScenarioResult, sources []string) error {
	var exchanges []struct {
		Method, URL, RequestBody, ResponseBody string
		Status                                 int
		Truncated                              bool
		ResponseHeaders                        map[string]string
	}
	if readGLMJSON(filepath.Join(base, "attempt-1-exchanges.json"), &exchanges) != nil || len(exchanges) != 1 || exchanges[0].Truncated {
		return glmBad("own exchange missing or truncated")
	}
	ex := exchanges[0]
	request, e := glmRead(filepath.Join(base, "request.json"))
	response, e2 := glmRead(filepath.Join(base, "response.sse"))
	suffix := "/chat/completions"
	maxField := "max_tokens"
	if r.API == "openai-responses" {
		suffix = "/responses"
		maxField = "max_output_tokens"
	}
	if r.API == "anthropic-messages" {
		suffix = "/v1/messages?beta=true"
	}
	var payload map[string]any
	if e != nil || e2 != nil || string(request) != ex.RequestBody || string(response) != ex.ResponseBody || ex.Method != "POST" || ex.URL != r.Endpoint+suffix || json.Unmarshal(request, &payload) != nil || payload["model"] != s.Model || payload[maxField] != float64(4096) {
		return glmBad("captured final request or response differs")
	}
	var outcome struct {
		Result struct {
			Metadata ai.CallMetadata
			Message  struct{ StopReason ai.StopReason }
		}
		Error string
	}
	if readGLMJSON(filepath.Join(base, "attempt-1-result.json"), &outcome) != nil {
		return glmBad("result absent")
	}
	md := outcome.Result.Metadata
	if md.ProviderID != ai.ProviderZhipu || string(md.API) != r.API || md.ModelID != s.Model || md.CatalogHash != r.CatalogHash || md.CatalogVersion != r.CatalogVersion || md.Operation != ai.OperationChat || !md.Resolved || md.BindingID != r.Combo || md.TenantID != "live-smoke" || md.AccountScopeID != "live-"+r.Combo || md.BindingVersion != "live" || md.CredentialVersion != "live" || len(md.Attempts) != s.Attempts || md.RequestID == "" {
		return glmBad("Result attribution differs")
	}
	for i, attempt := range md.Attempts {
		if attempt.ProviderRequestID != s.ProviderRequestIDs[i] {
			return glmBad("actual provider request ID differs")
		}
		headerID, headerName := "", ""
		for name, value := range ex.ResponseHeaders {
			if strings.EqualFold(name, "x-request-id") {
				headerID, headerName = value, name
			}
		}
		if attempt.ProviderRequestID != headerID {
			return glmBad("provider ID disagrees with actual optional header")
		}
		wantSource := "unreported"
		if headerID != "" {
			wantSource = "response header:" + headerName
		}
		if len(sources) != len(md.Attempts) || sources[i] != wantSource {
			return glmBad("provider request ID source differs")
		}
	}

	var observations []ai.Observation
	if readGLMJSON(filepath.Join(base, "observations-1.json"), &observations) != nil || len(observations) != 4 || observations[3].Kind != ai.ObservationCallFinished || !reflect.DeepEqual(observations[3].Call, md) {
		return glmBad("Observer terminal attribution differs")
	}
	if s.Outcome == Pass {
		cancel := strings.HasSuffix(s.ID, "/cancel")
		if ex.Status != 200 || (!cancel && (outcome.Error != "" && outcome.Error != "<nil>")) || cancel && outcome.Result.Message.StopReason != ai.StopReasonAborted {
			return glmBad("PASS lacks actual success or cancellation")
		}
	}
	return nil
}
