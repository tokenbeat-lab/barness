package ai

import (
	"context"
	"errors"
	"fmt"
)

// CallScope is the trusted identity a host establishes for one logical call.
// TenantID and RequestID are required. The library never derives a scope from
// request content, the environment or the context.
type CallScope struct {
	TenantID string
	// RequestID identifies one public logical call; internal retries reuse it
	// and a new call must use a new one. It is globally unique: no two
	// logical calls share one, whatever their tenants. The host guarantees
	// this; the library cannot check it. It is not a deduplication key.
	RequestID string
	ActorID   string // optional authorized user or service principal
	JobID     string // optional host job correlation
}

// Target names what to call: a binding owned by the scope's tenant and a model
// within it. Protocol, endpoint, account and credential come from the binding.
type Target struct {
	BindingID string
	ModelID   string
}

// ProviderID identifies the real service operator, independent of protocol.
type ProviderID string

const (
	// ProviderOpenAI is OpenAI.
	ProviderOpenAI ProviderID = "openai"
	// ProviderAnthropic is Anthropic.
	ProviderAnthropic ProviderID = "anthropic"
	// ProviderGoogle is Google, through the Gemini Developer API.
	ProviderGoogle ProviderID = "google"
)

// API identifies a wire protocol. Values match the frozen pi-ai API names.
type API string

const (
	// APIOpenAIResponses is the OpenAI Responses protocol.
	APIOpenAIResponses API = "openai-responses"
	// APIAnthropicMessages is the Anthropic Messages protocol.
	APIAnthropicMessages API = "anthropic-messages"
	// APIGoogleGenerativeAI is the Gemini Developer API's generateContent
	// protocol (pi-ai's google-generative-ai), not Vertex AI.
	APIGoogleGenerativeAI API = "google-generative-ai"
)

// AuthKind is how a binding authenticates. The first phase supports API keys only.
type AuthKind string

// AuthAPIKey authenticates with a bearer API key.
const AuthAPIKey AuthKind = "api_key"

// Binding is one service configuration a tenant may use, located by
// (TenantID, BindingID). A binding fixes exactly one API.
type Binding struct {
	TenantID  string
	BindingID string
	// Version identifies this configuration snapshot; it is required so the
	// credential read can be tied to it (ADR-0003).
	Version string
	// Enabled must be true for the binding to be used. The zero value refuses,
	// so a host that forgets to set it fails closed.
	Enabled    bool
	ProviderID ProviderID
	API        API
	// Endpoint is the protocol's base URL, as pi-ai's model baseUrl: e.g.
	// https://api.openai.com/v1 for Responses, https://api.anthropic.com for
	// Anthropic Messages.
	Endpoint       string
	AuthKind       AuthKind
	AccountScopeID string
	CredentialRef  string
	// AllowedModels is intersected with the model catalog; only models in both
	// can be called.
	AllowedModels []string
	// AllowedHostedTools are the provider-hosted tool types, as the protocol
	// names them (e.g. Responses "web_search"), that a trusted payload
	// callback may declare on calls through this binding. Naming a type is
	// the host vouching that the account's hosted resources and the network
	// targets the tool reaches are this tenant's to use (ADR-0005). Empty, the
	// default, refuses every hosted tool; barness-ai never declares one itself.
	AllowedHostedTools []string
	// Retry is the operator's retry policy for calls through this binding;
	// the zero value never retries. See RetryPolicy.
	Retry RetryPolicy
}

// Credential is a versioned credential snapshot together with its tenant and
// account ownership. Its secret never enters messages, events, results, logs
// or errors.
type Credential struct {
	OwnerTenantID string
	CredentialID  string
	// Version is required; an unversioned snapshot cannot be proven consistent.
	Version string
	// BindingVersion is the binding version the credential backend paired this
	// snapshot with, read from its own current record — not copied from the
	// binding it was handed. It must equal the version the call resolved;
	// otherwise the binding changed between the two reads (D2, ADR-0003).
	BindingVersion string
	// AccountScopeID is the vendor account the key belongs to; it must match
	// the binding's.
	AccountScopeID string
	// Active must be true for the credential to be used. The zero value
	// refuses, so a revoked or unknown state fails closed.
	Active bool
	APIKey Secret
}

// Secret holds secret material. Formatting and JSON encoding never reveal it.
type Secret struct{ value string }

// NewSecret wraps secret material.
func NewSecret(value string) Secret { return Secret{value: value} }

func (s Secret) reveal() string { return s.value }

// String implements fmt.Stringer without revealing the value.
func (Secret) String() string { return "[REDACTED]" }

// GoString implements fmt.GoStringer without revealing the value.
func (Secret) GoString() string { return "ai.Secret{[REDACTED]}" }

// Format makes every fmt verb, including %v on enclosing structs, redact.
func (s Secret) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte(s.String())) }

// MarshalJSON encodes the redaction marker, never the value.
func (Secret) MarshalJSON() ([]byte, error) { return []byte(`"[REDACTED]"`), nil }

// Errors a trusted resolver may return to classify a refusal. Any other
// resolver error is treated as the binding or credential being unavailable;
// its text is never surfaced, so backend internals cannot leak through it.
var (
	// ErrAccessDenied: the scope's actor may not use the binding or credential.
	// Classified as CodeTenantDenied.
	ErrAccessDenied = errors.New("ai: access denied")
	// ErrSnapshotConflict: the credential backend cannot pair the binding
	// version it was given with a credential consistently, e.g. because the
	// binding was updated or disabled after it was read. The call fails in
	// PhaseConsistency without re-resolving (D2, ADR-0003); the host may start
	// a new logical call with a new RequestID.
	ErrSnapshotConflict = errors.New("ai: configuration snapshot conflict")
)

// BindingResolver is implemented by the trusted host. It locates the binding
// by (scope.TenantID, bindingID) and verifies the scope may use it. Every
// logical call resolves afresh; the library never caches a binding.
type BindingResolver interface {
	ResolveBinding(ctx context.Context, scope CallScope, bindingID string) (Binding, error)
}

// CredentialResolver is implemented by the trusted host. It returns the
// credential snapshot the binding references, stamped with the binding
// version the backend paired it with (Credential.BindingVersion); the library
// rejects the snapshot when that differs from the version it resolved. A
// backend may instead return ErrSnapshotConflict itself. The library does not
// cache it.
type CredentialResolver interface {
	ResolveCredential(ctx context.Context, scope CallScope, binding Binding) (Credential, error)
}
