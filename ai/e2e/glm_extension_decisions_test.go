package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// Configuration acceptance: these records describe own GLM fixture behavior,
// never fabricated pi observations. Public behavior lives in the referenced
// cases and is re-run by ticket10's exact acceptance command.
func TestGLMExtensionDecisions(t *testing.T) {
	ev := run.Case(t, "GLM10-case-path-ledger")
	raw, err := os.ReadFile("testdata/glm-pidiff/ledger.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("glm-acceptance-ledger.json", raw)
	ledger, err := pioracle.ParseLedger(raw)
	if err != nil {
		t.Fatal(err)
	}
	var file struct{ Decisions []pioracle.Decision }
	mustUnmarshal(t, raw, &file)
	ev.Check("own case-path decisions are present with zero pending", len(file.Decisions) > 0 && ledger.Count(pioracle.Pending) == 0, "%d", len(file.Decisions))
	for _, d := range file.Decisions {
		ev.Check("decision requires exact cases and GLM protocol", strings.HasPrefix(d.Protocol, "zhipu/") && len(d.Cases) > 0 && !strings.Contains(d.Path, "[*]"), "%+v", d)
		for _, id := range d.Cases {
			verdict := ledger.Classify(d.Protocol, id, []pioracle.Diff{{Path: d.Path, Kind: pioracle.Changed}})
			if d.Classification == pioracle.Fixed {
				ev.Check("fixed behavior cannot excuse a recurrence", !verdict.Pass && verdict.Pending == 1, "%+v", verdict)
			} else {
				ev.Check("reviewed case-path is bounded extension", verdict.Pass && len(verdict.Findings) == 1 && verdict.Findings[0].Classification == pioracle.Extension, "%+v", verdict)
			}
			for _, probe := range []struct{ protocol, id, path string }{
				{d.Protocol, "GLM10-unreviewed-case", d.Path},
				{d.Protocol, id, "request.body.unreviewed"},
				{strings.TrimPrefix(d.Protocol, "zhipu/"), id, d.Path},
			} {
				verdict := ledger.Classify(probe.protocol, probe.id, []pioracle.Diff{{Path: probe.path, Kind: pioracle.Changed}})
				ev.Check("unreviewed case/path/native protocol stays pending", !verdict.Pass && verdict.Pending == 1, "%+v", verdict)
			}
		}
	}
}
