package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
)

var xaiKey = tenantKey{tenant: "tenant-a", secret: "synthetic-xai-responses-key", alias: "key:xai-responses@v1"}

func init() { knownKeys = append(knownKeys, xaiKey) }

// xaiTransport redirects the configured official target only at the injected
// transport seam. Binding validation never receives a loopback exception.
type xaiTransport struct {
	target *url.URL
	next   http.RoundTripper
}

func (r *xaiTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Hostname() != "api.x.ai" {
		return r.next.RoundTrip(req)
	}
	if req.URL.Scheme != "https" || req.URL.Path != "/v1/responses" || req.URL.RawQuery != "" {
		return nil, fmt.Errorf("unexpected xAI request target")
	}
	copy := req.Clone(req.Context())
	u := *req.URL
	u.Scheme = r.target.Scheme
	u.Host = r.target.Host
	copy.URL = &u
	return r.next.RoundTrip(copy)
}

func newXAIWorld(t *testing.T, mutate func(*ai.Catalog), configure ...func(*ai.Config)) *world {
	t.Helper()
	cat := xai.CandidateCatalog()
	if mutate != nil {
		mutate(&cat)
	}
	transport := &xaiTransport{next: provider.LoopbackTransport()}
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		c.Transport = transport
		for _, f := range configure {
			f(c)
		}
	}, tenantA)
	var err error
	transport.target, err = url.Parse(w.provider.URL())
	if err != nil {
		t.Fatal(err)
	}
	b := xaiBinding()
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(xaiKey.secret)})
	return w
}
func xaiBinding() ai.Binding {
	return ai.Binding{TenantID: tenantA.tenant, BindingID: "xai-responses", Version: "v1", Enabled: true, Operation: ai.OperationChat, ProviderID: ai.ProviderXAI, API: ai.APIOpenAIResponses, Endpoint: "https://api.x.ai/v1", AuthKind: ai.AuthAPIKey, AccountScopeID: "acct-xai", CredentialRef: "cred-xai", AllowedModels: []string{"grok-4.7"}}
}
func xaiTarget() ai.Target { return ai.Target{BindingID: "xai-responses", ModelID: "grok-4.7"} }
func xaiRequest() ai.Request {
	return ai.Request{SystemPrompt: "Be terse.", Messages: []ai.Message{ai.UserText("Say hello.")}}
}

const xaiMessage = `{"id":"message|α","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"Hello Grok.","annotations":[]}]}`
const xaiUsage = `{"input_tokens":32,"output_tokens":9,"total_tokens":41,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`

func xaiFrames() []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"message|α","type":"message","role":"assistant","content":[]}}`),
		json.RawMessage(`{"type":"response.output_text.delta","output_index":0,"item_id":"message|α","content_index":0,"delta":"Hello Grok."}`),
		json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":` + xaiMessage + `}`),
		json.RawMessage(`{"type":"response.completed","response":{"id":"xai-synthetic-response","model":"grok-4.7","status":"completed","output":[` + xaiMessage + `],"usage":` + xaiUsage + `}}`),
	}
}
func xaiReply(t *testing.T, frames ...json.RawMessage) provider.Reply {
	t.Helper()
	if len(frames) == 0 {
		frames = xaiFrames()
	}
	return sseEvents(t, frames, provider.FramingLF)
}
