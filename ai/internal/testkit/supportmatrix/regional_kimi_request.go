package supportmatrix

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func kimiRegionalBody(raw string) (map[string]any, error) {
	var body map[string]any
	d := json.NewDecoder(strings.NewReader(raw))
	d.UseNumber()
	if d.Decode(&body) != nil || body == nil {
		return nil, refuse("Kimi captured request is not an object")
	}
	return body, nil
}

// Scenario controls are checked from captured bytes, independently of the
// saved options and successful response. Resealing a modified capture cannot
// turn a different reasoning or tool scenario into its declared qualification.
func verifyKimiRegionalControls(r Report, s ScenarioResult, raw string) error {
	body, err := kimiRegionalBody(raw)
	if err != nil {
		return err
	}
	_, kind, _ := strings.Cut(s.ID, "/")
	effort := ""
	switch kind {
	case "full-complete", "simple-complete":
		effort = "high"
	case "simple-stream":
		effort = "low"
	}
	effortField, budgetField := "reasoning_effort", "max_completion_tokens"
	var wantEffort any = effort
	if r.API == "openai-responses" {
		effortField, budgetField = "reasoning", "max_output_tokens"
		wantEffort = map[string]any{"effort": effort}
	}
	if r.API == "anthropic-messages" {
		effortField, budgetField = "output_config", "max_tokens"
		wantEffort = map[string]any{"effort": effort}
	}
	got, present := body[effortField]
	if (effort == "" && present) || (effort != "" && (!present || !reflect.DeepEqual(got, wantEffort))) {
		return refuse("Kimi captured effort differs from declared scenario")
	}
	if body[budgetField] != json.Number("4096") {
		return refuse("Kimi captured output budget differs from finite scenario")
	}
	choice, present := body["tool_choice"]
	if kind == "tool-call" {
		switch r.API {
		case "openai-completions":
			if choice != "required" {
				return refuse("Kimi captured tool scenario requires required choice")
			}
		case "anthropic-messages":
			if !reflect.DeepEqual(choice, map[string]any{"type": "any"}) {
				return refuse("Kimi captured tool scenario requires any choice")
			}
		default:
			if present && choice != "auto" {
				return refuse("Kimi captured tool scenario requires auto choice")
			}
		}
	} else if present {
		return refuse("Kimi captured choice differs from declared text scenario")
	}
	return nil
}

// Feed the actual captured native payload through the public Client's final
// guards. Compare its controls/tools with the request generated from saved
// options first; replay may use a synthetic prompt but may not invent proof
// of the controls the actual process sent.
func kimiRegionalReplayHook(captured string) ai.Hooks {
	return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		body, err := kimiRegionalBody(captured)
		if err != nil {
			return ai.KeepPayload(), err
		}
		fields := []string{"model", "stream", "max_completion_tokens", "max_output_tokens", "max_tokens", "reasoning_effort", "reasoning", "output_config", "tool_choice", "tools", "prompt_cache_key", "prompt_cache_options", "cache_control", "temperature", "top_p", "n", "presence_penalty", "frequency_penalty"}
		for _, field := range fields {
			a, hasA := p.Body[field]
			b, hasB := body[field]
			// Omitted and auto are the same integrated Responses tool choice.
			if field == "tool_choice" && p.API == ai.APIOpenAIResponses {
				if a == "auto" {
					hasA = false
				}
				if b == "auto" {
					hasB = false
				}
			}
			if hasA != hasB || hasA && !reflect.DeepEqual(a, b) {
				return ai.KeepPayload(), refuse("Kimi captured request contradicts saved options or tools")
			}
		}
		return ai.ReplacePayload(body), nil
	}}
}

func kimiRegionalSameWire(a []byte, b string) bool {
	left, e1 := kimiRegionalBody(string(bytes.TrimSpace(a)))
	right, e2 := kimiRegionalBody(b)
	return e1 == nil && e2 == nil && reflect.DeepEqual(left, right)
}
