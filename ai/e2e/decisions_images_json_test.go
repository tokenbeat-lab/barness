package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesJSON(t *testing.T) {
	for _, sc := range []struct {
		name, images string
		valid        bool
	}{
		{"omitted", "", true}, {"null", `,"images":null`, true}, {"empty", `,"images":[]`, true},
		{"inline", `,"images":[{"data":"AAAA","mimeType":"image/png"}]`, true},
		{"object", `,"images":{}`, false}, {"element-null", `,"images":[null]`, false},
		{"string", `,"images":["url"]`, false}, {"number", `,"images":[1]`, false},
		{"data-null", `,"images":[{"data":null,"mimeType":"image/png"}]`, false},
		{"mime-null", `,"images":[{"data":"AAAA","mimeType":null}]`, false},
		{"data-number", `,"images":[{"data":1,"mimeType":"image/png"}]`, false},
		{"missing-data", `,"images":[{"mimeType":"image/png"}]`, false},
		{"missing-mime", `,"images":[{"data":"AAAA"}]`, false},
		{"unknown", `,"images":[{"data":"AAAA","mimeType":"image/png","detail":"low"}]`, false},
		{"alias", `,"images":[{"Data":"AAAA","mimeType":"image/png"}]`, false},
		{"duplicate", `,"images":[{"data":"AAAA","data":"BBBB","mimeType":"image/png"}]`, false},
		{"url", `,"images":[{"url":"https://example.invalid/x"}]`, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD18-images-json-"+sc.name)
			raw := `{"state":{},"questions":{"q":{"type":"choice","instructions":"classify","criteria":{"a":null,"b":null}}}` + sc.images + `}`
			var req ai.ClassifierRequest
			err := json.Unmarshal([]byte(raw), &req)
			ev.Fixture("input.json", []byte(raw))
			ev.Record("error", errString(err))
			ev.Check("closed public image JSON", (err == nil) == sc.valid, "got %v", err)
			if err == nil && sc.name != "inline" {
				var fields map[string]json.RawMessage
				mustUnmarshal(t, mustMarshal(t, req), &fields)
				_, present := fields["images"]
				ev.Check("no images remain omitted", !present, "got %s", mustMarshal(t, req))
			}
		})
	}
}
