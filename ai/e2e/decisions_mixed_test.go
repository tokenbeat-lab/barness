package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsMixedClosure(t *testing.T) {
	for _, name := range []string{"ordinary", "mixed-refusal", "all-refusal", "metadata", "refusal-metadata"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD27-mixed-"+name+"-"+entry)
				u := newDecisionsMixedWorld(t, nil)
				answers, expected := decisionsScore+`,`+decisionsAnswer+`,`+decisionsPredicate, decisionsMixedAnswers
				switch name {
				case "mixed-refusal":
					answers = decisionsScore + `,{"type":"refusal","name":"intent"},` + decisionsPredicate
					expected = `{"approved":{"type":"bool","probability":0.37},"intent":{"type":"refusal"},"urgency":{"type":"score","score":1.1,"confidence":0.6,"probabilities":{"0":0.2,"1":0.5,"2":0.3},"legend":{"0":"低","1":"低","2":"高"}}}`
				case "all-refusal", "refusal-metadata":
					answers = `{"type":"refusal","name":"urgency"},{"type":"refusal","name":"approved"},{"type":"refusal","name":"intent"}`
					expected = `{"urgency":{"type":"refusal"},"approved":{"type":"refusal"},"intent":{"type":"refusal"}}`
					if name == "refusal-metadata" {
						answers = strings.ReplaceAll(answers, `"type":`, `"vendor_metadata":{"trace":"synthetic"},"type":`)
					}
				case "metadata":
					answers = strings.ReplaceAll(answers, `"type":`, `"vendor_metadata":{"trace":"synthetic"},"type":`)
					answers = strings.ReplaceAll(answers, `"value":`, `"vendor_rank":42,"value":`)
				}
				body := `{"vendor_metadata":{"version":1},"answers":[` + answers + `],` + decisionsUsage + `}`
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: body}.script(t, decisionsProtocol, ""))
				req := decisionsMixedInput()
				original := mustMarshal(t, req)
				ev.Fixture("public-input.json", original)
				ev.Fixture("expected-wire.json", []byte(decisionsMixedWire))
				ev.Fixture("expected-answers.json", []byte(expected))
				res, err := decisionsPublicCall(t, u, entry, "req-mixed-"+name+"-"+entry, req, decisionsNoopHooks())
				ev.Record("result", res)
				ev.Record("error", errString(err))
				ev.Check("complete atomic typed result", err == nil && res.StopReason == ai.StopReasonStop && jsonEqual(mustMarshal(t, res.Answers), []byte(expected)), "got %+v / %v", res.Answers, err)
				ev.Check("public input untouched", jsonEqual(mustMarshal(t, req), original), "input changed")
				requests := u.provider.Requests()
				if ev.Check("one attempt for every 2xx", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata) {
					ev.Check("three native kinds and ordered duplicate labels", jsonEqual(requests[0].Body, []byte(decisionsMixedWire)), "got %s", requests[0].Body)
				}
				ev.Check("usage retained for refusals", res.Usage.Input == 296 && res.Usage.Output == 20 && res.Usage.Cost.Total == .000296, "got %+v", res.Usage)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsPredicateInstructions(t *testing.T) {
	for _, sc := range []struct{ name, yes, no, wire string }{
		{"absent", "", "", "Approved?"},
		{"empty", `""`, `""`, "Approved?"},
		{"true-only", `"同意"`, `""`, "Approved?\n\nTrue means: 同意"},
		{"false-only", `""`, `"不同意"`, "Approved?\n\nFalse means: 不同意"},
		{"both", `"同意"`, `"不同意"`, "Approved?\n\nTrue means: 同意\nFalse means: 不同意"},
	} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD13-predicate-"+sc.name+"-"+entry)
				u := newDecisionsMixedWorld(t, nil)
				q := ai.BoolQuestion{Instructions: json.RawMessage(`"Approved?"`)}
				if sc.yes != "" {
					q.Criteria = &ai.BoolCriteria{True: json.RawMessage(sc.yes), False: json.RawMessage(sc.no)}
				}
				req := ai.ClassifierRequest{State: json.RawMessage(`{}`), Questions: map[string]ai.ClassifierQuestion{"approved": q}}
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + decisionsPredicate + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
				res, err := decisionsPublicCall(t, u, entry, "req-predicate-"+sc.name+"-"+entry, req, decisionsNoopHooks())
				a, ok := res.Answers["approved"].(ai.BoolAnswer)
				ev.Check("probability remains unthresholded", err == nil && ok && a.Probability == .37, "got %+v / %v", a, err)
				requests := u.provider.Requests()
				wire := `{"model":"host-decisions","input":"{}","questions":[{"type":"predicate","name":"approved","instructions":` + string(mustMarshal(t, sc.wire)) + `}]}`
				ev.Fixture("expected-wire.json", []byte(wire))
				if ev.Check("one request", len(requests) == 1, "got %d", len(requests)) {
					ev.Check("fixed criteria concatenation", jsonEqual(requests[0].Body, []byte(wire)), "got %s", requests[0].Body)
				}
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}
