package ai

import (
	"bytes"
	"cmp"
	"encoding/json"
	"github.com/buger/jsonparser"
	"maps"
	"slices"
)

// ThinkingSwitchOnly is independent from mandatory thinking and the Host's
// narrowed canonical set. A single allowed depth still requires a wire effort.
func miniMaxChatHasEffort(m Model) bool {
	return m.Reasoning && !m.Compat.ThinkingSwitchOnly
}

func validateMiniMaxChatOptions(m Model, options Options) *Error {
	o, _ := options.(ChatOptions)
	if c, present := o.ToolChoice.Get(); present && (c.Mode != "auto" || c.Function != "" || c.AllowedTools != nil) {
		return newError(CodeInvalidRequest, PhaseCapability, "MiniMax Chat supports only the default or auto tool choice")
	}
	if o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort) {
		return newError(CodeInvalidRequest, PhaseCapability, "reasoning effort is not supported by this model")
	}
	return nil
}

// Restore original JSON numbers after the pi-compatible JavaScript sampling
// merge. Final guards see exact wire values; OnPayload can still repair them.
func preserveMiniMaxChatSampling(body []byte, m Model, o ChatOptions) ([]byte, error) {
	level := clampThinkingLevel(m, cmp.Or(o.ReasoningEffort, m.DefaultThinkingLevel, thinkingOff))
	for _, source := range []map[string]json.RawMessage{m.SamplingParams, m.SamplingParamsByThinkingLevel[level], o.SamplingParams} {
		for _, field := range slices.Sorted(maps.Keys(source)) {
			var err error
			body, err = jsonparser.Set(body, source[field], field)
			if err != nil {
				return nil, err
			}
		}
	}
	return body, nil
}

func validateMiniMaxChatFinalBody(m Model, raw []byte, authorize func(map[string]any) string) *Error {
	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if dec.Decode(&body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	for _, check := range []func() string{
		func() string { return validateMiniMaxChatCapabilities(body) },
		func() string { return validateMiniMaxChatThinking(m, body) },
		func() string { return validateMiniMaxChatBudget(m, body) },
	} {
		if problem := check(); problem != "" {
			return newError(CodeInvalidRequest, PhaseRequest, problem)
		}
	}
	return nil
}

func validateMiniMaxChatBudget(m Model, body map[string]any) string {
	if raw, present := body["max_completion_tokens"]; present {
		n, ok := raw.(json.Number)
		if !ok || !outputBudgetNumber(json.RawMessage(n), min(m.MaxTokens, 131072)) {
			return "max_completion_tokens must be an integer from 1 through the model output limit"
		}
	}
	return ""
}

func validateMiniMaxChatThinking(m Model, body map[string]any) string {
	if body["reasoning_split"] != true {
		return "MiniMax Chat requires the integrated reasoning_split:true format"
	}
	disabled := false
	if raw, present := body["thinking"]; present {
		thinking, ok := raw.(map[string]any)
		if !ok || len(thinking) != 1 || (thinking["type"] != "adaptive" && thinking["type"] != "disabled") {
			return "MiniMax thinking must declare only type adaptive or disabled"
		}
		disabled = thinking["type"] == "disabled"
		if disabled && (m.Compat.RequiresThinking || m.ThinkingLevelMap.Off.IsNull()) {
			return "this model requires thinking"
		}
		if !disabled && !m.Reasoning {
			return "this model does not support thinking"
		}
	}
	if raw, present := body["reasoning_effort"]; present {
		effort, ok := raw.(string)
		if !ok || disabled || !miniMaxChatHasEffort(m) {
			return "MiniMax reasoning_effort requires supported adaptive depth control"
		}
		for _, level := range supportedThinkingLevels(m) {
			if level != thinkingOff && m.ThinkingLevelMap.wireValue(level) == effort && slices.Contains([]string{"low", "medium", "high", "xhigh", "max"}, effort) {
				return ""
			}
		}
		return "reasoning_effort is not supported by this model"
	}
	return ""
}

// Fail explicitly if the provider returns an unintegrated split format. A
// successful turn must retain every native entry needed by tool continuation.
func validateMiniMaxChatReasoningDelta(d chatDelta) string {
	if rawTruthy(d.Reasoning) || rawTruthy(d.ReasoningText) {
		return "MiniMax returned an unintegrated reasoning field"
	}
	if len(d.ReasoningContent) > 0 && string(d.ReasoningContent) != "null" {
		if _, ok := rawString(d.ReasoningContent); !ok {
			return "MiniMax reasoning_content is not text"
		}
	}
	if len(d.ReasoningDetails) == 0 || string(d.ReasoningDetails) == "null" {
		return ""
	}
	if rawTruthy(d.ReasoningContent) {
		return "MiniMax returned conflicting reasoning formats"
	}
	var entries []json.RawMessage
	if json.Unmarshal(d.ReasoningDetails, &entries) != nil || entries == nil {
		return "MiniMax returned an unintegrated reasoning_details format"
	}
	for _, raw := range entries {
		entry, err := parseJSON(string(raw))
		if err != nil {
			return "MiniMax returned invalid reasoning_details"
		}
		if _, ok := chatReasoningDetail(entry); !ok {
			return "MiniMax returned an unintegrated reasoning entry"
		}
	}
	return ""
}
