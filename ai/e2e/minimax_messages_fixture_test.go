package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func newMiniMaxMessagesWorld(t *testing.T, mutate func(*ai.Catalog)) *world {
	t.Helper()
	return newMiniMaxMessagesConfigured(t, mutate, nil)
}

func newMiniMaxMessagesConfigured(t *testing.T, mutate func(*ai.Catalog), configure func(*ai.Config)) *world {
	t.Helper()
	cat := minimax.MessagesCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		if configure != nil {
			configure(c)
		}
	}, tenantA)
	b := minimax.MessagesBinding(tenantA.tenant, "acct-minimax-messages", "cred-minimax-messages", w.provider.URL()+"/anthropic")
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(anthropicA.secret)})
	return w
}

func miniMaxMessagesReply(t *testing.T, model string) provider.Reply {
	t.Helper()
	raw, err := os.ReadFile("testdata/minimax-messages/text.sse")
	if err != nil {
		t.Fatal(err)
	}
	// Templates carry a model token only; JSON and event order are independent
	// synthetic Messages fixtures, not a statement of live vendor behavior.
	raw = []byte(stringsReplaceModel(string(raw), model))
	r := provider.SSE([][]byte{raw})
	r.Header["Request-Id"] = "synthetic-minimax-request"
	return r
}

func stringsReplaceModel(s, model string) string {
	return strings.ReplaceAll(s, "MODEL", model)
}

func miniMaxMessagesTarget(model string) ai.Target {
	return ai.Target{BindingID: "minimax-messages", ModelID: model}
}
func miniMaxMessagesRequest() ai.Request {
	return ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input.")}}
}

func miniMaxMessagesFields(t *testing.T, body []byte) map[string]json.RawMessage {
	return bodyFields(t, body)
}

func miniMaxMessagesFixture(t *testing.T, file, model string) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/minimax-messages/" + file)
	if err != nil {
		t.Fatal(err)
	}
	return []byte(strings.ReplaceAll(string(raw), "MODEL", model))
}
