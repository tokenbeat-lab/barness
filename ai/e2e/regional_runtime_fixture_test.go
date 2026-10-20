package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Public Client scenarios use synthetic identities and prices, never live support.
// The same account/session aliases deliberately span services and tenants; only
// BindingID/CredentialRef are unique per service within a tenant.
var regionalRuntimeProviders = []ai.ProviderID{ai.ProviderZhipu, ai.ProviderZAI, ai.ProviderMiniMax, ai.ProviderMiniMaxCN}

func regionalRuntimeKey(tenant string, p ai.ProviderID) tenantKey {
	return tenantKey{tenant: tenant, secret: "synthetic-regional-" + tenant + "-" + string(p), alias: "key:regional-" + tenant + "-" + string(p) + "@v1"}
}
func init() {
	for _, tenant := range []string{"regional-a", "regional-b"} {
		for _, p := range regionalRuntimeProviders {
			knownKeys = append(knownKeys, regionalRuntimeKey(tenant, p))
		}
	}
}
func regionalRuntimePrefix(tenant string, p ai.ProviderID, api ai.API) string {
	return "/" + tenant + "/" + string(p) + "/" + string(api)
}
func regionalRuntimeTarget(p ai.ProviderID, api ai.API) ai.Target {
	return ai.Target{BindingID: string(p) + "-" + string(api), ModelID: mixedModel}
}
func regionalRuntimeCatalog() ai.Catalog {
	c := ai.Catalog{Version: "regional-runtime-synthetic-v1"}
	for _, p := range append(append([]ai.ProviderID{}, regionalRuntimeProviders...), ai.ProviderOpenAI, ai.ProviderDeepSeek) {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			c.Models = append(c.Models, ai.Model{Provider: p, API: api, ID: mixedModel, Name: "Synthetic regional runtime", Input: []ai.Modality{ai.ModalityText}, ContextWindow: 32768, MaxTokens: 4096, Compat: ai.ModelCompat{SupportsLongCacheRetention: ai.Value(true)}, Cost: ai.ModelCost{CostRates: ai.CostRates{Input: 1, Output: 2, CacheRead: 0.1}}})
		}
	}
	return c
}
func newRegionalRuntimeWorld(t *testing.T, configure func(*ai.Config)) *mixedWorld {
	t.Helper()
	w := newMixedWorld(t, func(c *ai.Config) {
		cat := regionalRuntimeCatalog()
		c.Catalog = &cat
		c.Policy = validPolicy()
		c.Policy.MaxQueuedObservations = 512
		if configure != nil {
			configure(c)
		}
	})
	for _, tenant := range []string{"regional-a", "regional-b"} {
		for _, p := range regionalRuntimeProviders {
			for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
				regionalRuntimeInstall(w, tenant, p, api, "same-account", regionalRuntimeKey(tenant, p))
			}
		}
	}
	return w
}
func regionalRuntimeInstall(w *mixedWorld, tenant string, p ai.ProviderID, api ai.API, account string, k tenantKey) {
	target := regionalRuntimeTarget(p, api)
	b := ai.Binding{TenantID: tenant, BindingID: target.BindingID, Operation: ai.OperationChat, Version: "b1", Enabled: true, ProviderID: p, API: api, Endpoint: w.provider.URL() + regionalRuntimePrefix(tenant, p, api), AuthKind: ai.AuthAPIKey, AccountScopeID: account, CredentialRef: "credential-" + string(p), AllowedModels: []string{mixedModel}}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: account, Active: true, APIKey: ai.NewSecret(k.secret)})
}
func regionalRuntimeReply(t *testing.T, api ai.API) provider.Reply {
	t.Helper()
	switch api {
	case ai.APIOpenAICompletions:
		return miniMaxChatReply(t, mixedModel)
	case ai.APIAnthropicMessages:
		return miniMaxMessagesReply(t, mixedModel)
	default:
		return minimaxResponsesReply(t, `{"type":"response.completed","response":{"id":"regional-runtime-response","object":"response","status":"completed","model":"same-model","output":[],"usage":{"input_tokens":12,"output_tokens":7,"total_tokens":19}}}`)
	}
}
func regionalRuntimeRecord(ev *evidence.Case, w *mixedWorld, o outcome) {
	o.record(ev)
	ev.Record(o.result.Metadata.RequestID+"-result", o.result)
	ev.Record(o.result.Metadata.RequestID+"-events", eventsJSON(o.events))
	w.record(ev, mixedOutcome{metadata: o.result.Metadata, usage: o.result.Message.Usage, chat: o.result, err: o.err})
	ev.Record("requests", w.provider.Requests())
}
func regionalRuntimeAdmitted(ev *evidence.Case, w *mixedWorld, tenant string, p ai.ProviderID, id string, n int) {
	got := []ai.AdmissionRequest{}
	for _, a := range w.adm.Requests() {
		if len(a.AttemptID) > len(id) && a.AttemptID[:len(id)+1] == id+"#" {
			got = append(got, a)
		}
	}
	ev.Record(id+"-admission", got)
	ev.Check("one admission per actual attempt", len(got) == n, "got %+v want %d", got, n)
	for i, a := range got {
		ev.Check("admission keeps actual provider and account", a.TenantID == tenant && a.AccountScopeID == "same-account" && a.AttemptID == fmt.Sprintf("%s#%d", id, i+1) && a.ProviderID == p, "got %+v want provider %s", a, p)
	}
}
func regionalRuntimeBodyKey(t *testing.T, r provider.Request) string {
	t.Helper()
	var b struct {
		Key string `json:"prompt_cache_key"`
	}
	if err := json.Unmarshal(r.Body, &b); err != nil {
		t.Fatal(err)
	}
	return b.Key
}

// Distinct response markers make concurrent response mixups observable at the
// public result boundary, in addition to credential/request attribution.
func regionalRuntimeMarkedReply(t *testing.T, api ai.API, marker string) provider.Reply {
	t.Helper()
	if api == ai.APIOpenAIResponses {
		item := fmt.Sprintf(`{"type":"message","id":"regional-message","role":"assistant","status":"completed","content":[{"type":"output_text","text":%s,"annotations":[]}]}`, mustMarshal(t, marker))
		return minimaxResponsesReply(t, `{"type":"response.output_item.done","output_index":0,"item":`+item+`}`, `{"type":"response.completed","response":{"id":"regional-response","object":"response","status":"completed","model":"same-model","output":[`+item+`],"usage":{"input_tokens":12,"output_tokens":7,"total_tokens":19}}}`)
	}
	r := regionalRuntimeReply(t, api)
	old := "Hello MiniMax."
	if api == ai.APIAnthropicMessages {
		old = "Synthetic reply."
	}
	for i, chunk := range r.Chunks {
		r.Chunks[i] = bytes.ReplaceAll(chunk, []byte(old), []byte(marker))
	}
	r.Script = string(bytes.Join(r.Chunks, nil))
	return r
}
