package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestClassifierStrictEnvelope(t *testing.T) {
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	sc := f.Scenarios[0]
	for _, name := range []string{"duplicate-envelope", "duplicate-usage", "duplicate-token-count", "duplicate-model", "duplicate-answer", "duplicate-probability", "duplicate-field", "overflow-probability", "malformed-answers", "malformed-usage", "tie", "one-choice", "question-limit-after-callback", "state-limit-after-callback", "unknown-field-after-callback"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E02-"+name)
			w := scenarioWorld(t, sc)
			r := classifierInput(t, sc)
			reply := sc.replies(t)[0]
			body := reply.Script
			switch name {
			case "duplicate-envelope":
				body = strings.Replace(body, `"answers":{`, `"answers":{},"answers":{`, 1)
			case "duplicate-usage":
				body = strings.Replace(body, `"usage":{`, `"usage":{"input_tokens":100,"output_tokens":20},"usage":{`, 1)
			case "duplicate-token-count":
				body = strings.Replace(body, `"input_tokens":296`, `"input_tokens":100,"input_tokens":296`, 1)
			case "duplicate-model":
				body = strings.Replace(body, `"model":`, `"model":"ambiguous","model":`, 1)
			case "duplicate-answer":
				body = strings.Replace(body, `"answers":{`, `"answers":{"intent":{"type":"choice"},`, 1)
			case "duplicate-probability":
				body = strings.Replace(body, `"refund":0.1`, `"refund":0.5,"refund":0.1`, 1)
			case "duplicate-field":
				body = strings.Replace(body, `"choice":"track"`, `"choice":"refund","choice":"track"`, 1)
			case "overflow-probability":
				body = strings.Replace(body, `"track":0.9`, `"track":1e9999`, 1)
			case "malformed-answers":
				body = strings.Replace(body, `"answers":{`, `"answers":[`, 1)
			case "malformed-usage":
				body = strings.Replace(body, `"input_tokens":296`, `"input_tokens":1.5`, 1)
			case "tie":
				body = strings.ReplaceAll(body, `"refund":0.1,"track":0.9`, `"refund":0.5,"track":0.5`)
			case "one-choice":
				q := r.Questions["intent"].(ai.ChoiceQuestion)
				delete(q.Criteria, "refund")
				r.Questions["intent"] = q
				body = strings.ReplaceAll(body, `"refund":0.1,"track":0.9`, `"track":1`)
			}
			reply.Chunks = [][]byte{[]byte(body)}
			reply.Script = body
			enqueue(ev, w, reply)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "question-limit-after-callback":
					p.Body["questions"] = map[string]any{"x": map[string]any{"type": "choice", "instructions": strings.Repeat("x", 5000), "criteria": map[string]any{"x": "x"}}}
				case "state-limit-after-callback":
					p.Body["state"] = strings.Repeat("x", 5000)
				case "unknown-field-after-callback":
					p.Body["unexpected"] = true
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-strict-"+name), sc.target(), r, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			switch name {
			case "tie", "one-choice":
				ev.Check("highest probability ties and a single choice accepted", err == nil && len(res.Answers) == 1, "got %v", err)
			case "question-limit-after-callback", "state-limit-after-callback", "unknown-field-after-callback":
				ev.Check("invalid callback refused before send", errors.Is(err, &ai.Error{Code: ai.CodeCallbackFailed, Phase: ai.PhaseRequest}) && len(w.provider.Requests()) == 0, "got %v", err)
			default:
				ev.Check("invalid envelope refused atomically", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0 && len(w.provider.Requests()) == 1, "got %v", err)
			}
			switch name {
			case "duplicate-envelope", "duplicate-model":
				ev.Check("unique usage retained on envelope failure", res.Usage.Input == 296 && res.Usage.Output == 20 && len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "got %+v / %+v", res.Usage, res.Metadata)
			case "duplicate-usage", "duplicate-token-count":
				ev.Check("ambiguous usage is not reported as complete", res.Usage == (ai.Usage{}) && len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "got %+v / %+v", res.Usage, res.Metadata)
			}
		})
	}
}

func TestClassifierObservations(t *testing.T) {
	ev := run.Case(t, "P07-E09-observe")
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	sc := f.Scenarios[0]
	rec := newRecorder()
	w := observedWorld(t, rec, configureClassifier, tenantA)
	installClassifierBinding(w, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	res, err := w.client.Classify(ctxFor(t), textScope("req-classifier-observe"), sc.target(), classifierInput(t, sc), nil)
	ev.Record("result", res)
	obs := rec.awaitCall(ev, "req-classifier-observe")
	ev.Check("call succeeds", err == nil, "got %v", err)
	ev.Check("four ordered observations", len(obs) == 4, "got %+v", obs)
	for _, o := range obs {
		ev.Check("operation on every observation", o.Call.Operation == ai.OperationClassifier, "got %+v", o.Call)
	}
	if len(obs) == 4 {
		ev.Check("terminal usage and snapshot retained", obs[3].Usage == res.Usage && obs[3].Call.ModelID == "host-jev" && len(obs[3].Call.Attempts) == 1, "got %+v", obs[3])
	}
	raw := mustMarshal(t, obs)
	ev.Check("no state questions answers or error text in observations", !strings.Contains(string(raw), "track order") && !strings.Contains(string(raw), "Choose intent") && !strings.Contains(string(raw), "probabilities"), "unexpected content")
}

func TestClassifierModelAndScopeRefusals(t *testing.T) {
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	sc := f.Scenarios[0]
	for _, name := range []string{"missing-tenant", "missing-request", "foreign-tenant", "not-allowed", "not-in-catalog", "wrong-protocol", "unsupported-kind", "callback-oversized-request", "oversized-input-key"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E07-"+name)
			w := scenarioWorld(t, sc)
			scope := textScope("req-model-" + name)
			target := sc.target()
			want := ai.CodeTenantDenied
			phase := ai.PhaseCapability
			switch name {
			case "missing-tenant":
				scope.TenantID = ""
				want = ai.CodeInvalidRequest
				phase = ai.PhaseScope
			case "missing-request":
				scope.RequestID = ""
				want = ai.CodeInvalidRequest
				phase = ai.PhaseScope
			case "foreign-tenant":
				scope.TenantID = "foreign"
				want = ai.CodeBindingNotFound
				phase = ai.PhaseBinding
			case "not-allowed":
				target.ModelID = "other"
			case "not-in-catalog":
				w.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowedModels = []string{"unknown"} })
				target.ModelID = "unknown"
				want = ai.CodeInvalidRequest
			case "wrong-protocol":
				w.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.API = ai.APIOpenAIResponses })
				want = ai.CodeInvalidRequest
			case "unsupported-kind":
				cat := w.client.Catalog()
				cat.ClassifierModels[0].Capabilities = ai.ClassifierCapabilities{Kinds: []ai.ClassifierQuestionKind{ai.ClassifierQuestionBool}}
				c, err := ai.NewClient(ai.Config{Policy: func() *ai.ResourcePolicy {
					p := validPolicy()
					p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
					return p
				}(), Bindings: w.host, Credentials: w.host, Catalog: &cat, Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true})
				if err != nil {
					t.Fatal(err)
				}
				w.client = c
				want = ai.CodeInvalidRequest
			case "oversized-input-key":
				rkey := strings.Repeat("k", 1<<20)
				sc.Classifier = mustMarshal(t, map[string]any{"state": "x", "questions": map[string]any{rkey: map[string]any{"type": "choice", "instructions": "x", "criteria": map[string]any{"x": "x"}}}})
				want = ai.CodeResourceLimit
				phase = ai.PhaseScope
			case "callback-oversized-request": // covered by final-body bounds in the shared runtime suite
				want = ai.CodeResourceLimit
				phase = ai.PhaseRequest
			}
			hooks := ai.Hooks{}
			if name == "callback-oversized-request" {
				hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					p.Body["questions"] = map[string]any{strings.Repeat("k", 1<<20): map[string]any{"type": "choice", "instructions": "x", "criteria": map[string]any{"x": "x"}}}
					return ai.KeepPayload(), nil
				}
			}
			res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), scope, target, classifierInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("classification", errors.Is(err, &ai.Error{Code: want, Phase: phase}), "got %v", err)
			ev.Check("no provider request", len(w.provider.Requests()) == 0, "sent %d", len(w.provider.Requests()))
			if name != "callback-oversized-request" {
				ev.Check("no credentials read", w.host.ReadsFor(scope.RequestID).Credentials == 0, "got %+v", w.host.ReadsFor(scope.RequestID))
			}
		})
	}
}
