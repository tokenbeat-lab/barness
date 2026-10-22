package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// geminiProtocol is Google × Gemini Developer API (P03): scenarios in
// testdata/gemini, the "gemini" binding on tenant A's Google project, and
// replies as unnamed SSE data events, as Gemini streams generateContent.
var geminiProtocol = &fixtureProtocol{
	dir: "gemini", binding: "gemini", api: ai.APIGoogleGenerativeAI, provider: ai.ProviderGoogle,
	key: googleKey, account: "acct-tenant-a-google",
	full: func(t *testing.T, raw json.RawMessage) ai.Options {
		var o ai.GeminiOptions
		if raw != nil {
			mustUnmarshal(t, raw, &o)
		}
		return o
	},
	sse: dataEvents, textMarker: `"text":`,
}

// geminiFixtureFiles are the P03 fixtures under testdata/gemini; the
// public E2E runs each scenario against its independent expected behavior.
var geminiFixtureFiles = []string{"text.json", "failures.json", "history.json", "options.json", "usage.json", "retry.json", "thinking-sampling.json"}

// dataEvents lays scripted chunks out as an SSE reply of unnamed data events.
func dataEvents(t *testing.T, events []json.RawMessage, framing provider.Framing) provider.Reply {
	t.Helper()
	chunks, err := provider.EncodeDataSSE(events, framing)
	if err != nil {
		t.Fatal(err)
	}
	return provider.SSE(chunks)
}
