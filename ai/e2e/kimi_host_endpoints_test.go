package e2e

import (
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
)

func TestKimiHostOfficialHostnameCannotBeGateway(t *testing.T) {
	for i, endpoint := range []string{"https://api.moonshot.cn/v1", "https://api.moonshot.ai/v1", "https://API.MOONSHOT.AI/v1", "https://api.moonshot.cn./v1"} {
		t.Run(endpoint, func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("KIMI-K02-official-host-gateway-bypass-%d", i))
			_, err := minimaxhost.Bindings(minimaxhost.Authorization{ProviderID: ai.ProviderMiniMax, TenantID: "tenant", AccountScopeID: "account", CredentialRef: "record", Version: "v1", Purpose: minimaxhost.Standard, TrustedGateway: true, Endpoints: map[ai.API]string{ai.APIOpenAICompletions: endpoint}})
			ev.Check("known official Kimi target cannot borrow another service's gateway authorization", err != nil, "endpoint=%s err=%v", endpoint, err)
		})
	}
}
