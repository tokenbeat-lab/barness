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
	// tenantA2 is tenant A's key after rotating K1→K2; same credential, new version.
	tenantA2 = tenantKey{tenant: "tenant-a", secret: "sk-test-tenant-a-0002", alias: "key:tenant-a@v2"}

	// knownKeys are every key a scenario may install; the Provider reports
	// each by alias and the evidence bundle redacts each.
	knownKeys = []tenantKey{tenantA, tenantA2, tenantB}
)

// primaryBinding is k's tenant's "primary" binding at version b1 pointing at
// the world's Provider.
func primaryBinding(k tenantKey, providerURL string) ai.Binding {
	return ai.Binding{
		TenantID:       k.tenant,
		BindingID:      "primary",
		Version:        "b1",
		Enabled:        true,
		ProviderID:     ai.ProviderOpenAI,
		API:            ai.APIOpenAIResponses,
		Endpoint:       providerURL + "/v1",
		AuthKind:       ai.AuthAPIKey,
		AccountScopeID: "acct-" + k.tenant,
		CredentialRef:  "cred-" + k.tenant,
		// gpt-legacy-x is allowed but absent from the catalog; gpt-4.1 is in
		// the catalog but not allowed. Neither may be called. gpt-4 is the
		// text-only model; the gpt-5 family are reasoning models.
		AllowedModels: []string{"gpt-4.1-mini", "gpt-4o-mini", "gpt-4", "gpt-legacy-x", "gpt-5", "gpt-5.1", "gpt-5.2", "gpt-5-pro"},
	}
}

// primaryCredential is the active credential snapshot holding k's key at
// version, referenced by primaryBinding.
func primaryCredential(k tenantKey, version string) ai.Credential {
	return ai.Credential{
		OwnerTenantID:  k.tenant,
		CredentialID:   "cred-" + k.tenant,
		Version:        version,
		AccountScopeID: "acct-" + k.tenant,
		Active:         true,
		APIKey:         ai.NewSecret(k.secret),
	}
}

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
	for _, k := range knownKeys {
		aliases[k.secret] = k.alias
		run.RedactSecret(k.secret, k.alias)
	}
	srv := provider.New(aliases)
	t.Cleanup(srv.Close)

	h := host.New()
	for _, k := range tenants {
		h.PutBinding(primaryBinding(k, srv.URL()))
		h.PutCredential(primaryCredential(k, "v1"))
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
	Model        string `json:"model"`
	SystemPrompt string `json:"systemPrompt"`
	User         string `json:"user"`
	// SimpleMaxOutputTokens is pi's simple-entry max_output_tokens.
	SimpleMaxOutputTokens int               `json:"simpleMaxOutputTokens"`
	Events                []json.RawMessage `json:"events"`
	Expect                struct {
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

// entryBody is the body pi sends on an entry given the body it sends on the
// full entry without options: the simple entry adds its output budget
// (buildBaseOptions: the model's maximum clamped to the context window), which
// each fixture records as pi's simpleMaxOutputTokens.
func entryBody(t *testing.T, body json.RawMessage, simple bool, simpleMaxOutputTokens int) []byte {
	t.Helper()
	if !simple {
		return body
	}
	var fields map[string]json.RawMessage
	mustUnmarshal(t, body, &fields)
	fields["max_output_tokens"] = mustMarshal(t, simpleMaxOutputTokens)
	return mustMarshal(t, fields)
}

// updateBinding rewrites k's tenant's stored "primary" binding.
func (w *world) updateBinding(k tenantKey, change func(*ai.Binding)) {
	b := w.host.Binding(k.tenant, "primary")
	change(&b)
	w.host.PutBindingAt(k.tenant, "primary", b)
}

// replaceCredential stores a variant of k's v1 credential (k's key) in place
// of whatever is stored.
func (w *world) replaceCredential(k tenantKey, change func(*ai.Credential)) {
	c := primaryCredential(k, "v1")
	change(&c)
	w.host.PutCredentialAt(k.tenant, "cred-"+k.tenant, c)
}
