package e2e

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Failure inventory at the public release consumer boundary: the original
// seventeen-route plan must no longer authorize a complete release; omitting
// any new market/protocol proof must fail independently. Historical rows and
// all original frozen differential obligations remain required.
func TestRegionalProviderReleaseContract(t *testing.T) {
	ev := run.Case(t, "RP18-additive-release-contract")
	raw, err := os.ReadFile("../release/traceability.json")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := release.ParseTraceMap(raw)
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("release-plan.json", raw)
	ev.Check("twenty-three own routes required", len(plan.LiveCombos) == 23 && release.CheckReleasePlan(plan).Current, "routes=%d contract=%+v", len(plan.LiveCombos), release.CheckReleasePlan(plan))
	for _, combo := range []string{"zai-chat", "zai-responses", "zai-messages", "minimax-cn-chat", "minimax-cn-responses", "minimax-cn-messages"} {
		changed := plan
		changed.LiveCombos = slices.DeleteFunc(slices.Clone(plan.LiveCombos), func(v string) bool { return v == combo })
		ev.Check("own route is mandatory "+combo, slices.Contains(plan.LiveCombos, combo) && !release.CheckReleasePlan(changed).Current, "missing own route accepted")
	}
	matrix, err := supportmatrix.LoadMatrix("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("matrix", matrix)
	ev.Check("matrix preserves original and new routes", len(matrix.Rows) == 23, "rows=%d", len(matrix.Rows))
	for _, provider := range []string{"zai", "minimax-cn"} {
		for _, route := range []struct{ name, api string }{{"chat", "openai-completions"}, {"responses", "openai-responses"}, {"messages", "anthropic-messages"}} {
			at := slices.IndexFunc(matrix.Rows, func(row supportmatrix.Row) bool { return row.Combo == provider+"-"+route.name })
			ok := at >= 0 && matrix.Rows[at].ValidIdentity() && matrix.Rows[at].Provider == provider && matrix.Rows[at].API == route.api
			ev.Check("own row "+provider+"-"+route.name, ok, "missing or relabeled")
			if _, err := supportmatrix.RegionalFormalCatalog(provider + "-" + route.name); err != nil && at >= 0 {
				ev.Check("unpromoted model cannot certify support", matrix.Rows[at].AllPassedAt == nil, "premature support")
			}
		}
	}
	ev.Check("ten frozen differentials", len(plan.Differential.Required) == 10, "%v", plan.Differential.Required)
	for _, id := range []string{"P01", "P02", "P03", "P04", "P06", "P07", "P10", "MM-MESSAGES", "RP-ZAI-CHAT", "RP-MINIMAX-CN-MESSAGES"} {
		ev.Check("mandatory differential "+id, slices.Contains(plan.Differential.Required, id), "missing %s", id)
	}
	for _, prefix := range []string{"REGIONAL-R", "REGIONAL-RP"} {
		for n := 1; n <= 20; n++ {
			id := fmt.Sprintf("%s%02d", prefix, n)
			at := slices.IndexFunc(plan.Items, func(i release.TraceItem) bool { return i.ID == id })
			finite := at >= 0 && len(plan.Items[at].RequiredCases) > 0
			ev.Check("finite complete item "+id, finite, "missing finite set")
			if !finite {
				continue
			}
			changed := plan
			changed.Items = slices.Clone(plan.Items)
			changed.Items[at].RequiredCases = nil
			ev.Check("cannot shrink acceptance "+id, !release.CheckReleasePlan(changed).Current, "finite set removed")
		}
	}
	for _, route := range []release.RouteRef{{Scenario: "RP-EXTENSIONS", Provider: "zai", API: "openai-responses"}, {Scenario: "RP-EXTENSIONS", Provider: "zai", API: "anthropic-messages"}, {Scenario: "RP-EXTENSIONS", Provider: "minimax-cn", API: "openai-completions"}, {Scenario: "RP-EXTENSIONS", Provider: "minimax-cn", API: "openai-responses"}} {
		ev.Check("own exact extension "+route.Provider+"/"+route.API, slices.Contains(plan.Differential.Routes, route), "absent")
	}
}
