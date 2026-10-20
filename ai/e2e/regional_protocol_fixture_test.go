package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

var regionalAPIs = []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}

// These are deliberately synthetic shared-shape tests. Reusing local fixtures
// checks adapter selection; it does not copy a regional catalog or certify live.
func regionalWorld(t *testing.T, p ai.ProviderID, api ai.API, configure ...func(*ai.Config)) (*world, ai.Target) {
	t.Helper()
	var cat ai.Catalog
	if p == "zai" || p == ai.ProviderZhipu {
		switch api {
		case ai.APIOpenAICompletions:
			cat = glm.ChatCatalog()
		case ai.APIOpenAIResponses:
			cat = glm.ResponsesCatalog()
		default:
			cat = glm.MessagesCatalog()
		}
	} else {
		cat = minimax.CandidateCatalog()
		cat.Models = cat.ModelsOf(ai.ProviderMiniMax, api)
	}
	for i := range cat.Models {
		cat.Models[i].Provider = p
	}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		for _, apply := range configure {
			apply(c)
		}
	}, tenantA)
	base := "/v1"
	if api == ai.APIAnthropicMessages {
		base = "/anthropic"
	}
	b := ai.Binding{TenantID: tenantA.tenant, BindingID: "regional", Version: "regional-synthetic-1", Enabled: true,
		ProviderID: p, API: api, Endpoint: w.provider.URL() + base, AuthKind: ai.AuthAPIKey,
		AccountScopeID: "shared-alias", CredentialRef: "regional-credential", AllowedModels: regionalModels(p)}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1",
		BindingVersion: b.Version, AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantA.secret)})
	return w, ai.Target{BindingID: b.BindingID, ModelID: cat.Models[0].ID}
}

func regionalModels(p ai.ProviderID) []string {
	if p == ai.ProviderZAI || p == ai.ProviderZhipu {
		return []string{"glm-5.3", "glm-5.2"}
	}
	return []string{"MiniMax-M3"}
}
