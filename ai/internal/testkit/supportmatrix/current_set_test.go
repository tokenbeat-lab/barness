package supportmatrix

import (
	"testing"
)

// Issue 17: a complete new report defines the active capability set. An
// obsolete historical probe must not survive it; narrowed/failed runs must
// still retain the prior set and cannot erase a failure.
func TestCurrentCapabilitySet(t *testing.T) {
	for _, mode := range []string{"complete", "narrowed", "failed"} {
		t.Run(mode, func(t *testing.T) {
			m := baseMatrix()
			m.Rows[1].Capabilities = append(m.Rows[1].Capabilities, Capability{ID: "obsolete-probe", Outcome: Fail, ErrorCategory: "upstream_auth", LastRunAt: &t0, LastPassedAt: &t0})
			r := report("deepseek-responses", "deepseek", "openai-responses", t1, pass("text-stream"))
			if mode == "narrowed" {
				r.Expected = append(r.Expected, "image")
			}
			if mode == "failed" {
				r.Scenarios[0].Outcome = Fail
				r.Scenarios[0].ErrorCategory = "upstream_auth"
			}
			got, err := Merge(m, r)
			if err != nil {
				t.Fatal(err)
			}
			want := 2
			if mode == "complete" {
				want = 1
			}
			if len(got.Rows[1].Capabilities) != want {
				t.Fatal("active capability set does not reflect complete versus partial evidence")
			}
			if len(m.Rows[1].Capabilities) != 2 {
				t.Fatal("caller history mutated")
			}
			if mode != "complete" && got.Rows[1].Capabilities[1].LastPassedAt != &t0 {
				t.Fatal("prior pass history discarded")
			}
		})
	}
}
