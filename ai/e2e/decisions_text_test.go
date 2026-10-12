package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

var decisionsProtocol = &fixtureProtocol{dir: "decisions", binding: "classifier", api: ai.API("openai-decisions"), provider: ai.ProviderOpenAI, key: func(k tenantKey) tenantKey { return k }, account: "acct-tenant-a"}

const decisionsAnswer = `{"type":"choice","name":"intent","choice":"track","probabilities":[{"value":"track","probability":0.9},{"value":"other","probability":0.1}],"confidence":0.7}`
const decisionsUsage = `"usage":{"input_tokens":296,"output_tokens":20}`

func decisionsInput() ai.ClassifierRequest {
	return ai.ClassifierRequest{State: json.RawMessage(`"track order"`), Questions: map[string]ai.ClassifierQuestion{
		"intent": ai.ChoiceQuestion{Instructions: json.RawMessage(`"Classify"`), Criteria: map[string]json.RawMessage{"track": json.RawMessage(`"Tracking"`), "other": json.RawMessage(`null`)}},
	}}
}

func newDecisionsWorld(t *testing.T, configure func(*ai.Config)) *unaryWorld {
	t.Helper()
	u := newUnaryWorld(t, func(c *ai.Config) {
		var model ai.ClassifierModel
		mustUnmarshal(t, []byte(`{"provider":"openai","api":"openai-decisions","id":"host-decisions","name":"Synthetic text classifier","input":["text"],"contextWindow":64000,"capabilities":{"kinds":["choice"],"minChoices":2,"maxChoices":255,"descriptionFormats":["string"]},"cost":{"input":1,"output":8,"cacheRead":4,"cacheWrite":4}}`), &model)
		c.Catalog.ClassifierModels = append(c.Catalog.ClassifierModels, model)
		if configure != nil {
			configure(c)
		}
	})
	u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) {
		b.ProviderID = ai.ProviderOpenAI
		b.API = decisionsProtocol.api
		b.AllowedModels = []string{"host-decisions"}
		b.Retry.MaxRetries = 1
	})
	return u
}

func TestDecisionsTextClosure(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "P10-OD16-text-"+entry)
			u := newDecisionsWorld(t, nil)
			enqueue(ev, u.world, fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-decisions"}, Body: `{"model":"response-version","answers":[` + decisionsAnswer + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
			order := []string{}
			hooks := ai.Hooks{
				TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { order = append(order, "headers"); return nil },
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					order = append(order, "payload")
					ev.Check("native API", p.API == decisionsProtocol.api && p.Operation == ai.OperationClassifier, "got %+v", p)
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					order = append(order, "response")
					return nil
				},
			}
			scope := textScope("req-decisions-" + entry)
			target := ai.Target{BindingID: "classifier", ModelID: "host-decisions"}
			var res ai.ClassifierResult
			var err error
			if entry == "plain" {
				res, err = u.client.Classify(ctxFor(t), scope, target, decisionsInput(), nil)
			} else {
				res, err = u.client.WithHooks(hooks).Classify(ctxFor(t), scope, target, decisionsInput(), nil)
			}
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("successful public HTTP closure", err == nil && res.StopReason == ai.StopReasonStop, "got %v", err)
			a, ok := res.Answers["intent"].(ai.ChoiceAnswer)
			ev.Check("typed choice and independent confidence", ok && a.Choice == "track" && a.Confidence == .7 && a.Probabilities["track"] == .9, "got %+v", res.Answers)
			requests := u.provider.Requests()
			wire := []byte(`{"model":"host-decisions","input":"track order","questions":[{"type":"choice","name":"intent","instructions":"Classify","choices":[{"value":"other"},{"value":"track","description":"Tracking"}]}]}`)
			ev.Fixture("expected-wire.json", wire)
			if ev.Check("one frozen request", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata) {
				ev.Check("native wire and fixed credential", requests[0].Path == "/v1/decisions" && requests[0].KeyAlias == tenantA.alias && jsonEqual(requests[0].Body, wire), "got %+v", requests[0])
			}
			ev.Check("known input-only usage", res.Usage.Input == 296 && res.Usage.Output == 20 && res.Usage.TotalTokens == 316 && res.Usage.Cost.Input == .000296 && res.Usage.Cost.Total == .000296 && res.Usage.Cost.Output == 0, "got %+v", res.Usage)
			if len(res.Metadata.Attempts) > 0 {
				ev.Check("usage presence is partial", res.Metadata.Attempts[0].UsageReporting == ai.UsagePartial && res.Metadata.Attempts[0].ProviderRequestID == "req-decisions", "got %+v", res.Metadata.Attempts[0])
			}
			model, _ := res.ResponseModel.Get()
			ev.Check("response identity separate", model == "response-version" && res.Metadata.ModelID == "host-decisions" && res.Metadata.API == decisionsProtocol.api, "got %+v", res.Metadata)
			want := []string{}
			if entry == "hooked" {
				want = []string{"headers", "payload", "response"}
			}
			ev.Check("callback order", reflect.DeepEqual(order, want), "got %v", order)
			reads := u.host.ReadsFor(scope.RequestID)
			ev.Record("host-reads", reads)
			ev.Check("one binding and credential", reads.Bindings == 1 && reads.Credentials == 1, "got %+v", reads)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
