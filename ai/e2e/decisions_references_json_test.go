package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsReferencesJSON(t *testing.T) {
	for _, sc := range []struct {
		name, refs string
		valid      bool
	}{
		{"zero", `[{"imageIndex":0,"placement":"context"}]`, true},
		{"value", `[{"imageIndex":0,"placement":"value","pointer":"/state/x"}]`, true},
		{"missing-index", `[{"placement":"context"}]`, false},
		{"null-index", `[{"imageIndex":null,"placement":"context"}]`, false},
		{"negative-index", `[{"imageIndex":-1,"placement":"context"}]`, false},
		{"fraction-index", `[{"imageIndex":0.5,"placement":"context"}]`, false},
		{"float-index", `[{"imageIndex":0.0,"placement":"context"}]`, false},
		{"overflow-index", `[{"imageIndex":99999999999999999999999,"placement":"context"}]`, false},
		{"string-index", `[{"imageIndex":"0","placement":"context"}]`, false},
		{"missing-placement", `[{"imageIndex":0}]`, false},
		{"null-placement", `[{"imageIndex":0,"placement":null}]`, false},
		{"unknown-placement", `[{"imageIndex":0,"placement":"state"}]`, false},
		{"pointer-in-context", `[{"imageIndex":0,"placement":"context","pointer":"/state"}]`, false},
		{"empty-pointer-in-context", `[{"imageIndex":0,"placement":"context","pointer":""}]`, false},
		{"null-pointer-in-context", `[{"imageIndex":0,"placement":"context","pointer":null}]`, false},
		{"missing-value-pointer", `[{"imageIndex":0,"placement":"value"}]`, false},
		{"null-value-pointer", `[{"imageIndex":0,"placement":"value","pointer":null}]`, false},
		{"empty-value-pointer", `[{"imageIndex":0,"placement":"value","pointer":""}]`, false},
		{"unknown-field", `[{"imageIndex":0,"placement":"context","data":"x"}]`, false},
		{"alias", `[{"ImageIndex":0,"placement":"context"}]`, false},
		{"duplicate", `[{"imageIndex":0,"imageIndex":1,"placement":"context"}]`, false},
		{"duplicate-placement", `[{"imageIndex":0,"placement":"context","placement":"value"}]`, false},
		{"null-record", `[null]`, false}, {"object", `{}`, false}, {"scalar", `[1]`, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC02-json-"+sc.name)
			raw := `{"state":{},"questions":{"q":{"type":"bool","instructions":"judge"}},"imageReferences":` + sc.refs + `}`
			var req ai.ClassifierRequest
			err := json.Unmarshal([]byte(raw), &req)
			ev.Fixture("public-request.json", []byte(raw))
			ev.Record("decode-error", errString(err))
			ev.Check("closed reference JSON", (err == nil) == sc.valid, "got %v", err)
			if err == nil {
				encoded := mustMarshal(t, req)
				var fields map[string]json.RawMessage
				mustUnmarshal(t, encoded, &fields)
				ev.Check("roundtrip preserves required zero and variant", jsonEqual(fields["imageReferences"], []byte(sc.refs)), "got %s", fields["imageReferences"])
			}
		})
	}
	for _, ref := range []ai.ImageReference{
		{ImageIndex: -1, Placement: ai.ImagePlacementContext}, {ImageIndex: 0, Placement: "unknown"},
		{ImageIndex: 0, Placement: ai.ImagePlacementContext, Pointer: "/state"}, {ImageIndex: 0, Placement: ai.ImagePlacementValue},
	} {
		if _, err := json.Marshal(ref); err == nil {
			t.Error("invalid typed reference serialized")
		}
	}
}
