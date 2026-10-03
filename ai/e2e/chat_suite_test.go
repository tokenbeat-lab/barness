package e2e

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// chatSuite is OpenAI Chat Completions (P04) to the shared suites. The
// adapter decodes the stream itself through barness's body wrapper
// (ADR-0013); it shares the Stainless SDKs' timeoutMs and retry rules — a
// header timeout is a connection failure without a response, as openai-node
// retries it — and runs onResponse.
var chatSuite = &protocolSuite{
	proto: chatProtocol, prefix: "P04", options: ai.ChatOptions{},

	toolArgs: `{"q":"x","n":2}`, // from its streamed fragments
	// Chat's failures have no JSON 429; its plain-text one is the limited
	// body.
	limited: "http-429-text",
	held: heldScripts{
		start:     "data: " + `{"id":"chatcmpl-001","object":"chat.completion.chunk","created":1760000000,"model":"gpt-4.1-mini-2025-04-14","usage":null,"choices":[{"index":0,"delta":{"content":"Hel"},"logprobs":null,"finish_reason":null}]}` + "\n\n",
		frameOpen: "data: ",
		filler:    strings.Repeat(": keep-alive\n\n", 200),
		errorOpen: `{"error":{"message":"`,
	},
	longOutput: chatLongOutput,

	timeoutOptions:       func(ms int) ai.Options { return ai.ChatOptions{TimeoutMs: ms} },
	headerTimeoutRetried: true,
	overloaded: provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"},
		Chunks: [][]byte{[]byte(`{"error":{"message":"Overloaded","type":"server_error"}}`)}},
	// gpt-5-pro is OpenAI's but served only by Responses.
	deniedModel: "gpt-5-pro",
	credential:  func(k tenantKey) ai.Credential { return primaryCredential(k, "v1") },

	keyHeader: "Authorization",
	keyValue:  func(k tenantKey) string { return "Bearer " + k.secret },
	callbacks: callbackData{
		onResponse: true,
		replaced: replacedPayload{
			name: "replaced-payload-still-streams",
			body: func(p *ai.Payload) map[string]any {
				return map[string]any{
					"model":                 p.ModelID,
					"messages":              []any{map[string]any{"role": "user", "content": "Replaced."}},
					"max_completion_tokens": 10,
					"stream":                false,
				}
			},
			sent:  `{"model":"gpt-4.1-mini","messages":[{"role":"user","content":"Replaced."}],"max_completion_tokens":10,"stream":true}`,
			about: "the replacement is sent with stream forced back on",
		},
		changedInPlace: &payloadChange{
			change: func(b map[string]any) {
				b["top_p"] = 0.5
				b["functions"] = []any{map[string]any{"name": "legacy", "parameters": map[string]any{"type": "object"}}}
			},
			sent:  func(b map[string]any) bool { return b["top_p"] == 0.5 && b["functions"] != nil && b["store"] == false },
			about: "the in-place changes are sent, legacy function tools included",
		},
		widen: []payloadWidening{
			{"model", func(b map[string]any) { b["model"] = "gpt-4o-mini" }},
			{"model-removed", func(b map[string]any) { delete(b, "model") }},
			{"cache-key", func(b map[string]any) { b["prompt_cache_key"] = "other-tenant-cache" }},
			{"stored", func(b map[string]any) { b["store"] = true }},
			{"hosted-tool", func(b map[string]any) {
				b["tools"] = []any{map[string]any{"type": "custom", "custom": map[string]any{"name": "grammar"}}}
			}},
			{"untyped-tool", func(b map[string]any) { b["tools"] = []any{map[string]any{"name": "x"}} }},
			{"web-search", func(b map[string]any) { b["web_search_options"] = map[string]any{} }},
			{"file-id", func(b map[string]any) {
				b["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "file",
					"file": map[string]any{"file_id": "file_123"}}}}}
			}},
			{"stored-audio", func(b map[string]any) {
				b["messages"] = append(b["messages"].([]any), map[string]any{"role": "assistant", "audio": map[string]any{"id": "audio_123"}})
			}},
		},
		hostedTool: hostedTool{
			name:  "web_search_options",
			add:   func(b map[string]any) { b["web_search_options"] = map[string]any{"search_context_size": "low"} },
			sent:  func(b map[string]any) bool { return b["web_search_options"] != nil },
			about: "the allowed hosted web search is sent",
		},
	},

	authFailure: provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "req_vendor_401"},
		Chunks: [][]byte{[]byte(chatAuthBody)}, Script: chatAuthBody},
	authFailureRequestID: "req_vendor_401",
	redaction: redactionData{
		kept: []string{"Let me think.", "call_1"},
		echo: func(secret string) (int, string) {
			return 401, fmt.Sprintf(`{"error":{"message":"Incorrect API key provided: %s. Received header Authorization: Bearer %s for input %q","type":"invalid_request_error","code":"invalid_api_key"}}`,
				secret, secret, "Look up x.")
		},
		sensitive:  []string{"Bearer", "Look up x.", `{"q":"x","n":2}`, `\"q\"`, "Let me think", "Still sure", "Checking.", "Incorrect API key"},
		notInError: []string{"Bearer sk-", `{"q":"x","n":2}`, "Let me think"},
	},
	foreignOptions: ai.AnthropicOptions{},
	rejections: []rejectCase{
		{id: "invalid-options", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: fullEntries(
			ai.ChatOptions{ReasoningEffort: ai.ThinkingLevel("extreme")},
			ai.ChatOptions{ToolChoice: ai.Value(ai.ChatToolChoice{Mode: "sometimes"})})},
		{id: "reserved-sampling-params", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: fullEntries(
			ai.ChatOptions{SamplingParams: map[string]json.RawMessage{"store": json.RawMessage(`true`)}},
			ai.ChatOptions{SamplingParams: map[string]json.RawMessage{"web_search_options": json.RawMessage(`{}`)}})},
		{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			arrange: func(w *world) {
				w.updateBindingOf(tenantA, "chat", func(b *ai.Binding) {
					b.AllowedModels = slices.DeleteFunc(slices.Clone(b.AllowedModels), func(id string) bool { return id == "gpt-4.1-mini" })
				})
			}},
		{id: "model-of-another-api", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "chat", ModelID: "gpt-5-pro"}},
		{id: "model-of-another-provider", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "chat", ModelID: "claude-haiku-4-5"}},
	},
	env: environmentData{
		vars:   map[string]string{"OPENAI_ORG_ID": "org-env-leak", "OPENAI_PROJECT_ID": "proj-env-leak"},
		leaked: []string{"OpenAI-Organization", "OpenAI-Project"},
	},
}

const chatAuthBody = `{"error":{"message":"Incorrect API key provided.","type":"invalid_request_error","param":null,"code":"invalid_api_key"}}`

// chatChunk is one streamed Chat Completions chunk of chatLongOutput with
// the given choice delta and finish reason.
func chatChunk(t *testing.T, delta map[string]any, finish any) json.RawMessage {
	t.Helper()
	return mustMarshal(t, map[string]any{"id": "chatcmpl-001", "object": "chat.completion.chunk", "created": 1760000000,
		"model": "gpt-4.1-mini-2025-04-14", "usage": nil,
		"choices": []any{map[string]any{"index": 0, "delta": delta, "logprobs": nil, "finish_reason": finish}}})
}

// chatLongOutput is a generated Chat stream of one text block in n content
// deltas of 9 bytes each, then the finish chunk, the usage chunk and
// [DONE], and the deltas.
func chatLongOutput(t *testing.T, n int) (provider.Reply, []string) {
	t.Helper()
	var deltas []string
	var events []json.RawMessage
	for i := range n {
		d := fmt.Sprintf("part-%03d ", i)
		deltas = append(deltas, d)
		events = append(events, chatChunk(t, map[string]any{"content": d}, nil))
	}
	events = append(events,
		chatChunk(t, map[string]any{}, "stop"),
		json.RawMessage(`{"id":"chatcmpl-001","object":"chat.completion.chunk","created":1760000000,"model":"gpt-4.1-mini-2025-04-14","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":200,"total_tokens":210}}`),
		json.RawMessage(provider.ChatDone),
	)
	return chatEvents(t, events, provider.FramingLF), deltas
}
