package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// This helper goes through the public codec, so fixture spelling is also
// covered by every call that uses references.
func decisionsReferenceRequest(t *testing.T, req ai.ClassifierRequest, refs string) ai.ClassifierRequest {
	t.Helper()
	var fields map[string]json.RawMessage
	mustUnmarshal(t, mustMarshal(t, req), &fields)
	fields["imageReferences"] = json.RawMessage(refs)
	var out ai.ClassifierRequest
	if err := json.Unmarshal(mustMarshal(t, fields), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDecisionsReferencesOrder(t *testing.T) {
	for _, sc := range []struct {
		name, refs string
		order      []int
	}{
		{"omitted", "", []int{0, 1}}, {"null", `null`, []int{0, 1}}, {"empty", `[]`, []int{0, 1}},
		{"same-bytes", `[{"imageIndex":1,"placement":"context"},{"imageIndex":0,"placement":"context"},{"imageIndex":1,"placement":"context"}]`, []int{1, 0, 1}},
		{"reuse", `[{"imageIndex":1,"placement":"context"},{"imageIndex":0,"placement":"context"},{"imageIndex":1,"placement":"context"}]`, []int{1, 0, 1}},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC02-order-"+sc.name)
			u := newClassifierImagesWorld(t, nil)
			req := decisionsInput()
			req.Images = classifierAssets(t)[:2]
			if sc.name == "same-bytes" {
				req.Images[1] = req.Images[0]
			}
			if sc.refs != "" {
				req = decisionsReferenceRequest(t, req, sc.refs)
			}
			// Exercise the same contract again after a public JSON round-trip.
			var round ai.ClassifierRequest
			mustUnmarshal(t, mustMarshal(t, req), &round)
			enqueueClassifierImages(t, ev, u)
			ev.Fixture("public-request.json", mustMarshal(t, round))
			recordClassifierAssets(ev, classifierAssets(t))
			res, err := u.client.Classify(ctxFor(t), textScope("req-reference-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, round, nil)
			ev.Check("context reference call succeeds", err == nil, "got %v", err)
			requests := u.provider.Requests()
			if ev.Check("one final request", len(requests) == 1, "got %d", len(requests)) {
				var body map[string]any
				mustUnmarshal(t, requests[0].Body, &body)
				ordered := make([]ai.Image, len(sc.order))
				for i, index := range sc.order {
					ordered[i] = req.Images[index]
				}
				ev.Check("State first; reference order and repeated uses survive", jsonEqual(mustMarshal(t, body["input"]), mustMarshal(t, classifierImageMessages(ordered))), "unexpected input")
			}
			ev.Record("result", res)
			ev.Record("error", errString(err))
			recordRuntimeConfig(ev, u)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
