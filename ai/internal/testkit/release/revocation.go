package release

import (
	"encoding/json"
	"path/filepath"
	"slices"

	"github.com/tokenbeat-lab/barness/ai"
)

// RevocationEvidence consumes issue 06's public runtime artifacts for every
// authorization mode at both initial-input and final-Hook boundaries.
func RevocationEvidence(b Bundle) EvidenceCheck {
	e := EvidenceCheck{Name: "authorization-revocation", Bundle: b.Dir, Status: Fail, Detail: "authorization revocation evidence incomplete"}
	if len(b.Problems) != 0 {
		return e
	}
	for _, mode := range []string{"disabled-binding", "removed-model", "revoked-images"} {
		for _, entry := range []string{"initial-images", "hook-images"} {
			id := "P10-OD58-runtime-" + mode + "-" + entry
			dir := b.caseDirs[id]
			if dir == "" || !loadCasePassed(b, id) {
				return e
			}
			var a struct {
				CaseID     string `json:"case_id"`
				Status     string
				Assertions []struct {
					Name string
					Pass bool
				}
			}
			if readJSON(filepath.Join(dir, "assertions.json"), &a) != nil || a.CaseID != id || a.Status != "PASS" {
				return e
			}
			for _, name := range []string{"new authorization rejected without fallback", "in-flight retry retains authorization and cost", "active attempts identical after revocation", "prior results cannot lose reported consumption", "restored authorization observed on fresh call"} {
				if !slices.ContainsFunc(a.Assertions, func(v struct {
					Name string
					Pass bool
				}) bool { return v.Name == name && v.Pass }) {
					return e
				}
			}
			var prior, active, denied, restored struct {
				Usage      ai.Usage
				StopReason ai.StopReason
				Metadata   ai.CallMetadata
			}
			for name, value := range map[string]any{"prior-result.json": &prior, "active-result.json": &active, "denied-result.json": &denied, "restored-result.json": &restored} {
				if readJSON(filepath.Join(dir, name), value) != nil {
					return e
				}
			}
			if prior.Usage.Cost.Total <= 0 || active.Usage.Cost.Total <= 0 || prior.StopReason != ai.StopReasonStop || active.StopReason != ai.StopReasonStop || active.Metadata.BindingVersion != "b1" || len(active.Metadata.Attempts) != 2 || active.Metadata.API != ai.APIOpenAIDecisions || denied.StopReason != ai.StopReasonError || len(denied.Metadata.Attempts) != 0 || restored.Metadata.BindingVersion != "b3" || restored.Metadata.API != ai.APIOpenAIDecisions || restored.StopReason != ai.StopReasonStop {
				return e
			}
			// Captures and per-call Observer artifacts are required even if a rewritten
			// manifest deletes their entries; LoadBundle verifies their recorded hashes.
			for _, name := range []string{"requests.json", "observations-req-active.json", "observations-req-prior.json", "observations-req-after-revoke.json", "revoked-binding.json"} {
				raw, err := readArtifact(filepath.Join(dir, name), 4<<20)
				if err != nil || !json.Valid(raw) || len(raw) < 3 {
					return e
				}
			}
			var gauges map[string]*int64
			if readJSON(filepath.Join(dir, "resources.json"), &gauges) != nil || !loadResourcesReleased(gauges) {
				return e
			}
		}
	}
	e.Status = Pass
	e.Detail = "six initial/final-Hook revocations: new calls denied, pinned attempts and reported cost preserved, no fallback, resources zero"
	return e
}
