package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

// Current route defaults follow today's stored release snapshot. Historical
// reports keep their recorded catalogs and cannot acquire current applicability.
func TestChineseEvaluationCurrentRoutePins(t *testing.T) {
	for _, tc := range []struct{ route, model string }{{"jev", "jev-1.13.0"}, {"luna", "gpt-6-luna"}} {
		t.Run(tc.route, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-current-"+tc.route)
			out := t.TempDir()
			output, err := evaluationCommand(t, "-route", tc.route, "-out", out)
			ev.Record("disabled-current-route-process", map[string]any{"exitZero": err == nil, "output": string(output), "actualVendorRequests": 0})
			if !ev.Check("current default config loads before disabled execution", err == nil, "%v", err) {
				return
			}
			dir, _ := evaluationBundle(t, out, ev)
			raw, err := os.ReadFile(filepath.Join(dir, "evaluation-report.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report chineseeval.Report
			mustUnmarshal(t, raw, &report)
			raw, err = os.ReadFile("../release/catalog-snapshot.json")
			if err != nil {
				t.Fatal(err)
			}
			var pin struct {
				Hash    string
				Catalog struct{ Version string }
			}
			mustUnmarshal(t, raw, &pin)
			ev.Check("disabled workflow keeps current catalog and exact route", report.Status == "NOT_RUN" && report.Budget.Calls == 0 && report.Config.Model == tc.model && report.Config.CatalogVersion == pin.Catalog.Version && report.Config.CatalogHash == pin.Hash, "current configuration or zero-call outcome differs")
		})
	}
}
