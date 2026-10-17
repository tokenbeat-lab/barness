package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func newMiniMaxResponsesWorld(t *testing.T, mutate func(*ai.Catalog)) *world {
	t.Helper()
	cat := minimax.ResponsesCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	b := minimax.ResponsesBinding(tenantA.tenant, "acct-minimax-responses", "cred-minimax-responses", w.provider.URL()+"/v1")
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantA.secret)})
	return w
}

func minimaxResponsesTarget(model string) ai.Target {
	return ai.Target{BindingID: "minimax-responses", ModelID: model}
}
func minimaxResponsesRequest() ai.Request {
	return ai.Request{Messages: []ai.Message{ai.UserText("Say hello.")}}
}
func minimaxResponsesReply(t *testing.T, events ...string) provider.Reply {
	t.Helper()
	raw := make([]json.RawMessage, len(events))
	for i, e := range events {
		raw[i] = json.RawMessage(e)
	}
	return sseEvents(t, raw, provider.FramingLF)
}

const minimaxResponsesCompleted = `{"type":"response.completed","response":{"id":"r-minimax","model":"MiniMax-M3","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}}`
