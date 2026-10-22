package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/kimihost"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
)

func TestKimiHostCredentialOwnership(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			for _, kind := range []string{"provider", "tenant", "account", "purpose", "reference", "inactive", "api", "version", "endpoint", "disabled", "model", "wrong-options"} {
				t.Run(fmt.Sprintf("%s/%s/%s", service, api, kind), func(t *testing.T) {
					ev := run.Case(t, fmt.Sprintf("KIMI-%s-%s-K02-host-ownership-%s", service, api, kind))
					a := kimihost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "account", CredentialRef: "record", Version: "v1", Purpose: kimihost.Standard}
					bindings, e := kimihost.Bindings(a)
					if e != nil {
						t.Fatal(e)
					}
					var b ai.Binding
					for _, v := range bindings {
						if v.API == api {
							b = v
						}
					}
					record := kimihost.CredentialRecord{ProviderID: service, TenantID: a.TenantID, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, Purpose: kimihost.Standard, Version: "secret-v1", BindingVersion: a.Version, Active: true, AllowedAPIs: []ai.API{api}}
					switch kind {
					case "provider":
						record.ProviderID = ai.ProviderMoonshot
						if service == ai.ProviderMoonshot {
							record.ProviderID = ai.ProviderMoonshotCN
						}
					case "tenant":
						record.TenantID = tenantB.tenant
					case "account":
						record.AccountScopeID = "other-account"
					case "purpose":
						record.Purpose = "subscription"
					case "reference":
						record.CredentialRef = "other-record"
					case "inactive":
						record.Active = false
					case "api":
						record.AllowedAPIs = nil
					case "version":
						record.BindingVersion = "v2"
					case "endpoint":
						b.Endpoint = "https://gateway.invalid/v1"
					case "disabled":
						b.Enabled = false
					}
					reads := 0
					credentials := kimihost.Credentials{Authorization: a, Record: record, ReadSecret: func(context.Context, string) (ai.Secret, error) { reads++; return ai.NewSecret(tenantA.secret), nil }}
					cat := kimi.CandidateCatalog()
					w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat; c.Credentials = credentials }, tenantA)
					w.host.PutBinding(b)
					target := ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}
					var opts ai.Options
					if kind == "model" {
						target.ModelID = "kimi-k2.5"
					}
					if kind == "wrong-options" {
						opts = ai.ChatOptions{}
						if api == ai.APIOpenAICompletions {
							opts = ai.ResponsesOptions{}
						}
					}
					result, err := w.client.Complete(ctxFor(t), textScope("ownership-"+kind), target, ai.Request{Messages: []ai.Message{ai.UserText("Check ownership.")}}, opts)
					ev.Record("result", result)
					ev.Record("secret-reads", reads)
					ev.Record("requests", w.provider.Requests())
					expected := ai.CodeTenantDenied
					if kind == "version" {
						expected = ai.CodeCredentialUnavailable
					}
					if kind == "wrong-options" {
						expected = ai.CodeInvalidRequest
					}
					failure, ok := err.(*ai.Error)
					ev.Check("ownership and capability refusals precede secret material and network", ok && failure.Code == expected && reads == 0 && len(result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "err=%v reads=%d", err, reads)
				})
			}
		}
	}
}
