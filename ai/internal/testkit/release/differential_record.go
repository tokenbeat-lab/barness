package release

import (
	"encoding/json"
	"regexp"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// These pins identify the source-equivalent 1.0.0 copy approved in ADR-0004.
// Changing the oracle requires a reviewed baseline migration, not a new input
// file claiming that a different version is acceptable.
const oracleCommit = "a13d35a742c6ef8462812a28fbe1d8c8b7431c32"
const oracleModelData = "8bd56835b763bb5e5e03ebc6e52fc0c6b2a935a02538a3c83c7b21ce7d4aee1e"

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
		{"pi_version", r.PiVersion == "1.0.0"},
		{"pi_commit", r.PiCommit == oracleCommit},
		{"pi_model_data_sha256", r.PiModelDataSHA256 == oracleModelData},
		{"sdk_versions", r.SDKVersions["@earendil-works/pi-ai"] == "1.0.0" && r.SDKVersions["go"] != ""},
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
	return problems
}
