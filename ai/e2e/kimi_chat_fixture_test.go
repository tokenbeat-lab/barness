package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"testing"
)

var kimiChatKey = tenantKey{tenant: "tenant-a", secret: "synthetic-kimi-chat-key", alias: "key:kimi-chat@v1"}

func init() { knownKeys = append(knownKeys, kimiChatKey) }

// Each market owns a literal candidate. Synthetic protocol evidence does not
// qualify a model for the formal catalog or real account access.
func kimiChatCatalog(t *testing.T) ai.Catalog {
	t.Helper()
	var cat ai.Catalog
	mustUnmarshal(t, []byte(`{"version":"kimi-chat-synthetic-2026-10-22.1","models":[
 {"provider":"moonshot","api":"openai-completions","id":"kimi-k3","name":"Kimi K3 international candidate","reasoning":true,"input":["text"],"contextWindow":1048576,"maxTokens":131072,"defaultThinkingLevel":"max","thinkingLevelMap":{"off":null,"minimal":null,"low":"low","medium":null,"high":"high","xhigh":null,"max":"max"},"compat":{"requiresThinking":true,"supportsLongCacheRetention":true},"cost":{"input":3,"output":15,"cacheRead":0.3,"cacheWrite":3}},
 {"provider":"moonshot-cn","api":"openai-completions","id":"kimi-k3","name":"Kimi K3 China candidate","reasoning":true,"input":["text"],"contextWindow":1048576,"maxTokens":131072,"defaultThinkingLevel":"max","thinkingLevelMap":{"off":null,"minimal":null,"low":"low","medium":null,"high":"high","xhigh":null,"max":"max"},"compat":{"requiresThinking":true,"supportsLongCacheRetention":true},"cost":{"input":2.9753049688,"output":14.8765248438,"cacheRead":0.2975304969,"cacheWrite":2.9753049688}}
 ]}`), &cat)
	return cat
}
func newKimiChatWorld(t *testing.T, market ai.ProviderID, mutate func(*ai.Catalog), configure ...func(*ai.Config)) *world {
	t.Helper()
	cat := kimiChatCatalog(t)
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		for _, f := range configure {
			f(c)
		}
	}, tenantA)
	b := ai.Binding{TenantID: tenantA.tenant, BindingID: "kimi-chat", Version: "v1", Enabled: true, Operation: ai.OperationChat, ProviderID: market, API: ai.APIOpenAICompletions, Endpoint: w.provider.URL() + "/v1", AuthKind: ai.AuthAPIKey, AccountScopeID: "acct-kimi-chat", CredentialRef: "cred-kimi-chat", AllowedModels: []string{"kimi-k3"}}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(kimiChatKey.secret)})
	return w
}
func kimiChatTarget() ai.Target { return ai.Target{BindingID: "kimi-chat", ModelID: "kimi-k3"} }
func kimiChatRequest() ai.Request {
	return ai.Request{SystemPrompt: "Be terse.", Messages: []ai.Message{ai.UserText("Say hello.")}}
}
func kimiChatFrames() []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"id":"kimi-chat-synthetic","model":"kimi-k3","choices":[{"delta":{"reasoning_content":"Check."}}]}`),
		json.RawMessage(`{"choices":[{"delta":{"content":"Hello Kimi."}}]}`),
		json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`),
		json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":200,"completion_tokens":50,"prompt_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"completion_tokens_details":{"reasoning_tokens":10}}}`),
	}
}
func kimiChatReply(t *testing.T) provider.Reply { return chatEvents(t, kimiChatFrames(), "") }
func kimiChatID(m ai.ProviderID, group string) string {
	if m == "moonshot-cn" {
		return "KIMI-CN-CHAT-" + group
	}
	return "KIMI-INTL-CHAT-" + group
}
