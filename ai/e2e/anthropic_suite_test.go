package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// anthropicSuite is Anthropic Messages (P02) to the shared suites. Its
// stream is decoded by barness itself (ADR-0011); it shares the Stainless
// SDKs' timeoutMs and retry rules and runs onResponse, and its beta
// features travel as the betas parameter a payload callback sees.
var anthropicSuite = &protocolSuite{
	proto: anthropicProtocol, prefix: "P02", options: ai.AnthropicOptions{},

	toolArgs: `{"q":"x"}`, // from two input_json_delta fragments
	limited:  "http-429",
	held: heldScripts{
		start:     "event: message_start\ndata: " + `{"type":"message_start","message":{"id":"msg_held","model":"claude-haiku-4-5","usage":{"input_tokens":5,"output_tokens":1}}}` + "\n\n",
		frameOpen: "event: content_block_delta\ndata: ",
		filler:    strings.Repeat("event: ping\ndata: {\"type\":\"ping\"}\n\n", 80),
		errorOpen: `{"type":"error","error":{"type":"api_error","message":"`,
	},
	longOutput: anthropicLongOutput,

	timeoutOptions:       func(ms int) ai.Options { return ai.AnthropicOptions{TimeoutMs: ms} },
	headerTimeoutRetried: true,
	overloaded: provider.Reply{Status: 529, Header: map[string]string{"Content-Type": "application/json"},
		Chunks: [][]byte{[]byte(`{"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`)}},
	deniedModel: "claude-legacy-x",
	credential:  anthropicCredential,

	keyHeader: "X-Api-Key",
	keyValue:  func(k tenantKey) string { return k.secret },
	callbacks: anthropicCallbacks,

	authFailure: provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json", "request-id": "req_vendor_401"},
		Chunks: [][]byte{[]byte(anthropicAuthBody)}, Script: anthropicAuthBody},
	authFailureRequestID: "req_vendor_401",
	redaction: redactionData{
		kept: []string{"sig-abc"},
		echo: func(secret string) (int, string) {
			return 401, fmt.Sprintf(`{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key: %s for %q"}}`, secret, "Look up x.")
		},
		sensitive:  []string{"Look up x.", `{"q":"x"}`, "sig-abc", "Let me think", "Hi there", "invalid x-api-key"},
		notInError: []string{`{"q":"x"}`, "sig-abc"},
	},
	foreignOptions: ai.ResponsesOptions{},
	rejections: []rejectCase{
		{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "claude", ModelID: "claude-legacy-x"}},
		{id: "model-of-another-api", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "claude", ModelID: "gpt-4.1-mini"}},
	},
	env: environmentData{
		vars:   map[string]string{"ANTHROPIC_PROFILE": "env-leak", "ANTHROPIC_CUSTOM_HEADERS": "X-Env-Leak: 1"},
		leaked: []string{"X-Env-Leak"},
	},
}

const anthropicAuthBody = `{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`

var anthropicCallbacks = callbackData{
	onResponse: true,
	replaced: replacedPayload{
		name: "replaced-payload-still-streams",
		body: func(p *ai.Payload) map[string]any {
			return map[string]any{
				"model":      p.ModelID,
				"messages":   []any{map[string]any{"role": "user", "content": "Replaced."}},
				"max_tokens": 10,
				"stream":     false,
				"betas":      []any{"custom-beta-1", "custom-beta-1", "b2"},
			}
		},
		sent:  `{"model":"claude-haiku-4-5","messages":[{"role":"user","content":"Replaced."}],"max_tokens":10,"stream":true}`,
		about: "the replacement is sent with stream forced back on and without betas",
		check: func(ev *evidence.Case, r provider.Request) {
			ev.Check("the betas parameter is sent as the anthropic-beta header as given, as pi's SDK does",
				r.Header["Anthropic-Beta"] == "custom-beta-1,custom-beta-1,b2", "got %q", r.Header["Anthropic-Beta"])
		},
	},
	widen: []payloadWidening{
		{"model", func(b map[string]any) { b["model"] = "claude-opus-4-5" }},
		{"mcp-servers", func(b map[string]any) {
			b["mcp_servers"] = []any{map[string]any{"type": "url", "url": "https://mcp.example"}}
		}},
		{"container", func(b map[string]any) { b["container"] = "container_123" }},
		{"hosted-tool", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "web_search_20250305", "name": "web_search"}}
		}},
		{"file-source", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "document",
				"source": map[string]any{"type": "file", "file_id": "file_123"}}}}}
		}},
	},
	hostedTool: hostedTool{
		name: "web_search_20250305",
		add: func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "web_search_20250305", "name": "web_search"}}
		},
		sent:  func(b map[string]any) bool { tools, _ := b["tools"].([]any); return len(tools) == 1 },
		about: "the allowed hosted tool is sent",
	},
	headerTransform: headerTransform{
		extra: func(h http.Header) {
			h.Set("Anthropic-Beta", "x-beta, x-beta,")
			// Even a host's transform cannot describe the shared host to
			// the vendor (issue 31).
			h.Set("X-Stainless-Arch", "host-chosen")
		},
		check: func(ev *evidence.Case, r provider.Request) {
			ev.Check("a configured anthropic-beta replaces the features, normalized as pi does",
				r.Header["Anthropic-Beta"] == "x-beta", "got %q", r.Header["Anthropic-Beta"])
			_, arch := r.Header["X-Stainless-Arch"]
			_, osSent := r.Header["X-Stainless-Os"]
			ev.Check("no host description sent, not even one the transform set", !arch && !osSent,
				"arch %q, os sent %t", r.Header["X-Stainless-Arch"], osSent)
		},
	},
	more: testAnthropicPayloadBetas,
}

// testAnthropicPayloadBetas: the payload callback sees the beta features as
// pi's betas parameter, and removing them in place removes the header.
func testAnthropicPayloadBetas(t *testing.T, s *protocolSuite) {
	options, oraw := loadFixture(t, s.proto, "options.json")
	budget := scenarioByID(t, options, "simple-budget-reasoning")
	t.Run("payload-sees-betas", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, s.caseID("E04-callbacks-payload-sees-betas"), budget, oraw)
		var betas any
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			betas = p.Body["betas"]
			delete(p.Body, "betas")
			p.Body["top_k"] = 3
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).CompleteSimple(ctxFor(t), textScope("req-cb-betas"), budget.target(), budget.request(t), budget.simpleOptions(t))
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("the callback sees the beta features as pi's betas parameter",
			reflect.DeepEqual(betas, []any{"interleaved-thinking-2025-05-14"}), "got %#v", betas)
		r := lastRequest(ev, w)
		_, beta := r.Header["Anthropic-Beta"]
		ev.Check("removing them in place removes the header", !beta, "got %q", r.Header["Anthropic-Beta"])
		var body map[string]any
		mustUnmarshal(t, r.Body, &body)
		_, hasBetas := body["betas"]
		ev.Check("an in-place change is sent; betas never travel in the body", body["top_k"] == 3.0 && !hasBetas, "got %s", r.Body)
	})
}

// anthropicLongOutput is a generated Anthropic stream of one text block in
// n deltas of 9 bytes each, and the deltas.
func anthropicLongOutput(t *testing.T, n int) (provider.Reply, []string) {
	t.Helper()
	var deltas []string
	events := []json.RawMessage{
		json.RawMessage(`{"type":"message_start","message":{"id":"msg_long","type":"message","role":"assistant","model":"claude-haiku-4-5","content":[],"usage":{"input_tokens":10,"output_tokens":1,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`),
		json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`),
	}
	for i := range n {
		d := fmt.Sprintf("part-%03d ", i)
		deltas = append(deltas, d)
		events = append(events, mustMarshal(t, map[string]any{"type": "content_block_delta", "index": 0,
			"delta": map[string]any{"type": "text_delta", "text": d}}))
	}
	events = append(events,
		json.RawMessage(`{"type":"content_block_stop","index":0}`),
		json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":200}}`),
		json.RawMessage(`{"type":"message_stop"}`),
	)
	return sseEvents(t, events, provider.FramingLF), deltas
}
