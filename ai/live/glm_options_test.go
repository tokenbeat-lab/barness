//go:build live

package live

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
)

func glmFullOptions(sc glmCasePlan) ai.Options {
	kind := sc.Kind
	if strings.HasPrefix(kind, "control-") {
		kind = "full-stream"
		if sc.Kind == "control-auto" {
			kind = "tool-call"
		}
		if sc.Kind == "control-unsigned" {
			kind = "native-history"
		}
	}
	switch glmSelected.api {
	case ai.APIOpenAICompletions:
		effort := ai.ThinkingLevel("")
		if kind == "full-stream" {
			effort = ai.ThinkingHigh
		}
		if kind == "full-complete" {
			effort = ai.ThinkingLow
		}
		o := ai.ChatOptions{MaxTokens: ai.Value(4096), ReasoningEffort: effort}
		if sc.Kind == "control-temperature" {
			o.Temperature = ai.Value(0.0)
		}
		if sc.Kind == "control-usage" {
			o.SamplingParams = map[string]json.RawMessage{"stream_options": json.RawMessage(`{"include_usage":true}`)}
		}
		return o
	case ai.APIOpenAIResponses:
		return ai.ResponsesOptions{MaxTokens: ai.Value(4096), ReasoningEffort: ai.ThinkingHigh}
	default:
		o := ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
		if sc.Model == "glm-5.2" && (kind == "full-stream" || kind == "native-history") && sc.Kind != "control-default" {
			o.ThinkingEnabled = ai.Value(true)
			o.ThinkingBudgetTokens = 1024
		}
		if sc.Kind == "control-auto" {
			o.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "auto"})
		}
		return o
	}
}
