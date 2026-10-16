package e2e

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevStateLifecycle(t *testing.T) {
	for _, name := range []string{"retry", "bad-answer", "mixed-bad-answer", "missing-usage", "same-bytes", "cancel"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC05-lifecycle-"+name+"-"+entry)
				u := newJevStateWorld(t, ev, nil)
				req := jevStateInput(t)
				if name == "same-bytes" {
					req.Images[1] = req.Images[0]
				}
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				if name == "retry" {
					enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""))
				}
				answers, usage := `"q":{"type":"noul","noul":0.5}`, `,"usage":{"input_tokens":100,"output_tokens":2}`
				if name == "bad-answer" {
					answers = `"q":{"type":"noul","noul":2}`
				}
				if name == "mixed-bad-answer" {
					req.Questions["broken"] = req.Questions["q"]
					answers += `,"broken":{"type":"noul","noul":2}`
				}
				if name == "missing-usage" {
					usage = ""
				}
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"model":"response-jev","answers":{` + answers + `}` + usage + `}`}.script(t, classifierProtocol, ""))
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				if name == "cancel" {
					cancel()
				}
				calls := 0
				hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
					calls++
					return ai.KeepPayload(), nil
				}}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctx, textScope("req-state-life-"+name+entry), jevContextTarget(), req, jevContextOptions())
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctx, textScope("req-state-life-"+name+entry), jevContextTarget(), req, jevContextOptions())
				}
				failed := name == "bad-answer" || name == "mixed-bad-answer" || name == "cancel"
				if failed {
					ev.Check("failure publishes no partial Answers or ImageInputs", err != nil && res.Answers == nil && res.ImageInputs == nil, "got %v", err)
					if name == "cancel" {
						ev.Check("canceled before entry", errors.Is(err, &ai.Error{Code: ai.CodeCanceled}), "got %v", err)
					} else {
						ev.Check("validated usage retained", res.Usage.Input == 100 && res.ResponseModel == ai.Value("response-jev"), "usage/model lost")
					}
				} else {
					ev.Check("success", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				}
				attempts := 1
				if name == "retry" {
					attempts = 2
				}
				if name == "cancel" {
					attempts = 0
				}
				ev.Check("only initial HTTP failure replays", len(res.Metadata.Attempts) == attempts && len(u.provider.Requests()) == attempts, "unexpected attempts")
				if entry == "hooked" && name != "cancel" {
					ev.Check("one payload Hook per call", calls == 1, "got %d", calls)
				}
				if name == "retry" && len(u.provider.Requests()) == 2 {
					r := u.provider.Requests()
					ev.Check("frozen bytes and headers across retry", string(r[0].Body) == string(r[1].Body) && reflect.DeepEqual(r[0].Header, r[1].Header), "retry changed")
				}
				if name == "missing-usage" && len(res.Metadata.Attempts) == 1 {
					ev.Check("no fabricated reported usage", res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "usage presence")
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}
