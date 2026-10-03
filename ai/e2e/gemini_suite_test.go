package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// geminiSuite is the Gemini Developer API (P03) to the shared suites. The
// adapter decodes the stream itself (ADR-0012). pi's Google path has no
// timeoutMs of its own and never retries a request that got no response,
// never calls onResponse, and its payload callback sees the REST body.
var geminiSuite = &protocolSuite{
	proto: geminiProtocol, prefix: "P03", options: ai.GeminiOptions{},

	toolArgs: `{"q":"x","n":2}`, // as Gemini sent it
	limited:  "http-429",
	held: heldScripts{
		start:     "data: " + `{"candidates":[{"content":{"parts":[{"text":"Hel"}],"role":"model"},"index":0}]}` + "\n\n",
		frameOpen: "data: ",
		filler:    strings.Repeat(": keep-alive\n\n", 200),
		errorOpen: `{"error":{"code":500,"message":"`,
	},
	longOutput: geminiLongOutput,

	timeoutOptions:       nil,
	headerTimeoutRetried: false,
	overloaded: provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json"},
		Chunks: [][]byte{[]byte(`{"error":{"code":503,"message":"The model is overloaded.","status":"UNAVAILABLE"}}`)}},
	deniedModel: "gemini-3.1-flash-live-preview",
	credential:  geminiCredential,

	keyHeader: "X-Goog-Api-Key",
	keyValue:  func(k tenantKey) string { return k.secret },
	callbacks: callbackData{
		onResponse: false,
		replaced: replacedPayload{
			name: "replaced-payload",
			body: func(*ai.Payload) map[string]any {
				return map[string]any{
					"contents":         []any{map[string]any{"role": "user", "parts": []any{map[string]any{"text": "Replaced."}}}},
					"generationConfig": map[string]any{"maxOutputTokens": 10},
				}
			},
			sent:  `{"contents":[{"role":"user","parts":[{"text":"Replaced."}]}],"generationConfig":{"maxOutputTokens":10}}`,
			about: "the replacement is sent as given",
		},
		changedInPlace: &payloadChange{
			change: func(b map[string]any) {
				b["generationConfig"].(map[string]any)["topK"] = 3
				b["safetySettings"] = []any{map[string]any{"category": "HARM_CATEGORY_HARASSMENT", "threshold": "BLOCK_NONE"}}
			},
			sent: func(b map[string]any) bool {
				config, _ := b["generationConfig"].(map[string]any)
				return config["topK"] == 3.0 && b["safetySettings"] != nil
			},
			about: "the in-place changes are sent",
		},
		widen: []payloadWidening{
			{"model", func(b map[string]any) { b["model"] = "models/gemini-2.5-pro" }},
			{"cached-content", func(b map[string]any) { b["cachedContent"] = "cachedContents/abc123" }},
			{"mcp-servers", func(b map[string]any) {
				b["tools"] = []any{map[string]any{"mcpServers": []any{map[string]any{"name": "x", "streamableHttpTransport": map[string]any{"url": "https://mcp.example"}}}}}
			}},
			{"file-search", func(b map[string]any) {
				b["tools"] = []any{map[string]any{"fileSearch": map[string]any{"fileSearchStoreNames": []any{"fileSearchStores/s1"}}}}
			}},
			{"hosted-tool", func(b map[string]any) { b["tools"] = []any{map[string]any{"googleSearch": map[string]any{}}} }},
			{"file-data", func(b map[string]any) {
				b["contents"] = []any{map[string]any{"role": "user", "parts": []any{map[string]any{
					"fileData": map[string]any{"mimeType": "application/pdf", "fileUri": "https://generativelanguage.googleapis.com/v1beta/files/abc"}}}}}
			}},
		},
		hostedTool: hostedTool{
			name:  "googleSearch",
			add:   func(b map[string]any) { b["tools"] = []any{map[string]any{"googleSearch": map[string]any{}}} },
			sent:  func(b map[string]any) bool { tools, _ := b["tools"].([]any); return len(tools) == 1 },
			about: "the allowed hosted tool is sent",
		},
	},

	authFailure: provider.Reply{Status: 403, Header: map[string]string{"Content-Type": "application/json"},
		Chunks: [][]byte{[]byte(geminiAuthBody)}, Script: geminiAuthBody},
	redaction: redactionData{
		kept: []string{"c2lnLWNhbGw="},
		echo: func(secret string) (int, string) {
			return 403, fmt.Sprintf(`{"error":{"code":403,"message":"API key %s cannot answer %q","status":"PERMISSION_DENIED"}}`, secret, "Look up x.")
		},
		sensitive:  []string{"Look up x.", `{"q":"x","n":2}`, "c2lnLXRoaW5r", "c2lnLWNhbGw=", "Let me think", "Checking.", "cannot answer"},
		notInError: []string{`{"q":"x","n":2}`, "c2lnLWNhbGw="},
	},
	foreignOptions: ai.AnthropicOptions{},
	rejections: []rejectCase{
		{id: "invalid-options", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: fullEntries(
			ai.GeminiOptions{ToolChoice: ai.Value(ai.GeminiToolChoice("required"))},
			ai.GeminiOptions{Thinking: ai.Value(ai.GeminiThinking{Enabled: true, BudgetTokens: ai.Value(-2)})})},
		{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "gemini", ModelID: "gemini-3.1-flash-live-preview"}},
		{id: "model-of-another-api", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "gemini", ModelID: "claude-haiku-4-5"}},
	},
	env: environmentData{
		vars: map[string]string{"GOOGLE_GENAI_USE_VERTEXAI": "true", "GOOGLE_CLOUD_PROJECT": "env-leak"},
	},
}

const geminiAuthBody = `{"error":{"code":403,"message":"Permission denied.","status":"PERMISSION_DENIED"}}`

// geminiLongOutput is a generated Gemini stream of one text block in n
// parts of 9 bytes each, ended by a chunk holding only the finish reason,
// and the parts.
func geminiLongOutput(t *testing.T, n int) (provider.Reply, []string) {
	t.Helper()
	var deltas []string
	var raw [][]byte
	for i := range n {
		d := fmt.Sprintf("part-%03d ", i)
		deltas = append(deltas, d)
		chunk := mustMarshal(t, map[string]any{"candidates": []any{map[string]any{
			"content": map[string]any{"parts": []any{map[string]any{"text": d}}, "role": "model"}, "index": 0}}})
		raw = append(raw, []byte("data: "+string(chunk)+"\n\n"))
	}
	raw = append(raw, []byte("data: "+`{"candidates":[{"finishReason":"STOP","index":0}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":200,"totalTokenCount":210}}`+"\n\n"))
	return provider.SSE(raw), deltas
}
