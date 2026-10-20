package e2e

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var glmA = tenantKey{tenant: "tenant-a", secret: "synthetic-glm-tenant-a", alias: "key:tenant-a-glm@v1"}

func init() { knownKeys = append(knownKeys, glmA) }

func newGLMChatWorld(t *testing.T, mutate func(*ai.Catalog)) *world {
	t.Helper()
	cat := glm.ChatCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	b := glm.ChatBinding(tenantA.tenant, "acct-tenant-a-glm", "cred-tenant-a-glm", w.provider.URL()+"/api/paas/v4")
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef,
		Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(glmA.secret)})
	return w
}

func glmChatReply(t *testing.T, model string) (provider.Reply, []byte) {
	t.Helper()
	events, raw := glmChatFrames(t, model)
	return chatEvents(t, events, ""), raw
}

func glmChatFrames(t *testing.T, model string) ([]json.RawMessage, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-chat/text.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Scenarios []struct {
			Model  string            `json:"model"`
			Events []json.RawMessage `json:"events"`
		} `json:"scenarios"`
	}
	mustUnmarshal(t, raw, &fixture)
	for _, sc := range fixture.Scenarios {
		if sc.Model == model {
			return sc.Events, raw
		}
	}
	t.Fatalf("missing GLM fixture %s", model)
	return nil, nil
}

func glmChatRequest() ai.Request {
	return ai.Request{SystemPrompt: "Be terse.", Messages: []ai.Message{
		ai.AssistantMessage{API: ai.APIOpenAICompletions, Provider: ai.ProviderID("zhipu"), Model: "glm-5.3", Content: []ai.AssistantContent{ai.Text{Text: "Earlier."}}},
		ai.UserText("Say hello."),
	}}
}
