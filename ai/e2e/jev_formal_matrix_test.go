package e2e

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestJevFormalMatrixHistory(t *testing.T) {
	ev := run.Case(t, "UC13-model-specific-matrix-history")
	before, err := supportmatrix.LoadMatrix("testdata/typesafe/formal-matrix-before-1.14.json")
	if err != nil {
		t.Fatal(err)
	}
	current, err := supportmatrix.LoadMatrix("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	i := slices.IndexFunc(current.Rows, func(r supportmatrix.Row) bool { return r.Combo == "typesafe-classifier-jev-1.14.0" })
	if !ev.Check("independent fixed 1.14 matrix identity exists", i >= 0, "new model row absent") {
		return
	}
	ev.Check("fixed model route cannot alias historical identity", current.Rows[i].Model == "jev-1.14.0" && current.Rows[i].ValidIdentity(), "wrong new identity")
	current.Rows = slices.Delete(slices.Clone(current.Rows), i, i+1)
	ev.Check("all historical rows states and timestamps preserved", reflect.DeepEqual(current, before), "historical support rewritten")
	raw, err := os.ReadFile("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	var v any
	json.Unmarshal(raw, &v)
	ev.Record("matrix", v)
}
