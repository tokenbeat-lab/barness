package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// tenantKey is one tenant's test credential. Secrets are synthetic; reports
// only ever show the alias.
type tenantKey struct {
	tenant, secret, alias string
}

var (
	tenantA = tenantKey{tenant: "tenant-a", secret: "sk-test-tenant-a-0001", alias: "key:tenant-a@v1"}
	tenantB = tenantKey{tenant: "tenant-b", secret: "sk-test-tenant-b-0001", alias: "key:tenant-b@v1"}
)

// world is one scenario's assembly: a local controlled Provider, a trusted-host
// double holding each tenant's same-named "primary" binding, and a Client
// built with the loopback-only transport.
type world struct {
	provider *provider.Server
	host     *host.Host
	client   *ai.Client
}

func newWorld(t *testing.T, tenants ...tenantKey) *world {
	t.Helper()
	return newWorldWith(t, func(*ai.Config) {}, tenants...)
}

// newWorldWith lets a scenario adjust the Client config before construction.
func newWorldWith(t *testing.T, configure func(*ai.Config), tenants ...tenantKey) *world {
	t.Helper()
	aliases := map[string]string{}
	for _, k := range tenants {
		aliases[k.secret] = k.alias
		run.RedactSecret(k.secret, k.alias)
	}
	srv := provider.New(aliases)
	t.Cleanup(srv.Close)

	h := host.New()
	for _, k := range tenants {
		h.PutBinding(ai.Binding{
			TenantID:       k.tenant,
			BindingID:      "primary",
			Version:        "b1",
			ProviderID:     ai.ProviderOpenAI,
			API:            ai.APIOpenAIResponses,
			Endpoint:       srv.URL() + "/v1",
			AuthKind:       ai.AuthAPIKey,
			AccountScopeID: "acct-" + k.tenant,
			CredentialRef:  "cred-" + k.tenant,
			// gpt-legacy-x is allowed but absent from the catalog; gpt-4.1 is in
			// the catalog but not allowed. Neither may be called.
			AllowedModels: []string{"gpt-4.1-mini", "gpt-4o-mini", "gpt-legacy-x"},
		})
		h.PutCredential(ai.Credential{
			OwnerTenantID: k.tenant,
			CredentialID:  "cred-" + k.tenant,
			Version:       "v1",
			APIKey:        ai.NewSecret(k.secret),
		})
	}

	cfg := ai.Config{
		Policy:      validPolicy(),
		Bindings:    h,
		Credentials: h,
		Transport:   provider.LoopbackTransport(),
		// Test assembly: the local controlled Provider is plain http on loopback.
		AllowLoopbackHTTP: true,
	}
	configure(&cfg)
	client, err := ai.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &world{provider: srv, host: h, client: client}
}

// textFixture is testdata/responses/text-*.json.
type textFixture struct {
	Model        string            `json:"model"`
	SystemPrompt string            `json:"systemPrompt"`
	User         string            `json:"user"`
	Events       []json.RawMessage `json:"events"`
	Expect       struct {
		Request struct {
			Method string          `json:"method"`
			Path   string          `json:"path"`
			Query  string          `json:"query"`
			Body   json.RawMessage `json:"body"`
		} `json:"request"`
		Deltas        []string `json:"deltas"`
		Text          string   `json:"text"`
		TextSignature string   `json:"textSignature"`
		ResponseID    string   `json:"responseId"`
		StopReason    string   `json:"stopReason"`
		Usage         ai.Usage `json:"usage"`
	} `json:"expect"`
}

func loadTextFixture(t *testing.T, name string) (textFixture, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "responses", name))
	if err != nil {
		t.Fatal(err)
	}
	var f textFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return f, raw
}

// jsonEqual compares JSON semantically: object keys are unordered, arrays
// keep their order, and null/zero/absent stay distinct.
func jsonEqual(a, b []byte) bool {
	var x, y any
	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
