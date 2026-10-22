package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"testing"
)

var kimiResponsesKey = tenantKey{tenant: "tenant-a", secret: "synthetic-kimi-responses-key", alias: "key:kimi-responses@v1"}

func init() { knownKeys = append(knownKeys, kimiResponsesKey) }
func newKimiResponsesWorld(t *testing.T, market ai.ProviderID, mutate func(*ai.Catalog), configure ...func(*ai.Config)) *world {
	t.Helper()
	cat := kimi.CandidateCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		for _, f := range configure {
			f(c)
		}
	}, tenantA)
	b := ai.Binding{TenantID: tenantA.tenant, BindingID: "kimi-responses", Version: "v1", Enabled: true, Operation: ai.OperationChat, ProviderID: market, API: ai.APIOpenAIResponses, Endpoint: w.provider.URL() + "/v1", AuthKind: ai.AuthAPIKey, AccountScopeID: "acct-kimi-responses", CredentialRef: "cred-kimi-responses", AllowedModels: []string{"kimi-k3"}}
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(kimiResponsesKey.secret)})
	return w
}
func kimiResponsesTarget() ai.Target {
	return ai.Target{BindingID: "kimi-responses", ModelID: "kimi-k3"}
}
func kimiResponsesRequest() ai.Request {
	return ai.Request{SystemPrompt: "Be terse.", Messages: []ai.Message{ai.UserText("Say hello.")}}
}

const kimiResponsesReasoning = `{"id":"reason|α","type":"reasoning","status":"completed","summary":[{"type":"summary_text","text":"Check."},{"type":"summary_text","text":"Then answer."}],"encrypted_content":null}`
const kimiResponsesMessage = `{"id":"message|α","type":"message","role":"assistant","status":"completed","phase":"final_answer","content":[{"type":"output_text","text":"Hello Kimi.","annotations":[]}]}`
const kimiResponsesUsage = `{"input_tokens":200,"output_tokens":50,"total_tokens":250,"input_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"output_tokens_details":{"reasoning_tokens":10}}`

func kimiResponsesFrames() []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"type":"response.created","response":{"id":"kimi-responses-synthetic","model":"kimi-k3","status":"in_progress","output":[]}}`),
		json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"type":"reasoning","id":"reason|α","summary":[],"encrypted_content":null}}`),
		json.RawMessage(`{"type":"response.reasoning_summary_text.delta","output_index":0,"content_index":0,"delta":"Check."}`),
		json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":` + kimiResponsesReasoning + `}`),
		json.RawMessage(`{"type":"response.output_item.added","output_index":1,"item":{"type":"message","id":"message|α","role":"assistant","content":[]}}`),
		json.RawMessage(`{"type":"response.output_text.delta","output_index":1,"content_index":0,"delta":"Hello Kimi."}`),
		json.RawMessage(`{"type":"response.output_item.done","output_index":1,"item":` + kimiResponsesMessage + `}`),
		json.RawMessage(`{"type":"response.completed","response":{"id":"kimi-responses-synthetic","model":"kimi-k3","status":"completed","output":[` + kimiResponsesReasoning + `,` + kimiResponsesMessage + `],"usage":` + kimiResponsesUsage + `}}`),
	}
}
func kimiResponsesReply(t *testing.T, frames ...json.RawMessage) provider.Reply {
	t.Helper()
	if len(frames) == 0 {
		frames = kimiResponsesFrames()
	}
	return sseEvents(t, frames, provider.FramingLF)
}
func kimiResponsesID(m ai.ProviderID, group string) string {
	if m == ai.ProviderMoonshotCN {
		return "KIMI-CN-RESPONSES-" + group
	}
	return "KIMI-INTL-RESPONSES-" + group
}
