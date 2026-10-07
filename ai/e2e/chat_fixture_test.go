package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// chatProtocol is OpenAI × Chat Completions (P04): scenarios in
// testdata/chat, the "chat" binding on tenant A's OpenAI account (shared
// with "primary"), and replies as unnamed SSE data events, the stream
// closed by data: [DONE].
var chatProtocol = &fixtureProtocol{
	dir: "chat", binding: "chat", api: ai.APIOpenAICompletions, provider: ai.ProviderOpenAI,
	key: func(k tenantKey) tenantKey { return k }, account: "acct-tenant-a",
	full: func(t *testing.T, raw json.RawMessage) ai.Options {
		var o ai.ChatOptions
		if raw != nil {
			mustUnmarshal(t, raw, &o)
		}
		return o
	},
	sse: chatEvents, textMarker: `"content":"Hello"`,
}

// chatFixtureFiles are the P04 fixtures under testdata/chat; the
// differential runs every scenario in them that does not opt out.
var chatFixtureFiles = []string{"text.json", "failures.json", "history.json", "options.json", "usage.json", "retry.json", "catalog.json", "parity-models.json", "parity-errors.json"}

// chatEvents lays scripted chunks out as an SSE reply of unnamed data
// events; the JSON string "[DONE]" is the bare end marker.
func chatEvents(t *testing.T, events []json.RawMessage, framing provider.Framing) provider.Reply {
	t.Helper()
	chunks, err := provider.EncodeChatSSE(events, framing)
	if err != nil {
		t.Fatal(err)
	}
	return provider.SSE(chunks)
}
