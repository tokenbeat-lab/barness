package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevContextLifecycle(t *testing.T) {
	for _, name := range []string{"retry", "bad-answer", "bad-usage", "null-usage", "missing-usage", "response-hook", "cancel", "text", "same-source-bytes", "renamed"} {
		for _, entry := range []string{"plain", "hooked"} {
			if entry == "plain" && (name == "response-hook" || name == "renamed") {
				continue
			}
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC04-lifecycle-"+name+"-"+entry)
				u := newJevContextWorld(t, ev, nil)
				req := jevContextInput(t)
				want := req.Images
				key := "q"
				if name == "same-source-bytes" {
					req.Images[1] = req.Images[0]
				}
				if name == "text" {
					req.Images = nil
					want = nil
				}
				if name == "renamed" {
					key = "renamed"
				}
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				if name == "retry" {
					enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""))
				}
				usage := `,"usage":{"input_tokens":100,"output_tokens":2}`
				answer := `{"type":"noul","noul":0.5}`
				switch name {
				case "bad-answer":
					answer = `{"type":"noul","noul":2}`
				case "bad-usage":
					usage = `,"usage":{"input_tokens":-1}`
				case "null-usage":
					usage = `,"usage":null`
				case "missing-usage":
					usage = ""
				}
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"model":"response-jev","answers":{"` + key + `":` + answer + `}` + usage + `}`}.script(t, classifierProtocol, ""))
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				hooks := ai.Hooks{}
				calls := [3]int{}
				hooks.TransformHeaders = func(context.Context, ai.CallScope, http.Header) error { calls[0]++; return nil }
				hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					calls[1]++
					if name == "renamed" {
						p.Body["questions"] = map[string]any{key: map[string]any{"type": "noul", "instructions": "Ask"}}
					}
					return ai.KeepPayload(), nil
				}
				hooks.OnResponse = func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					calls[2]++
					if name == "response-hook" {
						return errors.New("private callback")
					}
					return nil
				}
				if name == "cancel" {
					cancel()
				}
				scope := textScope("req-jev-life-" + name + entry)
				var res ai.ClassifierResult
				var err error
				opts := jevContextOptions()
				if name == "text" {
					opts = ai.TypeSafeOptions{}
				}
				if entry == "plain" {
					res, err = u.client.Classify(ctx, scope, jevContextTarget(), req, opts)
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctx, scope, jevContextTarget(), req, opts)
				}
				failed := name == "bad-answer" || name == "bad-usage" || name == "response-hook" || name == "cancel"
				if failed {
					ev.Check("failure is atomic", err != nil && res.Answers == nil && res.ImageInputs == nil, "got %v", err)
				} else {
					ev.Check("success", err == nil && len(res.Answers) == 1, "got %v", err)
					checkImageInputs(ev, res, want)
				}
				if name == "bad-answer" {
					ev.Check("reported consumption survives bad answer", res.Usage.Input == 100 && res.ResponseModel == ai.Value("response-jev"), "usage/model lost")
				}
				if name == "null-usage" || name == "missing-usage" {
					ev.Check("usage not fabricated", len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "usage presence")
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
					ev.Check("hooks once per Logical Call", calls == [3]int{1, 1, 1}, "got %v", calls)
				}
				captures := u.provider.Requests()
				if name == "retry" && len(captures) == 2 {
					ev.Check("retry bytes headers credential fixed", string(captures[0].Body) == string(captures[1].Body) && reflect.DeepEqual(captures[0].Header, captures[1].Header) && captures[0].KeyAlias == tenantA.alias, "retry snapshot changed")
				}
				if name == "text" && len(captures) == 1 {
					var body map[string]json.RawMessage
					mustUnmarshal(t, captures[0].Body, &body)
					ev.Check("text native State and no assets/experiment", jsonEqual(body["state"], req.State) && body["assets"] == nil && captures[0].Header["Typesafe-Danger-Zone-Images"] == "", "text wire changed")
				}
				recordJevContext(ev, u, req, res, err)
				obs := observationText(t, u.rec.all())
				for _, capture := range captures {
					var body struct {
						Assets map[string]any `json:"assets"`
					}
					mustUnmarshal(t, capture.Body, &body)
					for key := range body.Assets {
						ev.Check("private key excluded from result and Observer", !strings.Contains(string(mustMarshal(t, res)), key) && !strings.Contains(obs, key), "key leaked")
					}
				}
			})
		}
	}
}
