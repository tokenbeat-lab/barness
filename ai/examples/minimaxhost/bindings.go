// Package minimaxhost assembles explicit MiniMax market Host bindings. It
// validates trusted account/credential purpose before selected secret access.
package minimaxhost

import (
	"errors"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/internal/hostconfig"
)

// Purpose is the credential use the Host authorized for this deployment.
type Purpose string

const (
	Standard Purpose = "standard"
	MPlan    Purpose = "m-plan"
)

// Authorization comes from trusted Host configuration after checking account,
// deployment and model entitlement. It is not a client request or proof that
// MiniMax grants access. Secrets remain in a separate CredentialResolver.
type Authorization struct {
	Endpoints                                        map[ai.API]string
	TrustedGateway                                   bool
	TenantID, AccountScopeID, CredentialRef, Version string
	ProviderID                                       ai.ProviderID
	Purpose                                          Purpose
}

var ErrAuthorization = errors.New("minimaxhost: explicit service, tenant, account, credential, version and purpose are required")

// Bindings keeps the three protocols and two purposes independently revocable.
// Standard in either market grants only M3; the international M Plan candidate
// grants only M3.1. The Host supplies its reviewed Catalog to ai.NewClient;
// constructing these bindings does not publish support or validate pricing.
//
// This intentionally stays separate from the GLM example: allowed models,
// service, credential purposes and entitlement evolve under different contracts.
func Bindings(a Authorization) ([]ai.Binding, error) {
	if a.ProviderID != ai.ProviderMiniMax && a.ProviderID != ai.ProviderMiniMaxCN {
		return nil, ErrAuthorization
	}
	for _, value := range []string{a.TenantID, a.AccountScopeID, a.CredentialRef, a.Version} {
		if strings.TrimSpace(value) == "" {
			return nil, ErrAuthorization
		}
	}
	var model string
	switch a.Purpose {
	case Standard:
		model = "MiniMax-M3"
	case MPlan:
		if a.ProviderID != ai.ProviderMiniMax {
			return nil, ErrAuthorization
		}
		model = "MiniMax-M3.1-Flash-Preview"
	default:
		return nil, ErrAuthorization
	}
	routes := []struct {
		name     string
		api      ai.API
		endpoint string
	}{
		{"chat", ai.APIOpenAICompletions, "https://api.minimax.io/v1"},
		{"responses", ai.APIOpenAIResponses, "https://api.minimax.io/v1"},
		{"messages", ai.APIAnthropicMessages, "https://api.minimax.io/anthropic"},
	}
	if err := hostconfig.ValidateAPIs(a.Endpoints); err != nil {
		return nil, err
	}
	out := make([]ai.Binding, 0, len(routes))
	for _, r := range routes {
		if a.ProviderID == ai.ProviderMiniMaxCN {
			r.endpoint = strings.Replace(r.endpoint, "api.minimax.io", "api.minimax.cn", 1)
		}
		endpoint, err := hostconfig.Endpoint(r.endpoint, a.Endpoints[r.api], a.TrustedGateway)
		if err != nil {
			return nil, err
		}
		out = append(out, ai.Binding{TenantID: a.TenantID, BindingID: string(a.ProviderID) + "-" + string(a.Purpose) + "-" + r.name, Operation: ai.OperationChat, Version: a.Version, Enabled: true, ProviderID: a.ProviderID, API: r.api, Endpoint: endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, AllowedModels: []string{model}})
	}
	return out, nil
}
