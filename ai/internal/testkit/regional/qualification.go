package regional

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tokenbeat-lab/barness/ai"
)

// QualificationCatalog combines the Host's fixed catalog with the independent
// candidate snapshot in either formal state. A complete existing regional set
// must match every fact; partial promotion and metadata conflicts are errors.
// Duplicate full identities remain errors at the public Catalog boundary.
func QualificationCatalog(base ai.Catalog) (ai.Catalog, error) {
	if _, err := base.Hash(); err != nil {
		return ai.Catalog{}, err
	}
	candidate, err := CandidateCatalog()
	if err != nil {
		return ai.Catalog{}, err
	}
	present := 0
	for _, want := range candidate.Models {
		got, ok := base.Lookup(want.Provider, want.API, want.ID)
		if !ok {
			continue
		}
		present++
		if !reflect.DeepEqual(got, want) {
			return ai.Catalog{}, fmt.Errorf("regional: existing full identity conflicts with own candidate facts")
		}
	}
	if present != 0 && present != len(candidate.Models) {
		return ai.Catalog{}, fmt.Errorf("regional: partially promoted candidate identities")
	}
	// This test/Host fixture is itself an owned snapshot, not a mutable view of
	// the caller's catalog. The validated JSON boundary preserves all model kinds.
	raw, err := json.Marshal(base)
	if err != nil {
		return ai.Catalog{}, err
	}
	var out ai.Catalog
	if err = json.Unmarshal(raw, &out); err != nil {
		return ai.Catalog{}, err
	}
	out.Version = candidate.Version + "+host-qualification"
	if present == 0 {
		out.Models = append(out.Models, candidate.Models...)
	}
	if _, err = out.Hash(); err != nil {
		return ai.Catalog{}, err
	}
	return out, nil
}
