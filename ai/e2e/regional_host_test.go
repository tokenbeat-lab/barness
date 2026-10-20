package e2e

import (
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/glmhost"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
)

func TestRegionalProviderHostExplicitService(t *testing.T) {
	ev := run.Case(t, "regional-host-explicit-service")
	bindings, err := minimaxhost.Bindings(minimaxhost.Authorization{TenantID: "tenant", AccountScopeID: "account", CredentialRef: "record", Version: "v1", Purpose: minimaxhost.Standard})
	ev.Record("bindings", bindings)
	ev.Check("service must be explicitly authorized before assembly", errors.Is(err, minimaxhost.ErrAuthorization) && len(bindings) == 0, "err=%v count=%d", err, len(bindings))
}

func TestRegionalProviderHostEndpoints(t *testing.T) {
	for _, service := range []ai.ProviderID{"zhipu", "zai", "minimax", "minimax-cn"} {
		t.Run(string(service), func(t *testing.T) {
			ev := run.Case(t, "regional-host-endpoints-"+string(service))
			assemble := func(endpoint string, trusted bool) ([]ai.Binding, error) {
				if service == "zhipu" || service == "zai" {
					return glmhost.Bindings(glmhost.Authorization{ProviderID: service, TenantID: "tenant", AccountScopeID: "account", CredentialRef: string(service) + "-record", Version: "v1", Purpose: glmhost.Standard, Endpoints: map[ai.API]string{ai.APIOpenAICompletions: endpoint}, TrustedGateway: trusted})
				}
				return minimaxhost.Bindings(minimaxhost.Authorization{ProviderID: service, TenantID: "tenant", AccountScopeID: "account", CredentialRef: string(service) + "-record", Version: "v1", Purpose: minimaxhost.Standard, Endpoints: map[ai.API]string{ai.APIOpenAICompletions: endpoint}, TrustedGateway: trusted})
			}
			got, err := assemble("", false)
			if !ev.Check("explicit service owns three routes", err == nil && len(got) == 3, "%v", err) {
				return
			}
			ev.Record("official-bindings", got)
			want := map[ai.ProviderID][]string{
				"zhipu":      {"https://open.bigmodel.cn/api/paas/v4", "https://open.bigmodel.cn/api/v1", "https://open.bigmodel.cn/api/anthropic"},
				"zai":        {"https://api.z.ai/api/paas/v4", "https://api.z.ai/api/v1", "https://api.z.ai/api/anthropic"},
				"minimax":    {"https://api.minimax.io/v1", "https://api.minimax.io/v1", "https://api.minimax.io/anthropic"},
				"minimax-cn": {"https://api.minimax.cn/v1", "https://api.minimax.cn/v1", "https://api.minimax.cn/anthropic"},
			}
			for i, b := range got {
				ev.Check("actual provider and official protocol base", b.ProviderID == service && b.Endpoint == want[service][i], "%+v", b)
			}
			trailing, err := assemble(want[service][0]+"/", false)
			ev.Check("trailing slash canonicalized", err == nil && len(trailing) == 3 && trailing[0].Endpoint == want[service][0], "err=%v", err)
			invalid := []string{"https://api.openai.com/v1", "https://api.z.ai/api/v1", "https://api.minimax.cn/anthropic", "https://synthetic:secret@api.minimax.cn/v1", want[service][0] + "?api_key=synthetic", want[service][0] + "#fragment", want[service][0] + "#", want[service][0] + "/chat/completions", "http://gateway.example/v1", "https://api.minimax.cn:444/v1"}
			for _, endpoint := range invalid {
				bindings, err := assemble(endpoint, false)
				ev.Check("wrong service path or URL extras refused before calls", err != nil && len(bindings) == 0, "accepted endpoint")
			}
			for _, endpoint := range []string{"https://gateway.example/reviewed-base", "http://127.0.0.1:9001/controlled"} {
				bindings, err := assemble(endpoint, true)
				ev.Check("trusted gateway is an explicit Host choice", err == nil && len(bindings) == 3 && bindings[0].ProviderID == service && bindings[0].Endpoint == endpoint, "err=%v", err)
			}
			// An explicit gateway flag is for nonofficial hosts. It cannot turn
			// another official market/protocol into this service's route.
			for _, endpoint := range []string{"https://api.minimax.io/wrong", "https://api.minimax.cn/wrong", "https://api.z.ai/wrong", "https://open.bigmodel.cn/wrong", "https://API.Z.AI/wrong", "https://api.minimax.cn./wrong"} {
				bindings, err := assemble(endpoint, true)
				ev.Check("gateway flag cannot bypass official market/protocol check", err != nil && len(bindings) == 0, "accepted wrong official endpoint")
			}
			for _, endpoint := range []string{"http://remote.example/base", "https://user:password@gateway.example/base", "https://gateway.example/base?key=synthetic", "https://gateway.example/base#fragment"} {
				bindings, err := assemble(endpoint, true)
				ev.Check("gateway retains URL safety validation", err != nil && len(bindings) == 0, "accepted unsafe gateway")
			}
		})
	}
}
