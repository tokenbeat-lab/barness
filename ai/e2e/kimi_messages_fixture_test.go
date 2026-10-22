package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"testing"
)

// Synthetic protocol scenarios deliberately use controlled credentials and a
// candidate catalog. They do not confer live or formal support qualification.
func newKimiMessagesWorld(t *testing.T, market ai.ProviderID, mutate func(*ai.Catalog)) *world {
	t.Helper()
	return newKimiMessagesConfigured(t, market, mutate, nil)
}
func newKimiMessagesConfigured(t *testing.T, market ai.ProviderID, mutate func(*ai.Catalog), configure func(*ai.Config)) *world {
	t.Helper()
	cat := kimi.ProtocolCatalog(market, ai.APIAnthropicMessages)
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		if configure != nil {
			configure(c)
		}
	}, tenantA)
	b := ai.Binding{TenantID: tenantA.tenant, BindingID: "kimi-messages", Version: "v1", Enabled: true, ProviderID: market, API: ai.APIAnthropicMessages, Endpoint: w.provider.URL() + "/anthropic", AuthKind: ai.AuthAPIKey, AccountScopeID: "acct-kimi-" + string(market), CredentialRef: "cred-kimi-" + string(market), AllowedModels: []string{"kimi-k3"}}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(anthropicA.secret)})
	return w
}

var kimiMessagesTarget = ai.Target{BindingID: "kimi-messages", ModelID: "kimi-k3"}

func kimiMessagesRequest() ai.Request {
	return ai.Request{SystemPrompt: "Synthetic system.", Messages: []ai.Message{ai.UserText("Synthetic input.")}}
}
func kimiMessagesScript(usage string) string {
	return `event: message_start
data: {"type":"message_start","message":{"id":"msg-kimi-synthetic","model":"kimi-k3","usage":` + usage + `}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":"Synthetic reply."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":50}}

event: message_stop
data: {"type":"message_stop"}

`
}
func kimiMessagesReply() provider.Reply {
	return provider.SSE([][]byte{[]byte(kimiMessagesScript(`{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}`))})
}
