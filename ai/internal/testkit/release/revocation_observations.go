package release

import (
	"fmt"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func verifyRevocationObservations(files verifiedCase, id string, r revocationResult, phase ai.Phase) error {
	file := "observations-" + id + ".json"
	var records []ai.Observation
	if err := readRevocationArtifact(files, file, &records); err != nil {
		return err
	}
	fail := func() error {
		return fmt.Errorf("%s: Observer lifecycle, identity or result reconciliation failed", file)
	}
	if len(records) != 2+2*len(r.Metadata.Attempts) {
		return fail()
	}
	start, last := records[0], records[len(records)-1]
	if start.Kind != ai.ObservationCallStarted || start.Call.Resolved || start.Call.TenantID != r.Metadata.TenantID || start.Call.BindingID != r.Metadata.BindingID || start.Call.Operation != r.Metadata.Operation || last.Kind != ai.ObservationCallFinished || !reflect.DeepEqual(last.Call, r.Metadata) || !reflect.DeepEqual(last.Usage, r.Usage) || last.StopReason != r.StopReason {
		return fail()
	}
	if r.StopReason == ai.StopReasonError {
		if last.Error == nil || last.Error.Code != ai.CodeTenantDenied || last.Error.Phase != phase {
			return fail()
		}
	} else if last.Error != nil {
		return fail()
	}
	for i, o := range records {
		if o.Time.IsZero() || o.Call.RequestID != id || (i > 0 && o.Time.Before(records[i-1].Time)) {
			return fail()
		}
	}
	pinned := r.Metadata
	pinned.Attempts = nil
	for i, a := range r.Metadata.Attempts {
		began, ended := records[1+2*i], records[2+2*i]
		if began.Kind != ai.ObservationAttemptStarted || ended.Kind != ai.ObservationAttemptFinished || !reflect.DeepEqual(began.Call, pinned) || !reflect.DeepEqual(ended.Call, pinned) || began.Attempt == nil || began.Attempt.AttemptID != a.AttemptID || ended.Attempt == nil || !reflect.DeepEqual(*ended.Attempt, a) {
			return fail()
		}
	}
	return nil
}

func verifyRevocationCaptures(files verifiedCase, mode string) error {
	// Reuse the captured wire artifact schema, then independently validate it;
	// neither the manifest's PASS nor the runtime assertion text proves this.
	var requests []provider.Request
	if err := readRevocationArtifact(files, "requests.json", &requests); err != nil {
		return err
	}
	expected := 11
	if mode == "revoked-images" {
		expected++
	}
	var active []provider.Request
	tenantLuna := 0
	for _, r := range requests {
		if r.Path == "/tenant-a/luna/decisions" {
			tenantLuna++
		}
		if r.Header["X-Barness-Test-Call"] == "active" {
			active = append(active, r)
		}
	}
	expectedLuna := 4
	if mode == "revoked-images" {
		expectedLuna++
	}
	if len(requests) != expected || tenantLuna != expectedLuna || len(active) != 2 || !reflect.DeepEqual(active[0], active[1]) {
		return fmt.Errorf("requests.json: capture count or pinned retry differs; possible fallback")
	}
	r := active[0]
	if r.Method != "POST" || r.Path != "/tenant-a/luna/decisions" || r.Query != "" || r.KeyAlias != "key:tenant-a@v1" || len(r.Body) == 0 {
		return fmt.Errorf("requests.json: pinned request target or credential alias invalid")
	}
	return nil
}
