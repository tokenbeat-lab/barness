package e2e

import (
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/kimihost"
)

func TestKimiHostConfiguration(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(service), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("KIMI-%s-K02-host-official-targets", service))
			a := kimihost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "account", CredentialRef: "record", Version: "v1", Purpose: kimihost.Standard}
			bindings, err := kimihost.Bindings(a)
			ev.Record("bindings", bindings)
			if !ev.Check("explicit authorization creates three independent routes", err == nil && len(bindings) == 3, "err=%v routes=%v", err, bindings) {
				return
			}
			host, prefix := "https://api.moonshot.ai", "kimi-intl"
			other := "https://api.moonshot.cn/v1"
			if service == ai.ProviderMoonshotCN {
				host, prefix = "https://api.moonshot.cn", "kimi-cn"
				other = "https://api.moonshot.ai/v1"
			}
			for i, b := range bindings {
				path, name := "/v1", "chat"
				if i == 1 {
					name = "responses"
				}
				if i == 2 {
					path, name = "/anthropic", "messages"
				}
				ev.Check("each market fixes complete SDK base and exact allowlist", b.ProviderID == service && b.BindingID == prefix+"-"+name && b.Endpoint == host+path && b.AuthKind == ai.AuthAPIKey && len(b.AllowedModels) == 1 && b.AllowedModels[0] == "kimi-k3", "binding=%+v", b)
			}
			for i, bad := range []string{other, host + "/wrong", host + "/v1?key=synthetic", host + "/v1#fragment", "https://user:synthetic@api.moonshot.ai/v1", "http://remote.invalid/v1"} {
				a.TrustedGateway = true
				a.Endpoints = map[ai.API]string{ai.APIOpenAICompletions: bad}
				_, err := kimihost.Bindings(a)
				ev.Check("official target mismatch and credential-bearing URLs reject even with gateway trust", err != nil, "target-case=%d err=%v", i, err)
			}
			a.Endpoints = nil
			a.Purpose = "subscription"
			_, err = kimihost.Bindings(a)
			ev.Check("subscription purpose cannot substitute standard API authorization", err != nil, "err=%v", err)
		})
	}
}
