package release

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/tokenbeat-lab/barness/ai"
)

type revocationResult struct {
	Usage      ai.Usage
	StopReason ai.StopReason
	Metadata   ai.CallMetadata
}

// RevocationEvidence consumes issue 06's public runtime artifacts for every
// authorization mode at both initial-input and final-Hook boundaries.
func RevocationEvidence(b Bundle) EvidenceCheck {
	e := EvidenceCheck{Name: "authorization-revocation", Bundle: b.Dir, Status: Fail}
	if len(b.Problems) != 0 {
		e.Detail = "manifest.json: artifact inventory or hash validation failed"
		return e
	}
	for _, mode := range []string{"disabled-binding", "removed-model", "revoked-images"} {
		for _, entry := range []string{"initial-images", "hook-images"} {
			id := "P10-OD58-runtime-" + mode + "-" + entry
			dir := b.caseDirs[id]
			if dir == "" || !loadCasePassed(b, id) {
				e.Detail = id + ": manifest.json: required PASS case missing"
				return e
			}
			if err := verifyRevocationCase(dir, id, mode, entry); err != nil {
				e.Detail = id + ": " + err.Error()
				return e
			}
		}
	}
	e.Status = Pass
	e.Detail = "six initial/final-Hook revocations: new calls denied, pinned attempts and reported cost preserved, no fallback, resources zero"
	return e
}

// Diagnostics contain fixed case/file names and validation rules only. Artifact
// contents and JSON decoder errors may contain private data and are not echoed.
func readRevocationArtifact(dir, name string, v any) error {
	raw, err := readArtifact(filepath.Join(dir, name), 4<<20)
	if err != nil || json.Unmarshal(raw, v) != nil {
		return fmt.Errorf("%s: required JSON missing or invalid", name)
	}
	return nil
}

func verifyRevocationCase(dir, id, mode, entry string) error {
	var a struct {
		CaseID     string `json:"case_id"`
		Status     string
		Assertions []struct {
			Name string
			Pass bool
		}
	}
	if err := readRevocationArtifact(dir, "assertions.json", &a); err != nil {
		return err
	}
	if a.CaseID != id || a.Status != "PASS" {
		return fmt.Errorf("assertions.json: case identity or PASS status invalid")
	}
	for _, assertion := range a.Assertions {
		if !assertion.Pass {
			return fmt.Errorf("assertions.json: failed assertion")
		}
	}
	for _, name := range []string{"new authorization rejected without fallback", "in-flight retry retains authorization and cost", "active attempts identical after revocation", "prior results cannot lose reported consumption", "restored authorization observed on fresh call"} {
		if !slices.ContainsFunc(a.Assertions, func(v struct {
			Name string
			Pass bool
		}) bool { return v.Name == name && v.Pass }) {
			return fmt.Errorf("assertions.json: required revocation assertion missing")
		}
	}
	phase := ai.PhaseCapability
	if mode == "disabled-binding" {
		phase = ai.PhaseBinding
	}
	if mode == "revoked-images" && entry == "hook-images" {
		phase = ai.PhaseRequest
	}
	for _, call := range []struct {
		name, id, version string
		attempts          int
	}{
		{"prior", "req-prior", "b1", 1}, {"active", "req-active", "b1", 2}, {"denied", "req-after-revoke", "", 0}, {"restored", "req-restored", "b3", 1},
	} {
		file := call.name + "-result.json"
		var r revocationResult
		if err := readRevocationArtifact(dir, file, &r); err != nil {
			return err
		}
		m := r.Metadata
		if m.TenantID != "tenant-a" || m.BindingID != "luna" || m.Operation != ai.OperationClassifier || m.RequestID != call.id || len(m.Attempts) != call.attempts {
			return fmt.Errorf("%s: call identity or attempt count invalid", file)
		}
		if call.attempts == 0 {
			if r.StopReason != ai.StopReasonError || !reflect.DeepEqual(r.Usage, ai.Usage{}) {
				return fmt.Errorf("%s: denied call must fail without consumption", file)
			}
		} else {
			if !m.Resolved || m.BindingVersion != call.version || m.ProviderID != ai.ProviderOpenAI || m.API != ai.APIOpenAIDecisions || m.ModelID != "same-model" || m.AccountScopeID != "shared-openai" || m.CredentialVersion != "v1" || m.CatalogHash == "" || m.CatalogVersion == "" || r.StopReason != ai.StopReasonStop || r.Usage.Cost.Total <= 0 || r.Usage.TotalTokens <= 0 || !reflect.DeepEqual(r.Usage, m.Attempts[len(m.Attempts)-1].Usage) {
				return fmt.Errorf("%s: pinned identity, success or reported consumption invalid", file)
			}
			for i, attempt := range m.Attempts {
				if attempt.AttemptID != fmt.Sprintf("%s#%d", call.id, i+1) {
					return fmt.Errorf("%s: attempt identity invalid", file)
				}
			}
		}
		if err := verifyRevocationObservations(dir, call.id, r, phase); err != nil {
			return err
		}
	}
	if err := verifyRevokedBinding(dir, mode); err != nil {
		return err
	}
	if err := verifyRevocationCaptures(dir, mode); err != nil {
		return err
	}
	var gauges map[string]*int64
	if err := readRevocationArtifact(dir, "resources.json", &gauges); err != nil {
		return err
	}
	if !loadResourcesReleased(gauges) {
		return fmt.Errorf("resources.json: required resource gauges missing or nonzero")
	}
	return nil
}

func verifyRevokedBinding(dir, mode string) error {
	var original, revoked ai.Binding
	if err := readRevocationArtifact(dir, "binding-tenant-a-luna.json", &original); err != nil {
		return err
	}
	if err := readRevocationArtifact(dir, "revoked-binding.json", &revoked); err != nil {
		return err
	}
	if original.TenantID != "tenant-a" || original.BindingID != "luna" || original.Version != "b1" || !original.Enabled || original.Operation != ai.OperationClassifier || original.ProviderID != ai.ProviderOpenAI || original.API != ai.APIOpenAIDecisions || !original.AllowClassifierImages || !slices.Contains(original.AllowedModels, "same-model") {
		return fmt.Errorf("binding-tenant-a-luna.json: initial authorization invalid")
	}
	original.Version = "b2"
	switch mode {
	case "disabled-binding":
		original.Enabled = false
	case "removed-model":
		original.AllowedModels = nil
	case "revoked-images":
		original.AllowClassifierImages = false
	}
	if !reflect.DeepEqual(original, revoked) {
		return fmt.Errorf("revoked-binding.json: expected authorization change or pinned route disagrees")
	}
	return nil
}
