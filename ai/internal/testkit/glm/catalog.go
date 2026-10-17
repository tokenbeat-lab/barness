// Package glm holds independently sourced, versioned Host candidates. It is
// test assembly, not a built-in catalog or a claim of formal service support.
package glm

import (
	_ "embed"
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

//go:embed chat-catalog.json
var chatCatalog []byte

// ChatCatalog returns a fresh Chat-only snapshot. Other protocols must bring
// their own evidence and entries; they are never generated from these models.
func ChatCatalog() ai.Catalog {
	var catalog ai.Catalog
	if err := json.Unmarshal(chatCatalog, &catalog); err != nil {
		panic("invalid embedded GLM Chat candidate catalog")
	}
	return catalog
}

// ChatEndpoint is the documented China standard API SDK base, without the
// resource suffix. Coding Plan keys/endpoints are outside this candidate.
const ChatEndpoint = "https://open.bigmodel.cn/api/paas/v4"

// ChatBinding explicitly fixes one tenant/account to the standard Chat route.
// endpoint may be a controlled local Provider base in offline acceptance.
func ChatBinding(tenant, account, credentialRef, endpoint string) ai.Binding {
	return ai.Binding{TenantID: tenant, BindingID: "glm-chat", Version: "glm-chat-candidate-1",
		Operation: ai.OperationChat, Enabled: true, ProviderID: ai.ProviderZhipu,
		API: ai.APIOpenAICompletions, Endpoint: endpoint, AuthKind: ai.AuthAPIKey,
		AccountScopeID: account, CredentialRef: credentialRef,
		AllowedModels: []string{"glm-5.3", "glm-5.2"}}
}
