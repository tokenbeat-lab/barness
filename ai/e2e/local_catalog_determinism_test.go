package e2e

import (
	"github.com/tokenbeat-lab/barness/catalog"
	"testing"
)

func TestLocalCatalogDeterministicDiagnostic(t *testing.T) {
	ev := run.Case(t, "CAT-T08-deterministic-invalid-definition-diagnostic")
	raw := localCatalogFile(`{"version":"deterministic","models":[{"provider":"openai","api":"openai-responses","id":"bad","name":null,"zzUnknown":true}]}`)
	ev.Fixture("catalog.json", raw)
	diagnostics := map[string]int{}
	for i := 0; i < 128; i++ {
		_, report, err := catalog.Load(raw)
		if !ev.Check("same invalid model remains isolated", err == nil && report.Failed == 1 && len(report.Models) == 1, "iteration=%d err=%v report=%+v", i, err, report) {
			return
		}
		status := report.Models[0]
		diagnostics[string(status.Reason)+":"+status.Diagnostic]++
	}
	ev.Record("observed-diagnostics", diagnostics)
	ev.Check("identical bytes have one stable primary reason and diagnostic", len(diagnostics) == 1, "diagnostics=%+v", diagnostics)
}
