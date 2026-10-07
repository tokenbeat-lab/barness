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

	// anthropicA and anthropicB are each tenant's Anthropic key, on its own
	// vendor account.
	anthropicA = tenantKey{tenant: "tenant-a", secret: "sk-ant-test-tenant-a-0001", alias: "key:tenant-a-anthropic@v1"}
	anthropicB = tenantKey{tenant: "tenant-b", secret: "sk-ant-test-tenant-b-0001", alias: "key:tenant-b-anthropic@v1"}

	// googleA and googleB are each tenant's Gemini Developer API key, on
	// its own Google project, shaped like a real Google API key.
	googleA = tenantKey{tenant: "tenant-a", secret: "AIzaSyTestTenantA0001geminiKey000000000", alias: "key:tenant-a-google@v1"}
	googleB = tenantKey{tenant: "tenant-b", secret: "AIzaSyTestTenantB0001geminiKey000000000", alias: "key:tenant-b-google@v1"}

	// deepseekA and deepseekB are each tenant's DeepSeek key, on its own
	// DeepSeek account.
	deepseekA = tenantKey{tenant: "tenant-a", secret: "sk-test-deepseek-tenant-a-0001", alias: "key:tenant-a-deepseek@v1"}
	deepseekB = tenantKey{tenant: "tenant-b", secret: "sk-test-deepseek-tenant-b-0001", alias: "key:tenant-b-deepseek@v1"}

	// knownKeys are every key a scenario may install; the Provider reports
	// each by alias and the evidence bundle redacts each.
	knownKeys = []tenantKey{tenantA, tenantA2, tenantB, anthropicA, anthropicB, googleA, googleB, deepseekA, deepseekB}
)

// deepseekKey is tenant k's DeepSeek key.
func deepseekKey(k tenantKey) tenantKey {
	if k.tenant == deepseekB.tenant {
		return deepseekB
	}
	return deepseekA
}

// anthropicKey is tenant k's Anthropic key.
func anthropicKey(k tenantKey) tenantKey {
	if k.tenant == anthropicB.tenant {
		return anthropicB
	}
	return anthropicA
}

// googleKey is tenant k's Gemini key.
func googleKey(k tenantKey) tenantKey {
	if k.tenant == googleB.tenant {
		return googleB
	}
	return googleA
}

// builtinModelIDs are the built-in models of api, all allowed on its
// binding; with a provider, only that provider's.
func builtinModelIDs(api ai.API, provider ...ai.ProviderID) []string {
	var ids []string
	for _, m := range ai.BuiltinCatalog().Models {
		if m.API == api && (len(provider) == 0 || m.Provider == provider[0]) {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

// anthropicBinding is k's tenant's "claude" binding at version b1: Anthropic
// Messages at the world's Provider, on the tenant's Anthropic account.
func anthropicBinding(k tenantKey, providerURL string) ai.Binding {
	return ai.Binding{
		TenantID:       k.tenant,
		BindingID:      "claude",
		Version:        "b1",
		Enabled:        true,
		ProviderID:     ai.ProviderAnthropic,
		API:            ai.APIAnthropicMessages,
		Endpoint:       providerURL,
		AuthKind:       ai.AuthAPIKey,
		AccountScopeID: "acct-" + k.tenant + "-anthropic",
		CredentialRef:  "cred-" + k.tenant + "-anthropic",
		AllowedModels:  builtinModelIDs(ai.APIAnthropicMessages),
	}
}

// geminiBinding is k's tenant's "gemini" binding at version b1: the Gemini
// Developer API at the world's Provider (its version path, as the vendor's
// base URL carries it), on the tenant's Google project.
func geminiBinding(k tenantKey, providerURL string) ai.Binding {
	return ai.Binding{
		TenantID:       k.tenant,
		BindingID:      "gemini",
		Version:        "b1",
		Enabled:        true,
		ProviderID:     ai.ProviderGoogle,
		API:            ai.APIGoogleGenerativeAI,
		Endpoint:       providerURL + "/v1beta",
		AuthKind:       ai.AuthAPIKey,
		AccountScopeID: "acct-" + k.tenant + "-google",
		CredentialRef:  "cred-" + k.tenant + "-google",
		AllowedModels:  builtinModelIDs(ai.APIGoogleGenerativeAI),
	}
}

// geminiCredential is the active credential snapshot holding k's Gemini
// key, referenced by geminiBinding.
func geminiCredential(k tenantKey) ai.Credential {
	return ai.Credential{
		OwnerTenantID:  k.tenant,
		CredentialID:   "cred-" + k.tenant + "-google",
		Version:        "v1",
		AccountScopeID: "acct-" + k.tenant + "-google",
		Active:         true,
		APIKey:         ai.NewSecret(googleKey(k).secret),
	}
}

// anthropicCredential is the active credential snapshot holding k's
// Anthropic key, referenced by anthropicBinding.
func anthropicCredential(k tenantKey) ai.Credential {
	key := anthropicKey(k)
	return ai.Credential{
		OwnerTenantID:  k.tenant,
		CredentialID:   "cred-" + k.tenant + "-anthropic",
		Version:        "v1",
		AccountScopeID: "acct-" + k.tenant + "-anthropic",
		Active:         true,
		APIKey:         ai.NewSecret(key.secret),
	}
}

// deepseekBinding is k's tenant's "deepseek" binding at version b1:
// DeepSeek through the Responses protocol at the world's Provider (its root,
// as DeepSeek's base URL https://api.deepseek.com has no version path), on
// the tenant's DeepSeek account. gpt-4.1-mini is allowed but is OpenAI's
// model, absent from DeepSeek's catalog: it may not be called here.
func deepseekBinding(k tenantKey, providerURL string) ai.Binding {
	return ai.Binding{
		TenantID:       k.tenant,
		BindingID:      "deepseek",
		Version:        "b1",
		Enabled:        true,
		ProviderID:     ai.ProviderDeepSeek,
		API:            ai.APIOpenAIResponses,
		Endpoint:       providerURL,
		AuthKind:       ai.AuthAPIKey,
		AccountScopeID: "acct-" + k.tenant + "-deepseek",
		CredentialRef:  "cred-" + k.tenant + "-deepseek",
		AllowedModels:  []string{"deepseek-flash", "gpt-4.1-mini"},
	}
}

// deepseekChatBinding is k's tenant's "deepseek-chat" binding at version
// b1: DeepSeek through Chat Completions at the world's Provider root, on the
// same DeepSeek account and credential as "deepseek" — two bindings for one
// account's two protocols, each with its own model configuration.
// gpt-4.1-mini is allowed but absent from DeepSeek's catalog.
func deepseekChatBinding(k tenantKey, providerURL string) ai.Binding {
	return ai.Binding{
		TenantID:       k.tenant,
		BindingID:      "deepseek-chat",
		Version:        "b1",
		Enabled:        true,
		ProviderID:     ai.ProviderDeepSeek,
		API:            ai.APIOpenAICompletions,
		Endpoint:       providerURL,
		AuthKind:       ai.AuthAPIKey,
		AccountScopeID: "acct-" + k.tenant + "-deepseek",
		CredentialRef:  "cred-" + k.tenant + "-deepseek",
		AllowedModels:  append(builtinModelIDs(ai.APIOpenAICompletions, ai.ProviderDeepSeek), "gpt-4.1-mini"),
	}
}

// deepseekCredential is the active credential snapshot holding k's
// DeepSeek key, referenced by deepseekBinding and deepseekChatBinding.
func deepseekCredential(k tenantKey) ai.Credential {
	return ai.Credential{
		OwnerTenantID:  k.tenant,
		CredentialID:   "cred-" + k.tenant + "-deepseek",
		Version:        "v1",
		AccountScopeID: "acct-" + k.tenant + "-deepseek",
		Active:         true,
		APIKey:         ai.NewSecret(deepseekKey(k).secret),
	}
}

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
		// text-only model; the gpt-5 family are reasoning models, gpt-5.5-pro
		// one with tiered prices; gpt-5.4, gpt-5.4-pro, gpt-6-sol and
		// gpt-6-astra are the ones ADR-0018 listed that catalog.json calls.
		AllowedModels: []string{"gpt-4.1-mini", "gpt-4o-mini", "gpt-4", "gpt-legacy-x", "gpt-5", "gpt-5-mini", "gpt-5.1", "gpt-5.2", "gpt-5-pro", "gpt-5.5-pro",
			"gpt-5.4", "gpt-5.4-pro", "gpt-6-sol", "gpt-6-astra", "gpt-5.5", "gpt-6.1-sol", "gpt-daybreak-blue-latest", "gpt-daybreak-red-latest"},
	}
}

// chatBinding is k's tenant's "chat" binding at version b1: OpenAI Chat
// Completions at the world's Provider, on the same OpenAI account and
// credential as "primary" — two bindings for one account's two protocols.
func chatBinding(k tenantKey, providerURL string) ai.Binding {
	return ai.Binding{
		TenantID:       k.tenant,
		BindingID:      "chat",
		Version:        "b1",
		Enabled:        true,
		ProviderID:     ai.ProviderOpenAI,
		API:            ai.APIOpenAICompletions,
		Endpoint:       providerURL + "/v1",
		AuthKind:       ai.AuthAPIKey,
		AccountScopeID: "acct-" + k.tenant,
		CredentialRef:  "cred-" + k.tenant,
		AllowedModels:  builtinModelIDs(ai.APIOpenAICompletions, ai.ProviderOpenAI),
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
// double holding each tenant's same-named "primary" (Responses), "claude"
// (Anthropic Messages), "gemini" (Gemini Developer API), "chat" (OpenAI
// Chat Completions), "deepseek" (DeepSeek Responses) and "deepseek-chat"
// (DeepSeek Chat Completions) bindings, and a
// Client built with the loopback-only transport.
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
	}
	srv := provider.New(aliases)
	t.Cleanup(srv.Close)

	h := host.New()
	for _, k := range tenants {
		h.PutBinding(primaryBinding(k, srv.URL()))
		h.PutCredential(primaryCredential(k, "v1"))
		h.PutBinding(anthropicBinding(k, srv.URL()))
		h.PutCredential(anthropicCredential(k))
		h.PutBinding(geminiBinding(k, srv.URL()))
		h.PutCredential(geminiCredential(k))
		h.PutBinding(chatBinding(k, srv.URL()))
		h.PutBinding(deepseekBinding(k, srv.URL()))
		h.PutBinding(deepseekChatBinding(k, srv.URL()))
		h.PutCredential(deepseekCredential(k))
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
