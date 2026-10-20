package e2e

import (
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Public consumer failure inventory: adding a published identity must not make
// generic report metadata sufficient proof for a new service/model/account.
func TestRegionalProviderReleaseRequiresOwnCaptures(t *testing.T) {
	for _, combo := range []string{"zai-chat", "zai-responses", "zai-messages", "minimax-cn-chat", "minimax-cn-responses", "minimax-cn-messages"} {
		t.Run(combo, func(t *testing.T) {
			ev := run.Case(t, "RP18-own-captures-"+combo)
			route, _ := supportmatrix.RegionalRouteOf(combo)
			row := supportmatrix.Row{Combo: combo, Operation: "chat", Provider: route.Provider, API: route.API}
			r := supportmatrix.Report{Schema: supportmatrix.SchemaVersion, Combo: combo, Operation: "chat", Provider: route.Provider, API: route.API, Endpoint: route.Endpoint, Model: supportmatrix.RegionalModels(combo)[0], AccountAlias: "synthetic-own-account", SDK: "synthetic-sdk", GitCommit: "synthetic-source", CatalogVersion: "synthetic-catalog", CatalogHash: "synthetic-hash", StartedAt: time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC), FinishedAt: time.Date(2026, 10, 20, 0, 1, 0, 0, time.UTC), Expected: supportmatrix.RegionalExpected(combo)}
			for _, id := range r.Expected {
				r.Scenarios = append(r.Scenarios, supportmatrix.ScenarioResult{ID: id, Model: r.Model, Outcome: supportmatrix.Pass, CaseID: "synthetic-own-case", Calls: 1, Attempts: 1})
			}
			_, err := supportmatrix.Merge(supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{row}}, r)
			ev.Record("unqualified-report", r)
			ev.Check("metadata cannot certify new service", err != nil, "unowned PASS accepted")
		})
	}
}
