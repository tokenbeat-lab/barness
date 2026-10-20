package glmhost

import (
	"errors"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/internal/hostconfig"
)

// Purpose is an explicit account entitlement, never inferred from a URL/key.
type Purpose string

const Standard Purpose = "standard"

// Authorization is trusted deployment configuration. This standard API example
// does not grant Coding Plan eligibility or discover account entitlements.
type Authorization struct {
	ProviderID                                       ai.ProviderID
	TenantID, AccountScopeID, CredentialRef, Version string
	Purpose                                          Purpose
	Endpoints                                        map[ai.API]string
	TrustedGateway                                   bool
}

var ErrAuthorization = errors.New("glmhost: explicit service, tenant, account, credential, version and standard purpose required")

// Bindings builds three independently revocable routes. Market data and purpose
// policy intentionally remain separate from the MiniMax example.
func Bindings(a Authorization) ([]ai.Binding, error) {
	if (a.ProviderID != ai.ProviderZhipu && a.ProviderID != ai.ProviderID("zai")) || a.Purpose != Standard {
		return nil, ErrAuthorization
	}
	for _, v := range []string{a.TenantID, a.AccountScopeID, a.CredentialRef, a.Version} {
		if strings.TrimSpace(v) == "" {
			return nil, ErrAuthorization
		}
	}
	host := "https://open.bigmodel.cn"
	prefix := "glm"
	if a.ProviderID == ai.ProviderID("zai") {
		host = "https://api.z.ai"
		prefix = "zai"
	}
	routes := []struct {
		name string
		api  ai.API
		path string
	}{
		{"chat", ai.APIOpenAICompletions, "/api/paas/v4"},
		{"responses", ai.APIOpenAIResponses, "/api/v1"},
		{"messages", ai.APIAnthropicMessages, "/api/anthropic"},
	}
	if err := hostconfig.ValidateAPIs(a.Endpoints); err != nil {
		return nil, err
	}
	out := make([]ai.Binding, 0, len(routes))
	for _, r := range routes {
		endpoint, err := hostconfig.Endpoint(host+r.path, a.Endpoints[r.api], a.TrustedGateway)
		if err != nil {
			return nil, err
		}
		out = append(out, ai.Binding{TenantID: a.TenantID, BindingID: prefix + "-" + r.name, Operation: ai.OperationChat, Version: a.Version, Enabled: true, ProviderID: a.ProviderID, API: r.api, Endpoint: endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, AllowedModels: []string{"glm-5.3", "glm-5.2"}})
	}
	return out, nil
}
