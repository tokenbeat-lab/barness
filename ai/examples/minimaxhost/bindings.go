// Package minimaxhost assembles explicit MiniMax international Host bindings. It
// reads no credentials and does not discover entitlements or select fallbacks.
package minimaxhost

import (
	"errors"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
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
	TenantID, AccountScopeID, CredentialRef, Version string
	Purpose                                          Purpose
}

var ErrAuthorization = errors.New("minimaxhost: explicit tenant, account, credential, version and purpose are required")

// Bindings keeps the three protocols and two purposes independently revocable.
// Standard here grants only M3; the explicitly qualified M Plan candidate
// grants only M3.1. The Host supplies its reviewed Catalog to ai.NewClient;
// constructing these bindings does not publish support or validate pricing.
//
// This intentionally stays separate from the GLM example: allowed models,
// region, credential purposes and entitlement evolve under different contracts.
func Bindings(a Authorization) ([]ai.Binding, error) {
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
	out := make([]ai.Binding, 0, len(routes))
	for _, r := range routes {
		out = append(out, ai.Binding{TenantID: a.TenantID, BindingID: "minimax-" + string(a.Purpose) + "-" + r.name, Operation: ai.OperationChat, Version: a.Version, Enabled: true, ProviderID: ai.ProviderMiniMax, API: r.api, Endpoint: r.endpoint, AuthKind: ai.AuthAPIKey, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, AllowedModels: []string{model}})
	}
	return out, nil
}
