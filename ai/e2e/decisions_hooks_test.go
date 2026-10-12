package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsFinalQuestions(t *testing.T) {
	for _, entry := range []string{"keep", "replace", "old-answer"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "P10-OD22-final-"+entry+"-hooked")
			u := newDecisionsWorld(t, nil)
			answer := strings.ReplaceAll(decisionsAnswer, `"intent"`, `"changed"`)
			if entry == "old-answer" {
				answer = decisionsAnswer
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + answer + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
			public := decisionsInput()
			var borrowed map[string]any
			payloads := 0
			res, err := u.client.WithHooks(ai.Hooks{
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					payloads++
					borrowed = p.Body
					q := borrowed["questions"].([]any)[0].(map[string]any)
					q["name"] = "changed"
					q["instructions"] = "New question"
					borrowed["input"] = "New state"
					if entry == "replace" {
						return ai.ReplacePayload(borrowed), nil
					}
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					// A synchronous barrier after send proves references are no
					// longer borrowed without racing the request encoder.
					borrowed["questions"].([]any)[0].(map[string]any)["name"] = "late"
					borrowed["input"] = "late"
					public.Questions["intent"] = nil
					public.State[1] = 'X'
					return nil
				},
			}).Classify(ctxFor(t), textScope("req-final-"+entry), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, public, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("payload once", payloads == 1, "got %d", payloads)
			if entry == "old-answer" {
				ev.Check("old collection refused", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0, "got %v", err)
			} else {
				a, ok := res.Answers["changed"].(ai.ChoiceAnswer)
				ev.Check("final names define answers", err == nil && ok && a.Choice == "track" && len(res.Answers) == 1, "got %v", err)
			}
			requests := u.provider.Requests()
			if ev.Check("one request", len(requests) == 1, "got %d", len(requests)) {
				var body map[string]any
				mustUnmarshal(t, requests[0].Body, &body)
				ev.Check("body independent of late references", body["input"] == "New state" && body["questions"].([]any)[0].(map[string]any)["name"] == "changed", "got %s", requests[0].Body)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsHookGuard(t *testing.T) {
	cases := []struct {
		name   string
		change func(map[string]any)
		code   ai.Code
	}{
		{"model", func(p map[string]any) { p["model"] = "foreign" }, ai.CodeTenantDenied},
		{"unknown", func(p map[string]any) { p["temperature"] = .5 }, ai.CodeCallbackFailed},
		{"safety", func(p map[string]any) { p["safety_identifier"] = "id" }, ai.CodeCallbackFailed},
		{"input-null", func(p map[string]any) { p["input"] = nil }, ai.CodeCallbackFailed},
		{"input-messages", func(p map[string]any) {
			p["input"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_text", "text": "ok"}}}}
		}, ai.CodeCallbackFailed},
		{"input-external-image", func(p map[string]any) {
			p["input"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_image", "image_url": "https://foreign.invalid/a.png"}}}}
		}, ai.CodeTenantDenied},
		{"input-file", func(p map[string]any) { p["input"] = []any{map[string]any{"file_id": "foreign"}} }, ai.CodeTenantDenied},
		{"input-inline-image", func(p map[string]any) {
			p["input"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_image", "image_url": "data:image/png;base64,AAAA"}}}}
		}, ai.CodeCallbackFailed},
		{"questions-null", func(p map[string]any) { p["questions"] = nil }, ai.CodeCallbackFailed},
		{"questions-empty", func(p map[string]any) { p["questions"] = []any{} }, ai.CodeCallbackFailed},
	}
	for _, key := range []string{"endpoint", "api_key", "headers", "provider", "api", "operation", "credential", "tools", "allowClassifierImages", "allow_classifier_images"} {
		cases = append(cases, struct {
			name   string
			change func(map[string]any)
			code   ai.Code
		}{key, func(p map[string]any) { p[key] = "forbidden" }, ai.CodeTenantDenied})
	}
	for _, name := range []string{"score", "predicate", "unknown-kind", "name-null", "name-empty", "name-duplicate", "instruction-null", "instruction-empty", "choices-null", "one-choice", "duplicate-choice", "boolean-choice", "description-null", "description-object", "unknown-question", "question-case", "unknown-choice", "choice-case"} {
		cases = append(cases, struct {
			name   string
			change func(map[string]any)
			code   ai.Code
		}{name, func(p map[string]any) {
			q := p["questions"].([]any)[0].(map[string]any)
			choices := q["choices"].([]any)
			switch name {
			case "score", "predicate":
				q["type"] = name
			case "unknown-kind":
				q["type"] = "bool"
			case "name-null":
				q["name"] = nil
			case "name-empty":
				q["name"] = ""
			case "name-duplicate":
				p["questions"] = []any{q, q}
			case "instruction-null":
				q["instructions"] = nil
			case "instruction-empty":
				q["instructions"] = ""
			case "choices-null":
				q["choices"] = nil
			case "one-choice":
				q["choices"] = choices[:1]
			case "duplicate-choice":
				q["choices"] = []any{choices[0], choices[0]}
			case "boolean-choice":
				choices[0].(map[string]any)["value"] = true
			case "description-null":
				choices[0].(map[string]any)["description"] = nil
			case "description-object":
				choices[0].(map[string]any)["description"] = map[string]any{"x": 1}
			case "unknown-question":
				q["metadata"] = true
			case "question-case":
				q["Name"] = "intent"
			case "unknown-choice":
				choices[0].(map[string]any)["label"] = "bad"
			case "choice-case":
				choices[0].(map[string]any)["Value"] = "other"
			}
		}, ai.CodeCallbackFailed})
	}
	for _, sc := range cases {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD23-guard-"+sc.name+"-hooked")
			u := newDecisionsWorld(t, nil)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				sc.change(p.Body)
				ev.Record("final-payload", p.Body)
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-guard-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			checkUnaryFailure(ev, res, err, sc.code, ai.PhaseRequest, ai.StopReasonError)
			ev.Check("no attempt or transmission", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0 && len(u.adm.Requests()) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsHeaderGuard(t *testing.T) {
	for _, name := range []string{"Authorization", "Host", "Openai-Project", "X-Api-Key"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD23-header-"+name+"-hooked")
			u := newDecisionsWorld(t, nil)
			res, err := u.client.WithHooks(ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error { h.Set(name, "foreign"); return nil }}).Classify(ctxFor(t), textScope("req-header-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			checkUnaryFailure(ev, res, err, ai.CodeTenantDenied, ai.PhaseRequest, ai.StopReasonError)
			ev.Check("zero attempt", len(res.Metadata.Attempts) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

// Raw replacements exercise duplicate JSON keys through the public Hook.
type decisionsRaw json.RawMessage

func (r decisionsRaw) MarshalJSON() ([]byte, error) { return []byte(r), nil }

func TestDecisionsHookDuplicateFields(t *testing.T) {
	for _, name := range []string{"name", "value"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD25-duplicate-"+name+"-hooked")
			u := newDecisionsWorld(t, nil)
			q := `{"type":"choice","name":"intent","instructions":"Classify","choices":[{"value":"other"},{"value":"track"}]}`
			if name == "name" {
				q = strings.Replace(q, `"name":"intent"`, `"name":"intent","name":"intent"`, 1)
			} else {
				q = strings.Replace(q, `"value":"track"`, `"value":"track","value":"track"`, 1)
			}
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["questions"] = []any{decisionsRaw(q)}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-duplicate-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			checkUnaryFailure(ev, res, err, ai.CodeCallbackFailed, ai.PhaseRequest, ai.StopReasonError)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}
