//go:build live

package live

import (
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// combo is one first-phase Provider × API combination and how the smoke
// drives it. Defaults pick an inexpensive catalog model that has every
// capability the combination promises (spec §5: choose a supporting model
// rather than skip).
type combo struct {
	name string
	// spec is the spec's protocol scenario, P01–P06, prefixing case IDs.
	spec     string
	provider ai.ProviderID
	api      ai.API
	endpoint string
	sdk      func() string
	model    string
	// quiet is the simple reasoning level of the scenarios that are not
	// about reasoning: the model's lowest, so the output budget goes to the
	// answer rather than to thinking.
	quiet ai.ThinkingLevel
	// full is the protocol's full options with output budget maxTokens,
	// reasoning at quiet and tool use as choice says.
	full func(maxTokens int, choice toolChoice) ai.Options
	// forcesTool: the tool round trip forces the call (spec P01/P04 "强制
	// 工具往返"; elsewhere it makes the smoke deterministic). Without it the
	// round trip relies on the prompt, with automatic choice.
	forcesTool bool
	// reasoning is the level of the reasoning-history scenario; empty means
	// the protocol returns nothing to replay and noReplay says why.
	reasoning ai.ThinkingLevel
	noReplay  string
	// extra are the combination's own scenarios.
	extra []scenario
}

func openAISDK() string {
	return "github.com/openai/openai-go/v3 " + evidence.ModuleVersion("github.com/openai/openai-go/v3")
}

func anthropicSDK() string {
	return "github.com/anthropics/anthropic-sdk-go " + evidence.ModuleVersion("github.com/anthropics/anthropic-sdk-go") +
		" (HTTP; SSE decoded by barness-ai, ADR-0011)"
}

func directHTTP() string { return "direct HTTP (ADR-0012)" }

func responsesOptions(effort ai.ThinkingLevel) func(int, toolChoice) ai.Options {
	return func(maxTokens int, choice toolChoice) ai.Options {
		o := ai.ResponsesOptions{MaxTokens: ai.Value(maxTokens), ReasoningEffort: effort}
		switch choice {
		case toolForced:
			o.ToolChoice = ai.Value(ai.ResponsesToolChoice{Function: weatherTool.Name})
		case toolsOff:
			o.ToolChoice = ai.Value(ai.ResponsesToolChoice{Mode: "none"})
		}
		return o
	}
}

func chatOptions(effort ai.ThinkingLevel) func(int, toolChoice) ai.Options {
	return func(maxTokens int, choice toolChoice) ai.Options {
		o := ai.ChatOptions{MaxTokens: ai.Value(maxTokens), ReasoningEffort: effort}
		switch choice {
		case toolForced:
			o.ToolChoice = ai.Value(ai.ChatToolChoice{Function: weatherTool.Name})
		case toolsOff:
			o.ToolChoice = ai.Value(ai.ChatToolChoice{Mode: "none"})
		}
		return o
	}
}

var combos = []combo{
	{
		name: "openai-responses", spec: "P01", provider: ai.ProviderOpenAI, api: ai.APIOpenAIResponses,
		endpoint: "https://api.openai.com/v1", sdk: openAISDK, model: "gpt-5-mini",
		quiet: ai.ThinkingMinimal, full: responsesOptions(ai.ThinkingMinimal), forcesTool: true,
		reasoning: ai.ThinkingLow,
	},
	{
		name: "anthropic-messages", spec: "P02", provider: ai.ProviderAnthropic, api: ai.APIAnthropicMessages,
		endpoint: "https://api.anthropic.com", sdk: anthropicSDK, model: "claude-haiku-4-5",
		full: func(maxTokens int, choice toolChoice) ai.Options {
			o := ai.AnthropicOptions{MaxTokens: ai.Value(maxTokens)}
			switch choice {
			case toolForced:
				o.ToolChoice = ai.Value(ai.AnthropicToolChoice{Tool: weatherTool.Name})
			case toolsOff:
				o.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "none"})
			}
			return o
		},
		forcesTool: true, reasoning: ai.ThinkingLow,
	},
	{
		name: "google-gemini", spec: "P03", provider: ai.ProviderGoogle, api: ai.APIGoogleGenerativeAI,
		endpoint: "https://generativelanguage.googleapis.com/v1beta", sdk: directHTTP, model: "gemini-3.8-flash",
		full: func(maxTokens int, choice toolChoice) ai.Options {
			o := ai.GeminiOptions{MaxTokens: ai.Value(maxTokens), Thinking: ai.Value(ai.GeminiThinking{Enabled: false})}
			switch choice {
			case toolForced:
				o.ToolChoice = ai.Value(ai.GeminiToolChoiceAny)
			case toolsOff:
				o.ToolChoice = ai.Value(ai.GeminiToolChoiceNone)
			}
			return o
		},
		forcesTool: true, reasoning: ai.ThinkingLow,
	},
	{
		name: "openai-chat", spec: "P04", provider: ai.ProviderOpenAI, api: ai.APIOpenAICompletions,
		endpoint: "https://api.openai.com/v1", sdk: openAISDK, model: "gpt-5-mini",
		quiet: ai.ThinkingMinimal, full: chatOptions(ai.ThinkingMinimal), forcesTool: true,
		noReplay: "OpenAI Chat Completions returns no reasoning content or signature to replay; reasoning is only counted in usage",
		extra:    []scenario{chatUsagePosition},
	},
	{
		// DeepSeek's Responses: no level sends effort "none" (its map has
		// no off entry). The forced choice is not part of P05's smoke, so
		// the round trip uses automatic choice.
		name: "deepseek-responses", spec: "P05", provider: ai.ProviderDeepSeek, api: ai.APIOpenAIResponses,
		endpoint: "https://api.deepseek.com", sdk: openAISDK, model: "deepseek-flash",
		full: responsesOptions(""), reasoning: ai.ThinkingLow,
		extra: []scenario{reasoningLevel("deepseek-flash", ai.ThinkingHigh), reasoningLevel("deepseek-flash", ai.ThinkingMax),
			authRefused},
	},
	{
		// DeepSeek's Chat: no level switches thinking off, so the forced
		// tool runs with thinking off (spec P06); thinking with tools,
		// forced and automatic, runs in its own scenarios.
		name: "deepseek-chat", spec: "P06", provider: ai.ProviderDeepSeek, api: ai.APIOpenAICompletions,
		endpoint: "https://api.deepseek.com", sdk: openAISDK, model: "deepseek-flash",
		full: chatOptions(""), forcesTool: true, reasoning: ai.ThinkingLow,
		extra: []scenario{chatUsagePosition, forcedToolWithThinking, promptCacheLongRetention,
			reasoningLevel("deepseek-v4-pro", ai.ThinkingHigh), reasoningLevel("deepseek-v4-pro", ai.ThinkingMax),
			authRefused},
	},
}
