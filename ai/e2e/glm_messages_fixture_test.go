package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type glmMessagesScenario struct {
	Model, SSE, Frames, Source, SSESha256, Text, ResponseID string
	Content                                                 json.RawMessage
	Input, Output                                           int64
	RequestID                                               []string
}

func loadGLMMessages(t *testing.T) ([]glmMessagesScenario, string, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-messages/text.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Prompt    string
		Scenarios []glmMessagesScenario
	}
	mustUnmarshal(t, raw, &fixture)
	return fixture.Scenarios, fixture.Prompt, raw
}

func newGLMMessagesWorld(t *testing.T, mutate func(*ai.Catalog)) *world {
	t.Helper()
	cat := glm.MessagesCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	b := glm.MessagesBinding(tenantA.tenant, "acct-tenant-a-glm-messages", "cred-tenant-a-glm-messages", w.provider.URL()+"/api/anthropic")
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef,
		Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(glmA.secret)})
	return w
}

func glmMessagesReply(t *testing.T, sc glmMessagesScenario) provider.Reply {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-messages/" + sc.SSE)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	if hex.EncodeToString(hash[:]) != sc.SSESha256 {
		t.Fatal("Messages raw SSE changed from immutable live capture")
	}
	r := provider.SSE([][]byte{raw})
	r.Header["X-Request-Id"] = sc.RequestID[0]
	return r
}
