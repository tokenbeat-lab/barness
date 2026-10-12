package release

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// These pins identify the source-equivalent 1.1.0 copy approved in ADR-0004.
// Changing the oracle requires a reviewed baseline migration, not a new input
// file claiming that a different version is acceptable.
const oracleCommit = "abe508e1b89912adde45528136c3221eb69acdd7"
const oracleModelData = "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8"

var hexHash = regexp.MustCompile(`^[a-f0-9]{64}$`)
var catalogHash = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

func differentialRecordProblems(id string, fields map[string]json.RawMessage) []string {
	raw, err := json.Marshal(fields)
	var r pioracle.Record
	if err != nil || json.Unmarshal(raw, &r) != nil {
		return []string{"invalid differential record values"}
	}
	var problems []string
	for _, c := range []struct {
		name  string
		valid bool
	}{
		{"case_id", r.CaseID == id},
		{"pi_version", r.PiVersion == "1.1.0"},
		{"pi_commit", r.PiCommit == oracleCommit},
		{"pi_model_data_sha256", r.PiModelDataSHA256 == oracleModelData},
		{"sdk_versions", r.SDKVersions["@earendil-works/pi-ai"] == "1.1.0" && r.SDKVersions["go"] != ""},
		{"model_catalog_hash", catalogHash.MatchString(r.ModelCatalogHash)},
		{"request_sha256", len(r.RequestSHA256) == 2 && hexHash.MatchString(r.RequestSHA256["pi"]) && hexHash.MatchString(r.RequestSHA256["barness"])},
		{"frame_sha256", hexHash.MatchString(r.FrameSHA256)},
		{"gate", r.Gate == pioracle.GatePass && r.Pass && r.Pending == 0},
	} {
		if !c.valid {
			problems = append(problems, c.name+" mismatch")
		}
	}
	pending := 0
	for _, f := range r.Findings {
		if f.Classification == pioracle.Pending {
			pending++
		}
	}
	if pending != r.Pending {
		problems = append(problems, "findings/pending mismatch")
	}
	if strings.HasPrefix(id, "PIDIFF-P10-") {
		if r.Protocol != "openai-decisions" {
			problems = append(problems, "protocol mismatch")
		}
		if r.PiSourceSHA256 != pioracle.SourceSHA256 {
			problems = append(problems, "pi_source_sha256 mismatch")
		}
		if r.SDKLockSHA256 != pioracle.SDKLockSHA256 {
			problems = append(problems, "sdk_lock_sha256 mismatch")
		}
		if len(r.ResultSHA256) != 2 || !hexHash.MatchString(r.ResultSHA256["pi"]) || !hexHash.MatchString(r.ResultSHA256["barness"]) {
			problems = append(problems, "result_sha256 mismatch")
		}
	}
	return problems
}

// P10 requires full observations and binds their lossless compact JSON to
// the record. Format-valid hashes alone cannot substitute for captured work.
func decisionsCaptureProblems(dir string, fields map[string]json.RawMessage) []string {
	raw, _ := json.Marshal(fields)
	var r pioracle.Record
	if json.Unmarshal(raw, &r) != nil {
		return []string{"invalid capture record"}
	}
	var problems []string
	for _, side := range []string{"pi", "barness"} {
		var obs pioracle.Observation
		if err := readJSON(dir+"/"+side+"-observation.json", &obs); err != nil {
			problems = append(problems, side+" observation missing or invalid")
			continue
		}
		request, err := json.Marshal(obs.Request)
		if err != nil || pioracle.SHA256(request) != r.RequestSHA256[side] {
			problems = append(problems, side+" request hash mismatch")
		}
		result, err := json.Marshal(obs.Result)
		if err != nil || pioracle.SHA256(result) != r.ResultSHA256[side] {
			problems = append(problems, side+" result hash mismatch")
		}
	}
	return problems
}
