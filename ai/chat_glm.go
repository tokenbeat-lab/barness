package ai

import (
	"encoding/json"
	"math"
	"slices"
)

// China ChatCompletionTextRequest.max_tokens.maximum, retrieved 2026-10-17:
// https://docs.bigmodel.cn/openapi/openapi.json. Models may narrow this cap.
const glmChatMaxOutputTokens = 131072

// validateChatModelOptions runs after binding/model authorization and simple
// clamping, before credentials. GLM's full entry rejects unsupported named
// levels; OpenAI/DeepSeek retain their existing unclamped full-entry contract.
func validateChatModelOptions(m Model, options Options) *Error {
	if m.Provider != ProviderZhipu || m.API != APIOpenAICompletions {
		return nil
	}
	o, _ := options.(ChatOptions)
	if o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort) {
		return newError(CodeInvalidRequest, PhaseCapability, "reasoning effort is not supported by this model")
	}
	return nil
}

// validateChatFinalBody also runs without OnPayload: sampling can replace
// reasoning fields on that path too. Authorization precedes capability/model
// checks, so an invalid effort never masks an unauthorized model or reference.
func validateChatFinalBody(m Model, raw []byte, authorize func(map[string]any) string) *Error {
	if m.Provider != ProviderZhipu {
		return nil
	}
	var body map[string]any
	if json.Unmarshal(raw, &body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if problem := validateGLMChatThinking(m, body); problem != "" {
		return newError(CodeInvalidRequest, PhaseRequest, problem)
	}
	if problem := validateGLMChatBudget(m, body); problem != "" {
		return newError(CodeInvalidRequest, PhaseRequest, problem)
	}
	return nil
}

func validateGLMChatBudget(m Model, body map[string]any) string {
	if _, present := body["max_completion_tokens"]; present {
		return "Zhipu Chat requires max_tokens instead of max_completion_tokens"
	}
	if raw, present := body["max_tokens"]; present {
		// raw was decoded from the final JSON, so only float64 is numeric.
		n, ok := raw.(float64)
		if !ok || math.Trunc(n) != n || n < 1 || n > float64(min(m.MaxTokens, glmChatMaxOutputTokens)) {
			return "max_tokens must be an integer from 1 through the model output limit"
		}
	}
	return ""
}

func validateGLMChatThinking(m Model, body map[string]any) string {
	if raw, present := body["thinking"]; present {
		thinking, ok := raw.(map[string]any)
		if !ok || (thinking["type"] != "enabled" && thinking["type"] != "disabled") {
			return "thinking must declare type enabled or disabled"
		}
		if thinking["type"] == "disabled" && (m.Compat.RequiresThinking || m.ThinkingLevelMap.Off.IsNull()) {
			return "this model requires thinking and cannot use thinking.disabled"
		}
		if thinking["type"] == "enabled" && !m.Reasoning {
			return "this model does not support thinking"
		}
	}
	if raw, present := body["reasoning_effort"]; present {
		effort, ok := raw.(string)
		if !ok || !slices.Contains([]string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}, effort) {
			return "reasoning_effort must be a supported GLM effort string"
		}
		if m.Compat.RequiresThinking && (effort == "none" || effort == "minimal") {
			return "this model requires thinking and cannot use a disabling effort"
		}
		// A documented alias can be supplied by a hook or sampling even when
		// named canonical options normally normalize it through the map.
		for _, level := range supportedThinkingLevels(m) {
			if string(level) == effort || m.ThinkingLevelMap.wireValue(level) == effort {
				return ""
			}
		}
		return "reasoning_effort is not supported by this model"
	}
	return ""
}
