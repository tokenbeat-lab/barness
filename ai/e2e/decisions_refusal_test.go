package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsRefusal(t *testing.T) {
	for _, extra := range []string{"", `,"choice":"track"`, `,"confidence":1`, `,"probabilities":[]`, `,"reason":"invented"`} {
		name := "valid"
		if extra != "" {
			name = strings.Split(strings.TrimPrefix(extra, `,"`), `"`)[0]
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD28-refusal-"+name+"-hooked")
			u := newDecisionsWorld(t, nil)
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[{"type":"refusal","name":"intent"` + extra + `}],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
			res, err := u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-refusal-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if extra == "" {
				ev.Check("refusal is successful typed result", err == nil && res.StopReason == ai.StopReasonStop && len(res.Answers) == 1, "got %v", err)
				ev.Check("refusal JSON has no fabricated fields", jsonEqual(mustMarshal(t, res.Answers), []byte(`{"intent":{"type":"refusal"}}`)), "got %+v", res.Answers)
			} else {
				ev.Check("damaged refusal rejected atomically", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0, "got %v", err)
			}
			ev.Check("known usage kept", res.Usage.Input == 296 && res.Usage.Cost.Total == .000296, "got %+v", res.Usage)
			ev.Check("never replay 2xx", len(u.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
