package e2e

import (
	"context"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/glmhost"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
)

func TestRegionalProviderHostCredentialOwnership(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderZhipu, ai.ProviderMiniMax} {
		for _, failure := range []string{"provider", "tenant", "account", "purpose", "reference", "inactive", "api", "version"} {
			t.Run(string(service)+"/"+failure, func(t *testing.T) {
				ev := run.Case(t, "regional-host-credential-"+string(service)+"-"+failure)
				secretReads := 0
				read := func(context.Context, string) (ai.Secret, error) {
					secretReads++
					return ai.NewSecret(tenantA.secret), nil
				}
				var credentials ai.CredentialResolver
				var bindings []ai.Binding
				if service == ai.ProviderZhipu {
					a := glmhost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "account", CredentialRef: "record", Version: "v1", Purpose: glmhost.Standard}
					var err error
					bindings, err = glmhost.Bindings(a)
					if err != nil {
						t.Fatal(err)
					}
					record := glmhost.CredentialRecord{ProviderID: service, TenantID: a.TenantID, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, Purpose: glmhost.Standard, Version: "key-v1", BindingVersion: a.Version, Active: true, AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}}
					switch failure {
					case "provider":
						record.ProviderID = "zai"
					case "tenant":
						record.TenantID = tenantB.tenant
					case "account":
						record.AccountScopeID = "other"
					case "purpose":
						record.Purpose = "coding-plan"
					case "reference":
						record.CredentialRef = "other"
					case "inactive":
						record.Active = false
					case "api":
						record.AllowedAPIs = []ai.API{ai.APIAnthropicMessages}
					case "version":
						record.BindingVersion = "new-version"
					}
					credentials = glmhost.Credentials{Authorization: a, Record: record, ReadSecret: read}
				} else {
					a := minimaxhost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "account", CredentialRef: "record", Version: "v1", Purpose: minimaxhost.Standard}
					var err error
					bindings, err = minimaxhost.Bindings(a)
					if err != nil {
						t.Fatal(err)
					}
					record := minimaxhost.CredentialRecord{ProviderID: service, TenantID: a.TenantID, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, Purpose: minimaxhost.Standard, Version: "key-v1", BindingVersion: a.Version, Active: true, AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}}
					switch failure {
					case "provider":
						record.ProviderID = "minimax-cn"
					case "tenant":
						record.TenantID = tenantB.tenant
					case "account":
						record.AccountScopeID = "other"
					case "purpose":
						record.Purpose = minimaxhost.MPlan
					case "reference":
						record.CredentialRef = "other"
					case "inactive":
						record.Active = false
					case "api":
						record.AllowedAPIs = []ai.API{ai.APIAnthropicMessages}
					case "version":
						record.BindingVersion = "new-version"
					}
					credentials = minimaxhost.Credentials{Authorization: a, Record: record, ReadSecret: read}
				}
				w := newWorldWith(t, func(c *ai.Config) { c.Credentials = credentials }, tenantA)
				b := bindings[0]
				b.Endpoint = w.provider.URL()
				w.host.PutBinding(b)
				result, err := w.client.Complete(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: "ownership"}, ai.Target{BindingID: b.BindingID, ModelID: b.AllowedModels[0]}, ai.Request{}, ai.ChatOptions{MaxTokens: ai.Value(64)})
				ev.Record("result", result)
				ev.Record("requests", w.provider.Requests())
				ev.Record("secret-read-count", secretReads)
				code, phase := ai.CodeTenantDenied, ai.PhaseCredential
				if failure == "version" {
					code, phase = ai.CodeCredentialUnavailable, ai.PhaseConsistency
				}
				failureErr, ok := err.(*ai.Error)
				ev.Check("metadata mismatch fails before secret or network", ok && failureErr.Code == code && failureErr.Phase == phase && secretReads == 0 && len(w.provider.Requests()) == 0 && len(result.Metadata.Attempts) == 0, "err=%v reads=%d", err, secretReads)
			})
		}
	}
}
