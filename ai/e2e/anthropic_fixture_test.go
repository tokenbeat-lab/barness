package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// anthropicProtocol is Anthropic × Messages (P02): scenarios in
// testdata/anthropic, the "claude" binding on tenant A's Anthropic account,
// and replies as named SSE events.
var anthropicProtocol = &fixtureProtocol{
	dir: "anthropic", binding: "claude", api: ai.APIAnthropicMessages, provider: ai.ProviderAnthropic,
	key: anthropicKey, account: "acct-tenant-a-anthropic",
	full: func(t *testing.T, raw json.RawMessage) ai.Options {
		var o ai.AnthropicOptions
		if raw != nil {
			mustUnmarshal(t, raw, &o)
		}
		return o
	},
	sse: sseEvents, textMarker: `"text_delta"`,
}

// anthropicFixtureFiles are the P02 fixtures under testdata/anthropic; the
// differential runs every scenario in them that does not opt out.
var anthropicFixtureFiles = []string{"text.json", "failures.json", "history.json", "options.json", "usage.json", "retry.json", "catalog.json"}
