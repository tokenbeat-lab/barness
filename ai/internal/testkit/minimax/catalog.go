// Package minimax assembles controlled Host candidates. It makes no formal
// support or price claim; each protocol must earn its own live acceptance.
package minimax

import (
	_ "embed"
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

//go:embed candidate.json
var candidate []byte

// CandidateCatalog returns six independently declared protocol/model entries.
// All rates are synthetic test data, never subscription or standard API fees.
func CandidateCatalog() ai.Catalog {
	var cat ai.Catalog
	if err := json.Unmarshal(candidate, &cat); err != nil {
		panic("invalid embedded MiniMax candidate catalog")
	}
	return cat
}

// ChatCatalog returns the independent Chat entries from the candidate snapshot.
func ChatCatalog() ai.Catalog { return protocolCatalog(ai.APIOpenAICompletions) }

// ResponsesCatalog returns the independent Responses candidate entries.
func ResponsesCatalog() ai.Catalog { return protocolCatalog(ai.APIOpenAIResponses) }

// MessagesCatalog returns the independent Messages candidate entries.
func MessagesCatalog() ai.Catalog { return protocolCatalog(ai.APIAnthropicMessages) }

func protocolCatalog(api ai.API) ai.Catalog {
	cat := CandidateCatalog()
	cat.Models = cat.ModelsOf(ai.ProviderMiniMax, api)
	return cat
}

// Documented international SDK bases; account purpose never changes the host.
const (
	ChatEndpoint      = "https://api.minimax.io/v1"
	ResponsesEndpoint = "https://api.minimax.io/v1"
	MessagesEndpoint  = "https://api.minimax.io/anthropic"
)

// ChatBinding explicitly fixes one Host-authorized Chat candidate account.
func ChatBinding(tenant, account, credentialRef, endpoint string) ai.Binding {
	return binding(tenant, account, credentialRef, endpoint, "minimax-chat", ai.APIOpenAICompletions)
}

// ResponsesBinding fixes a Responses account independently of Chat/Messages.
func ResponsesBinding(tenant, account, credentialRef, endpoint string) ai.Binding {
	return binding(tenant, account, credentialRef, endpoint, "minimax-responses", ai.APIOpenAIResponses)
}

// MessagesBinding fixes a Messages account independently of Chat/Responses.
func MessagesBinding(tenant, account, credentialRef, endpoint string) ai.Binding {
	return binding(tenant, account, credentialRef, endpoint, "minimax-messages", ai.APIAnthropicMessages)
}

func binding(tenant, account, credentialRef, endpoint, id string, api ai.API) ai.Binding {
	return ai.Binding{TenantID: tenant, BindingID: id, Version: "minimax-candidate-1",
		Operation: ai.OperationChat, Enabled: true, ProviderID: ai.ProviderMiniMax,
		API: api, Endpoint: endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: account,
		CredentialRef: credentialRef, AllowedModels: []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"}}
}
