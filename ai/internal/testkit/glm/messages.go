package glm

import (
	_ "embed"
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

// Messages candidates come from their own Messages wire exploration; canonical
// effort mappings remain absent until that protocol has evidence for them.
//
//go:embed messages-catalog.json
var messagesCatalog []byte

func MessagesCatalog() ai.Catalog {
	var catalog ai.Catalog
	if err := json.Unmarshal(messagesCatalog, &catalog); err != nil {
		panic("invalid embedded GLM Messages candidate catalog")
	}
	return catalog
}

// MessagesEndpoint is the China standard SDK base. The locked SDK appends
// /v1/messages?beta=true; both models accepted it in the bounded wire probe.
const MessagesEndpoint = "https://open.bigmodel.cn/api/anthropic"

func MessagesBinding(tenant, account, credentialRef, endpoint string) ai.Binding {
	return ai.Binding{TenantID: tenant, BindingID: "glm-messages", Version: "glm-messages-candidate-1",
		Operation: ai.OperationChat, Enabled: true, ProviderID: ai.ProviderZhipu,
		API: ai.APIAnthropicMessages, Endpoint: endpoint, AuthKind: ai.AuthAPIKey,
		AccountScopeID: account, CredentialRef: credentialRef,
		AllowedModels: []string{"glm-5.3", "glm-5.2"}}
}
