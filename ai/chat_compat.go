package ai

// chatCompat is the part of pi's openai-completions compat that differs
// between the vendors barness-ai serves over Chat Completions: what pi's
// detectCompat derives from the provider (and base URL) before a model's own
// compat overrides it. The flags pi carries on models that barness-ai
// implements (supportsStrictMode, supportsMidConvoSystemMessages,
// supportsLongCacheRetention) stay on ModelCompat.
//
// Like responsesCapabilities (ADR-0014) the profile is keyed by the
// binding's provider: it describes the vendor's endpoint, the same for each
// of its models, and pi's own DeepSeek model data repeats exactly the
// detected values. A host catalog therefore cannot give a DeepSeek model
// OpenAI's request shape or the reverse (ADR-0015).
type chatCompat struct {
	// store: the request declares store:false.
	store bool
	// developerRole: a reasoning model's instructions go out as
	// "developer" instead of "system".
	developerRole bool
	// maxTokens: the output budget is max_tokens instead of
	// max_completion_tokens.
	maxTokens bool
	// deepseekThinking: a reasoning model's thinking is switched by
	// thinking:{type:enabled|disabled}, with reasoning_effort only beside an
	// enabled one (pi's thinkingFormat "deepseek"); otherwise
	// reasoning_effort alone, the "off" mapping included (pi's "openai").
	deepseekThinking bool
	// reasoningContentOnAssistant: every replayed assistant message of a
	// reasoning model carries reasoning_content, empty when the turn
	// streamed none (pi's requiresReasoningContentOnAssistantMessages).
	reasoningContentOnAssistant bool
}

// chatCompatOf is p's profile, as pi 0.87.1 detects it: DeepSeek is a
// non-standard vendor (no store, no developer role, max_tokens) with its
// own thinking format that requires reasoning_content on assistant turns;
// every other provider gets the standard OpenAI compat (ADR-0013).
func chatCompatOf(p ProviderID) chatCompat {
	if p == ProviderDeepSeek {
		return chatCompat{maxTokens: true, deepseekThinking: true, reasoningContentOnAssistant: true}
	}
	return chatCompat{store: true, developerRole: true}
}
