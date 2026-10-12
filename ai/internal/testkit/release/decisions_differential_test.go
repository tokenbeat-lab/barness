package release

import (
	"slices"
	"testing"
)

// Issue 07 / OD48-50: P10 has an upstream oracle, and its evidence must be
// required independently of the existing six protocols and extension routes.
func TestDecisionsDifferentialCannotBeDeleted(t *testing.T) {
	trace, err := LoadTraceMap("../../../release/traceability.json")
	if err != nil {
		t.Fatal(err)
	}
	if count(trace.Differential.Required, "P10") != 1 || count(trace.P0, "P10") != 1 {
		t.Fatal("P10 is not required by the release contract")
	}
	for _, route := range trace.Differential.Routes {
		if route.Scenario == "P10" {
			t.Fatal("Decisions misclassified as an extension route")
		}
	}
	for _, name := range []string{"missing", "not-run", "ordinary-pass", "mapping-removed"} {
		t.Run(name, func(t *testing.T) {
			in := goodInputs(t)
			in.Trace.Differential.Required = append(in.Trace.Differential.Required, "P10")
			in.Trace.Scenarios["P10"] = trace.Scenarios["P10"]
			in.Trace.patterns["P10"] = trace.patterns["P10"]
			if name == "ordinary-pass" {
				in.Offline.Cases = append(in.Offline.Cases, Case{ID: "P10-OD48-object-plain", Status: Pass})
			}
			if name == "not-run" {
				in.Offline.Cases = append(in.Offline.Cases, Case{ID: "PIDIFF-P10-OD48-object-plain-classify", Status: NotRun})
			}
			if name == "mapping-removed" {
				trace.Differential.Required = slices.DeleteFunc(trace.Differential.Required, func(s string) bool { return s == "P10" })
				if CheckReleasePlan(trace).Current {
					t.Fatal("P10 mapping deletion passed")
				}
				return
			}
			expectFail(t, Evaluate(in), GateDifferential, "P10")
		})
	}
}
