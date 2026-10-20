package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

// An official storage promotion does not establish a cache-write token rate.
func TestRegionalProviderCatalogUnknownWrites(t *testing.T) {
	for _, service := range []ai.ProviderID{"zai", "minimax-cn"} {
		t.Run(string(service), func(t *testing.T) {
			ev := run.Case(t, "regional-catalog-unknown-writes-"+string(service))
			cat, err := regional.CandidateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
			model := "glm-5.3"
			if service == "minimax-cn" {
				model = "MiniMax-M3"
			}
			b := ai.Binding{TenantID: tenantA.tenant, BindingID: "regional-pricing", Version: "v1", Operation: ai.OperationChat, Enabled: true, ProviderID: service, API: ai.APIAnthropicMessages, Endpoint: w.provider.URL(), AuthKind: ai.AuthAPIKey, AccountScopeID: "account", CredentialRef: "record", AllowedModels: []string{model}}
			w.host.PutBinding(b)
			w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: "record", Version: "key-v1", AccountScopeID: "account", Active: true, APIKey: ai.NewSecret(tenantA.secret)})
			frames := []json.RawMessage{
				json.RawMessage(`{"type":"message_start","message":{"id":"own-source","type":"message","role":"assistant","model":"` + model + `","content":[],"usage":{"input_tokens":20,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":3}}}`),
				json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}`),
				json.RawMessage(`{"type":"message_stop"}`),
			}
			ev.Fixture("script.json", mustMarshal(t, frames))
			w.provider.Enqueue(sseEvents(t, frames, provider.FramingLF))
			result, err := w.client.Complete(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: "unknown-writes"}, ai.Target{BindingID: b.BindingID, ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic price input.")}}, ai.AnthropicOptions{MaxTokens: ai.Value(64)})
			ev.Record("result", result)
			ev.Record("requests", w.provider.Requests())
			ev.Record("sources", regional.Candidates())
			failure, ok := err.(*ai.Error)
			ev.Check("positive writes retain usage but cannot certify zero price", ok && failure.Code == ai.CodeProtocol && failure.Phase == ai.PhaseStream && result.Message.Usage.CacheWrite == 3 && len(result.Metadata.Attempts) == 1, "err=%v usage=%+v", err, result.Message.Usage)
		})
	}
}
