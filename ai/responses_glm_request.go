package ai

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/buger/jsonparser"
)

// validateResponsesModelOptions runs after authorization and simple clamping,
// before credentials. GLM's model contract rejects unsupported full efforts;
// the other providers retain their existing full-entry behavior.
func validateResponsesModelOptions(m Model, options Options) *Error {
	if m.Provider == ProviderMiniMax && m.API == APIOpenAIResponses {
		return validateMiniMaxResponsesOptions(m, options)
	}
	if m.Provider != ProviderZhipu || m.API != APIOpenAIResponses {
		return nil
	}
	o, _ := options.(ResponsesOptions)
	if problem := responsesCapabilitiesOf(m.Provider).unsupportedOption(o); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if !o.ToolChoice.IsZero() {
		choice, ok := o.ToolChoice.Get()
		if !ok || (choice.Mode != "auto" && choice.Mode != "none") || choice.Function != "" || choice.AllowedTools != nil {
			return newError(CodeInvalidRequest, PhaseCapability, "GLM Responses supports only default, none or auto tool choice")
		}
	}
	if o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort) {
		return newError(CodeInvalidRequest, PhaseCapability, "reasoning effort is not supported by this model")
	}
	return nil
}

// validateGLMResponsesFinalBody validates the actual JSON after every sampling
// layer and the trusted hook. Raw numbers preserve exact schema integer limits.
func validateGLMResponsesFinalBody(m Model, raw []byte, authorize func(map[string]any) string) *Error {
	if m.Provider != ProviderZhipu {
		return nil
	}
	var body map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if problem := validateGLMResponsesControls(m, body); problem != "" {
		return newError(CodeInvalidRequest, PhaseRequest, problem)
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	if budget, present := fields["max_output_tokens"]; present && !outputBudgetNumber(budget, min(m.MaxTokens, 131072)) {
		return newError(CodeInvalidRequest, PhaseRequest, "max_output_tokens must be an integer from 1 through the model output limit")
	}
	if rawReasoning, present := body["reasoning"]; present {
		reasoning, ok := rawReasoning.(map[string]any)
		if !ok || len(reasoning) != 1 {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM reasoning must contain only a supported effort")
		}
		effort, ok := reasoning["effort"].(string)
		valid := false
		if ok {
			for _, level := range supportedThinkingLevels(m) {
				if m.ThinkingLevelMap.wireValue(level) == effort || string(level) == effort {
					valid = true
					break
				}
			}
		}
		if !valid {
			return newError(CodeInvalidRequest, PhaseRequest, "reasoning effort is not supported by this model")
		}
	}
	for _, field := range []string{"include", "service_tier", "prompt_cache_retention", "prompt_cache_options"} {
		if _, present := body[field]; present {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Responses does not serve "+field)
		}
	}
	return nil
}

// GLM budget numbers must survive JavaScript-compatible sampling serialization
// exactly until schema validation. Other fields and other providers keep the
// established wire serialization; this provider-specific bound starts at 1.
func preserveGLMResponsesBudget(body []byte, m Model, level ThinkingLevel, overrides map[string]json.RawMessage) ([]byte, error) {
	if level == "" {
		level = thinkingOff
	}
	level = clampThinkingLevel(m, level)
	var budget json.RawMessage
	for _, source := range []map[string]json.RawMessage{m.SamplingParams, m.SamplingParamsByThinkingLevel[level], overrides} {
		if raw, present := source["max_output_tokens"]; present {
			budget = raw
		}
	}
	if budget == nil {
		return body, nil
	}
	return jsonparser.Set(body, budget, "max_output_tokens")
}
