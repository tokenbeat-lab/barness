package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// deepseekChatProtocol is DeepSeek × Chat Completions (P06): scenarios in
// testdata/deepseek-chat, the "deepseek-chat" binding on tenant A's
// DeepSeek account and credential (shared with the Responses binding
// "deepseek"), and replies as Chat SSE data events. It shares the Chat
// adapter with OpenAI under DeepSeek's compat (ADR-0015).
var deepseekChatProtocol = &fixtureProtocol{
	dir: "deepseek-chat", binding: "deepseek-chat", api: ai.APIOpenAICompletions, provider: ai.ProviderDeepSeek,
	key: deepseekKey, account: "acct-tenant-a-deepseek",
	full: func(t *testing.T, raw json.RawMessage) ai.Options {
		var o ai.ChatOptions
		if raw != nil {
			mustUnmarshal(t, raw, &o)
		}
		return o
	},
	sse: chatEvents, textMarker: `"content":"Hello`,
}

// deepseekChatFixtureFiles are the P06 fixtures; frozen pi routes DeepSeek
// through openai-completions, so the differential runs every scenario in
// them that does not opt out.
var deepseekChatFixtureFiles = []string{"text.json", "failures.json", "history.json", "options.json", "usage.json", "retry.json"}

// TestDeepSeekChatText is P06/E01 through the public Client: text with
// thinking off, reasoning_content, thinking with the automatic tool choice
// and a forced tool choice with thinking off — the two tool cases are
// separate scenarios, so neither proves the other (spec P06). The text
// scenarios also run through Result without Next and through Complete.
func TestDeepSeekChatText(t *testing.T) {
	runProtocolScenarios(t, deepseekChatProtocol, "P06", "text.json", "E01",
		func(sc fixtureScenario) bool { return sc.ID == "text" || sc.ID == "text-simple" })
}

// TestDeepSeekChatFailures is P06/E02: missing and error finish reasons,
// cut and canceled reasoning, and DeepSeek's HTTP refusals, including its
// refusal of a forced tool choice in thinking mode.
func TestDeepSeekChatFailures(t *testing.T) {
	runProtocolScenarios(t, deepseekChatProtocol, "P06", "failures.json", "E02",
		func(sc fixtureScenario) bool { return sc.CancelAfterEvents == 0 })
}

// TestDeepSeekChatHistory is P06/E03: reasoning_content replay beside tool
// calls and results, the empty reasoning_content DeepSeek requires, other
// providers' and APIs' turns, images and placeholders.
func TestDeepSeekChatHistory(t *testing.T) {
	runProtocolScenarios(t, deepseekChatProtocol, "P06", "history.json", "E03", nil)
}

// TestDeepSeekChatOptions is P06/E04: max_tokens, the thinking switch and
// reasoning_effort through the level map, tool choice, cache fields and
// samplingParams, on the full and the simple entry.
func TestDeepSeekChatOptions(t *testing.T) {
	runProtocolScenarios(t, deepseekChatProtocol, "P06", "options.json", "E04", nil)
}

// TestDeepSeekChatUsage is P06/E11: usage in or after the finish chunk is
// never missed before [DONE]; cache hits and reasoning are priced at
// DeepSeek's rates.
func TestDeepSeekChatUsage(t *testing.T) {
	runProtocolScenarios(t, deepseekChatProtocol, "P06", "usage.json", "E11", nil)
}

// TestDeepSeekChatRetry is P06/E05 under the binding's retry policy.
func TestDeepSeekChatRetry(t *testing.T) {
	runProtocolScenarios(t, deepseekChatProtocol, "P06", "retry.json", "E05", nil)
}
