package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"os"
	"testing"
)

func newGLMResponsesWorld(t *testing.T, mutate func(*ai.Catalog)) *world {
	t.Helper()
	cat := glm.ResponsesCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	b := glm.ResponsesBinding(tenantA.tenant, "acct-tenant-a-glm-responses", "cred-glm-responses", w.provider.URL()+"/api/v1")
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(glmA.secret)})
	return w
}

type glmResponsesScenario struct {
	Model      string            `json:"model"`
	Name       string            `json:"name"`
	Events     []json.RawMessage `json:"events"`
	Thinking   string            `json:"thinking"`
	Text       string            `json:"text"`
	Native     json.RawMessage   `json:"native"`
	StopReason string            `json:"stopReason"`
	Failure    bool              `json:"failure"`
}

func glmResponsesFixtures(t *testing.T) ([]glmResponsesScenario, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-responses/text.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Scenarios []glmResponsesScenario `json:"scenarios"`
	}
	mustUnmarshal(t, raw, &f)
	return f.Scenarios, raw
}

func glmResponsesReply(t *testing.T, events []json.RawMessage) provider.Reply {
	return sseEvents(t, events, provider.FramingLF)
}

func glmResponsesTarget(model string) ai.Target {
	return ai.Target{BindingID: "glm-responses", ModelID: model}
}
func glmResponsesRequest() ai.Request {
	return ai.Request{Messages: []ai.Message{ai.UserText("Say hello.")}}
}
