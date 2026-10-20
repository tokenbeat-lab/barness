package e2e

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
)

func TestMiniMaxHostBindings(t *testing.T) {
	for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
		t.Run(string(purpose), func(t *testing.T) {
			ev := run.Case(t, "MM08-bindings-"+string(purpose))
			cfg := minimaxhost.Authorization{ProviderID: ai.ProviderMiniMax, TenantID: "example-tenant", AccountScopeID: "example-account-" + string(purpose), CredentialRef: "credential-" + string(purpose), Version: "b7", Purpose: purpose}
			got, err := minimaxhost.Bindings(cfg)
			ev.Record("bindings", got)
			if !ev.Check("three separately authorized protocol routes", err == nil && len(got) == 3, "%v", err) {
				return
			}
			model := "MiniMax-M3"
			if purpose == minimaxhost.MPlan {
				model = "MiniMax-M3.1-Flash-Preview"
			}
			endpoints := map[ai.API]string{ai.APIOpenAICompletions: "https://api.minimax.io/v1", ai.APIOpenAIResponses: "https://api.minimax.io/v1", ai.APIAnthropicMessages: "https://api.minimax.io/anthropic"}
			ids := map[string]bool{}
			for _, b := range got {
				ev.Check("official endpoint and explicit purpose-specific identity", b.Operation == ai.OperationChat && b.ProviderID == ai.ProviderID("minimax") && b.Endpoint == endpoints[b.API] && b.TenantID == cfg.TenantID && b.AccountScopeID == cfg.AccountScopeID && b.CredentialRef == cfg.CredentialRef && b.Version == cfg.Version && b.Enabled && b.AuthKind == ai.AuthAPIKey && reflect.DeepEqual(b.AllowedModels, []string{model}) && b.Retry.MaxRetries == 0, "%+v", b)
				ev.Check("independent BindingID", !ids[b.BindingID], "duplicate binding")
				ids[b.BindingID] = true
			}
			got[0].AllowedModels[0] = "mutation"
			ev.Check("allowlists have no shared mutable state", got[1].AllowedModels[0] == model && got[2].AllowedModels[0] == model, "%+v", got)
			other := cfg
			other.Purpose = minimaxhost.Standard
			if purpose == minimaxhost.Standard {
				other.Purpose = minimaxhost.MPlan
			}
			different, _ := minimaxhost.Bindings(other)
			for _, b := range different {
				ev.Check("purpose change cannot overwrite another binding", !ids[b.BindingID], "%s", b.BindingID)
			}
		})
	}
	for _, cfg := range []minimaxhost.Authorization{{}, {TenantID: "t", AccountScopeID: "a", CredentialRef: "c", Version: "v", Purpose: "guess-from-key"}} {
		ev := run.Case(t, "MM08-bindings-invalid-"+string(cfg.Purpose))
		got, err := minimaxhost.Bindings(cfg)
		ev.Check("missing identity or unknown purpose fails closed", errors.Is(err, minimaxhost.ErrAuthorization) && len(got) == 0, "%v", err)
	}
}
