package glm

import (
	_ "embed"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
)

//go:embed responses-catalog.json
var responsesCatalog []byte

// ResponsesCatalog is independently versioned from Chat. The Responses schema
// supplies its wire effort/output limits; model pages supply text and context.
func ResponsesCatalog() ai.Catalog {
	var c ai.Catalog
	if json.Unmarshal(responsesCatalog, &c) != nil {
		panic("invalid GLM Responses candidate catalog")
	}
	return c
}

// ResponsesEndpoint is the China standard Responses SDK base.
const ResponsesEndpoint = "https://open.bigmodel.cn/api/v1"

// ResponsesBinding fixes one tenant/account to the independent Responses route.
func ResponsesBinding(tenant, account, credentialRef, endpoint string) ai.Binding {
	return ai.Binding{TenantID: tenant, BindingID: "glm-responses", Version: "glm-responses-candidate-1", Operation: ai.OperationChat, Enabled: true, ProviderID: ai.ProviderZhipu, API: ai.APIOpenAIResponses, Endpoint: endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: account, CredentialRef: credentialRef, AllowedModels: []string{"glm-5.3", "glm-5.2"}}
}
