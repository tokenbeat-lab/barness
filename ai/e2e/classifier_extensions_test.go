package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

func TestClassifierExtensionsRegistered(t *testing.T) {
	ev := run.Case(t, "P07-E11-extensions-registered")
	ledger := loadLedger(t)
	check := func(id, path string, kind pioracle.DiffKind) {
		v := ledger.Classify(string(ai.APITypeSafeSystemOne), id, []pioracle.Diff{{Path: path, Kind: kind}})
		ev.Record(id+"-"+path, v)
		ev.Check("extension "+id+" "+path, v.Pass && len(v.Findings) == 1 && v.Findings[0].Classification == pioracle.Extension && v.Findings[0].Ref != "", "got %+v", v)
	}
	f, _ := loadFixture(t, classifierProtocol, "parity.json")
	for _, sc := range f.Scenarios {
		id := "PIDIFF-P07-" + sc.ID + "-classify"
		check(id, "request.body.model", pioracle.Changed)
		check(id, "result.usage.cost", pioracle.Changed)
		if sc.Expect.Code == "" {
			if _, ok := classifierInput(t, sc).Questions["anger"]; ok {
				check(id, "result.answers.anger.probabilities", pioracle.OnlyBarness)
				check(id, "result.answers.anger.legend", pioracle.OnlyBarness)
			}
		}
		for _, path := range []string{"request.body.unregistered", "result.unregistered", "request.headers.X-Unregistered"} {
			v := ledger.Classify(string(ai.APITypeSafeSystemOne), id, []pioracle.Diff{{Path: path, Kind: pioracle.OnlyBarness}})
			ev.Check("unknown field blocks "+path, !v.Pass && v.Pending == 1, "got %+v", v)
		}
	}
	for _, name := range []string{"string", "array", "object"} {
		id := "P07-E01-native-" + name
		for _, path := range []string{"request.body.state", "request.body.questions.intent.instructions", "request.body.questions.intent.criteria", "request.body.questions.anger.instructions", "request.body.questions.anger.criteria", "request.body.questions.urgent.criteria"} {
			check(id, path, pioracle.Changed)
		}
	}
	check("PIDIFF-P07-E11-catalog", "result.id", pioracle.Changed)
	check("PIDIFF-P07-E11-catalog", "result.cost", pioracle.Changed)
	ev.Check("entire ledger zero pending", ledger.Count(pioracle.Pending) == 0, "got %d", ledger.Count(pioracle.Pending))
}
