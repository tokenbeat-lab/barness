package ai

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/buger/jsonparser"
)

func validateMiniMaxResponsesOptions(m Model, options Options) *Error {
	o, _ := options.(ResponsesOptions)
	if problem := responsesCapabilitiesOf(m.Provider).unsupportedOption(o); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if !o.ServiceTier.IsZero() {
		tier, ok := o.ServiceTier.Get()
		if !ok || tier != "standard" {
			return newError(CodeInvalidRequest, PhaseCapability, "MiniMax Responses supports only standard service tier")
		}
	}
	if !o.ToolChoice.IsZero() {
		c, ok := o.ToolChoice.Get()
		if !ok || (c.Mode != "auto" && c.Mode != "none") || c.Function != "" || c.AllowedTools != nil {
			return newError(CodeInvalidRequest, PhaseCapability, "MiniMax Responses supports only default, auto or none tool choice")
		}
	}
	if o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort) {
		return newError(CodeInvalidRequest, PhaseCapability, "reasoning effort is not supported by this model")
	}
	return nil
}

func validateMiniMaxResponsesFinalBody(m Model, raw []byte, authorize func(map[string]any) string) *Error {
	if m.Provider != ProviderMiniMaxCN {
		return nil
	}
	var body map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if d.Decode(&body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if problem := validateMiniMaxResponsesControls(m, body); problem != "" {
		return newError(CodeInvalidRequest, PhaseRequest, problem)
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	if budget, ok := fields["max_output_tokens"]; ok && !outputBudgetNumber(budget, m.MaxTokens) {
		return newError(CodeInvalidRequest, PhaseRequest, "max_output_tokens must be an integer from 1 through the model output limit")
	}
	return nil
}

// Sampling numbers stay exact through the final schema boundary. This follows
// GLM's established wire-number preservation, but has MiniMax's own limits.
func preserveMiniMaxResponsesBudget(body []byte, m Model, level ThinkingLevel, overrides map[string]json.RawMessage) ([]byte, error) {
	if level == "" {
		level = m.DefaultThinkingLevel
		if level == "" {
			level = thinkingOff
		}
	}
	level = clampThinkingLevel(m, level)
	var budget json.RawMessage
	for _, source := range []map[string]json.RawMessage{m.SamplingParams, m.SamplingParamsByThinkingLevel[level], overrides} {
		if raw, ok := source["max_output_tokens"]; ok {
			budget = raw
		}
	}
	if budget == nil {
		return body, nil
	}
	return jsonparser.Set(body, budget, "max_output_tokens")
}
