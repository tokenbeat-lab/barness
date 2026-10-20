package release

import (
	"slices"
	"strings"
	"testing"
)

// Regression failure inventory: a remaining scenario PASS cannot stand in for
// a missing finite required case; the complete positive control must pass this
// exact gate even though unrelated synthetic release inputs are incomplete.
func TestOrdinaryChecksFiniteTrace(t *testing.T) {
	trace, err := ParseTraceMap([]byte(`{"schema":1,"scenarios":{"MM":"^MM01-"},"p0":["MM"],"items":[{"id":"MM-R01","source":"controlled","requirement":"finite ordinary cases","scenarios":["MM"],"requiredCases":["MM01-default-metadata-canonical"]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, complete := range []bool{false, true} {
		cases := []Case{{ID: "MM01-other", Status: Pass}}
		if complete {
			cases = append(cases, Case{ID: "MM01-default-metadata-canonical", Status: Pass})
		}
		checks := OrdinaryChecks(Inputs{Trace: trace, Offline: Bundle{Cases: cases}})
		i := slices.IndexFunc(checks, func(g Gate) bool { return g.ID == "ordinary-traceability" })
		if i < 0 {
			t.Fatal("finite ordinary trace gate missing")
		}
		if checks[i].Pass != complete {
			t.Fatalf("complete=%t gate=%+v", complete, checks[i])
		}
		if !complete && !strings.Contains(strings.Join(checks[i].Problems, " "), "MM-R01") {
			t.Fatal("missing required item was not identified")
		}
	}
}
