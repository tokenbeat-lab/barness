package e2e

import (
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
			ok := at >= 0 && matrix.Rows[at].ValidIdentity() && matrix.Rows[at].Provider == provider && matrix.Rows[at].API == route.api && matrix.Rows[at].AllPassedAt == nil
			ev.Check("own pending row "+provider+"-"+route.name, ok, "missing, relabeled or prematurely certified")
		}
	}
	ev.Check("all eight frozen differentials retained", len(plan.Differential.Required) == 8 && slices.Contains(plan.Differential.Required, "MM-MESSAGES"), "%v", plan.Differential.Required)
}
