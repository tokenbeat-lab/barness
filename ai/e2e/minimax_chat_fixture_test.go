package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var minimaxChatKey = tenantKey{tenant: "tenant-a", secret: "synthetic-minimax-chat-a", alias: "key:minimax-chat-a@v1"}

func init() { knownKeys = append(knownKeys, minimaxChatKey) }

func newMiniMaxChatWorld(t *testing.T, mutate func(*ai.Catalog), configure ...func(*ai.Config)) *world {
	t.Helper()
	cat := minimax.ChatCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		for _, apply := range configure {
			apply(c)
		}
	}, tenantA)
	b := minimax.ChatBinding(tenantA.tenant, "acct-minimax-chat-a", "cred-minimax-chat-a", w.provider.URL()+"/v1")
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(minimaxChatKey.secret)})
	return w
}

func miniMaxChatTarget(model string) ai.Target {
	return ai.Target{BindingID: "minimax-chat", ModelID: model}
}
func miniMaxChatRequest() ai.Request {
	return ai.Request{SystemPrompt: "Be terse.", Messages: []ai.Message{ai.UserText("Say hello.")}}
}
func miniMaxChatFrames(model string) []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"id":"minimax-synthetic","model":"` + model + `","choices":[{"delta":{"reasoning_content":"Check."}}]}`),
		json.RawMessage(`{"choices":[{"delta":{"content":"Hello MiniMax."}}]}`),
		json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`),
		json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2},"completion_tokens_details":{"reasoning_tokens":3}}}`),
	}
}
func miniMaxChatReply(t *testing.T, model string) provider.Reply {
	return chatEvents(t, miniMaxChatFrames(model), "")
}
