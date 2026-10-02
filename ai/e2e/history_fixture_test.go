package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// historyFixture is testdata/responses/history.json.
type historyFixture struct {
	Events    []json.RawMessage `json:"events"`
	Scenarios []historyScenario `json:"scenarios"`
}

type historyScenario struct {
	ID    string `json:"id"`
	About string `json:"about"`
	Model string `json:"model"`
	// SimpleMaxOutputTokens is pi's simple-entry max_output_tokens.
	SimpleMaxOutputTokens int `json:"simpleMaxOutputTokens"`
	// ModelCompat replaces the catalog model's compat flags on both sides.
	ModelCompat *ai.ModelCompat `json:"modelCompat"`
	Context     struct {
		SystemPrompt string            `json:"systemPrompt"`
		Tools        []ai.Tool         `json:"tools"`
		Messages     []json.RawMessage `json:"messages"`
	} `json:"context"`
	Expect struct {
		Body       json.RawMessage          `json:"body"`
		Downgrades ai.NativeStateDowngrades `json:"downgrades"`
	} `json:"expect"`
}

func loadHistoryFixture(t *testing.T) (historyFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", "history.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f historyFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("history.json: %v", err)
	}
	return f, raw
}

func (f historyFixture) reply(t *testing.T) provider.Reply {
	return sseEvents(t, f.Events, provider.FramingLF)
}

// request restores the scenario's history the way the test host restores
// stored records for tenant k: messages are decoded into barness-ai's types,
// and a message marked trusted gets the envelope the host vouches for after
// its own storage checks (tenant k, k's account at the message's provider,
// the message's own provider, API and model).
func (sc historyScenario) request(t *testing.T, k tenantKey) ai.Request {
	t.Helper()
	req := ai.Request{SystemPrompt: sc.Context.SystemPrompt, Tools: sc.Context.Tools}
	for _, raw := range sc.Context.Messages {
		req.Messages = append(req.Messages, decodeStoredMessage(t, raw, k))
	}
	return req
}

// configure gives the scenario's model its compat flags.
func (sc historyScenario) configure(cfg *ai.Config) {
	withModelPatch(sc.Model, sc.ModelCompat, nil)(cfg)
}

func decodeStoredMessage(t *testing.T, raw json.RawMessage, k tenantKey) ai.Message {
	t.Helper()
	var head struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
		Trusted bool            `json:"trusted"`
	}
	mustUnmarshal(t, raw, &head)
	var blocks []json.RawMessage
	if head.Role != "system" && head.Content != nil {
		mustUnmarshal(t, head.Content, &blocks)
	}
	switch head.Role {
	case "user":
		var m ai.UserMessage
		for _, b := range blocks {
			m.Content = append(m.Content, decodeStoredBlock(t, b).(ai.UserContent))
		}
		return m
	case "toolResult":
		var m ai.ToolResultMessage
		mustUnmarshal(t, raw, &struct {
			ToolCallID *string `json:"toolCallId"`
			ToolName   *string `json:"toolName"`
			IsError    *bool   `json:"isError"`
		}{&m.ToolCallID, &m.ToolName, &m.IsError})
		for _, b := range blocks {
			m.Content = append(m.Content, decodeStoredBlock(t, b).(ai.ToolResultContent))
		}
		return m
	case "system":
		var m ai.SystemMessage
		mustUnmarshal(t, raw, &m)
		return m
	case "assistant":
		var s storedMessage
		mustUnmarshal(t, raw, &s)
		var errorMessage struct {
			ErrorMessage string `json:"errorMessage"`
		}
		mustUnmarshal(t, raw, &errorMessage)
		m := ai.AssistantMessage{API: s.API, Provider: s.Provider, Model: s.Model, Usage: s.Usage,
			StopReason: s.StopReason, Timestamp: s.Timestamp, ErrorMessage: errorMessage.ErrorMessage}
		for _, b := range blocks {
			m.Content = append(m.Content, decodeStoredBlock(t, b).(ai.AssistantContent))
		}
		if head.Trusted {
			// The host vouches with the account of the binding that served
			// the message: the tenant's OpenAI, Anthropic or Google account.
			account := primaryCredential(k, "v1").AccountScopeID
			switch m.Provider {
			case ai.ProviderAnthropic:
				account = anthropicCredential(k).AccountScopeID
			case ai.ProviderGoogle:
				account = geminiCredential(k).AccountScopeID
			}
			env := ai.NativeStateEnvelope{TenantID: k.tenant, AccountScopeID: account,
				ProviderID: m.Provider, API: m.API, ModelID: m.Model}
			trusted, err := ai.TrustNativeState(ai.CallScope{TenantID: k.tenant, RequestID: "req-history-restore"}, env)
			if err != nil {
				t.Fatal(err)
			}
			m.NativeState = trusted
		}
		return m
	}
	t.Fatalf("stored message has unknown role %q", head.Role)
	return nil
}

// decodeStoredBlock decodes a content block by its type tag.
func decodeStoredBlock(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	var kind struct {
		Type string `json:"type"`
	}
	mustUnmarshal(t, raw, &kind)
	switch kind.Type {
	case "text":
		var b ai.Text
		mustUnmarshal(t, raw, &b)
		return b
	case "image":
		var b ai.Image
		mustUnmarshal(t, raw, &b)
		return b
	case "thinking":
		var b ai.Thinking
		mustUnmarshal(t, raw, &b)
		return b
	case "toolCall":
		var b ai.ToolCall
		mustUnmarshal(t, raw, &b)
		return b
	}
	t.Fatalf("stored block has unknown type %q", kind.Type)
	return nil
}

func mustUnmarshal(t *testing.T, raw []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
}

// pidiff is the scenario of TestHistoryNormalization for the differential.
func (f historyFixture) pidiff(t *testing.T, raw []byte, sc historyScenario) pidiffScenario {
	t.Helper()
	return pidiffScenario{fixture: "history.json", raw: raw, model: sc.Model, req: sc.request(t, tenantA),
		reply: f.reply(t), requests: 1, modelCompat: sc.ModelCompat}
}
