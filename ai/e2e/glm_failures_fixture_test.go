package e2e

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type glmHTTPCase struct {
	Name, Code, ShouldRetry   string
	Status, Retries, Requests int
	Body                      json.RawMessage
}

func glmFailureFixtures(t *testing.T) (map[string][]glmHTTPCase, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/glm-failures/http.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Protocols map[string][]glmHTTPCase }
	mustUnmarshal(t, raw, &fixture)
	return fixture.Protocols, raw
}

func glmFailureWorld(t *testing.T, protocol string, retries int) (*world, ai.Target, *provider.BodyTracker) {
	t.Helper()
	var cat ai.Catalog
	var binding func(string, string, string, string) ai.Binding
	var suffix string
	switch protocol {
	case "responses":
		cat, binding, suffix = glm.ResponsesCatalog(), glm.ResponsesBinding, "/api/v1"
	case "chat":
		cat, binding, suffix = glm.ChatCatalog(), glm.ChatBinding, "/api/paas/v4"
	case "messages":
		cat, binding, suffix = glm.MessagesCatalog(), glm.MessagesBinding, "/api/anthropic"
	default:
		t.Fatal(protocol)
	}
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	clk := clock.NewFake(time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC), 0)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog, c.Clock, c.Transport = &cat, clk, bodies }, tenantA)
	b := binding(tenantA.tenant, "acct-glm-failure-"+protocol, "cred-glm-failure-"+protocol, w.provider.URL()+suffix)
	b.Retry.MaxRetries = retries
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(glmA.secret)})
	return w, ai.Target{BindingID: b.BindingID}, bodies
}
