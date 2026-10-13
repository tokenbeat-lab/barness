package e2e

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"reflect"
	"testing"
)

// The data migration is one-time; runtime LoadMatrix/Merge cannot accept schema
// 2. Nine historical rows remain byte-equivalent JSON values after promotion.
func TestDecisionsFormalMigration(t *testing.T) {
	ev := run.Case(t, "P10-OD53-formal-matrix-migration")
	before, err := supportmatrix.LoadMatrix("testdata/decisions/matrix-schema2-before.json")
	if err != nil {
		t.Fatal(err)
	}
	after, err := supportmatrix.LoadMatrix("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("before", before)
	ev.Record("after", after)
	ev.Check("schema three and independent tenth route", before.Schema == 2 && after.Schema == 3 && len(before.Rows) == 9 && len(after.Rows) == 10, "wrong schema or route count")
	if len(after.Rows) != 10 {
		return
	}
	ev.Check("nine exact history records retained", reflect.DeepEqual(before.Rows, after.Rows[:9]), "history was rewritten")
	ev.Check("new route fixed independent identity", after.Rows[9].Combo == "openai-decisions" && after.Rows[9].ValidIdentity(), "wrong identity")
	raw, err := os.ReadFile("testdata/decisions/matrix-schema3-migrated.json")
	if err != nil {
		t.Fatal(err)
	}
	var migrated supportmatrix.Matrix
	mustUnmarshal(t, raw, &migrated)
	row := migrated.Rows[9]
	ev.Check("migration initializes NOT_RUN without invented history", row.LastRunAt == nil && row.AllPassedAt == nil && len(row.Capabilities) == 9, "invented timestamps")
	for _, c := range row.Capabilities {
		ev.Check("required capability starts unexecuted", c.Outcome == supportmatrix.NotRun && c.LastRunAt == nil && c.LastPassedAt == nil, "invented result")
	}
}
