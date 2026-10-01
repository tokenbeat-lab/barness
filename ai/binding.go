package ai

import (
	"context"
	"fmt"
)

// CallScope is the trusted identity a host establishes for one logical call.
// TenantID and RequestID are required. The library never derives a scope from
// request content, the environment or the context.
type CallScope struct {
	TenantID string
	// RequestID identifies one public logical call; internal retries reuse it
	// and a new call must use a new one. It is not a deduplication key.
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

// ProviderOpenAI is OpenAI.
const ProviderOpenAI ProviderID = "openai"

// API identifies a wire protocol. Values match the frozen pi-ai API names.
type API string

// APIOpenAIResponses is the OpenAI Responses protocol.
const APIOpenAIResponses API = "openai-responses"

// AuthKind is how a binding authenticates. The first phase supports API keys only.
type AuthKind string

// AuthAPIKey authenticates with a bearer API key.
const AuthAPIKey AuthKind = "api_key"

// Binding is one service configuration a tenant may use, located by
// (TenantID, BindingID). A binding fixes exactly one API.
type Binding struct {
	TenantID       string
	BindingID      string
	Version        string
	ProviderID     ProviderID
	API            API
	Endpoint       string // base URL, e.g. https://api.openai.com/v1
	AuthKind       AuthKind
	AccountScopeID string
	CredentialRef  string
	// AllowedModels is intersected with the model catalog; only models in both
	// can be called.
	AllowedModels []string
}

// Credential is a versioned credential snapshot. Its secret never enters
// messages, events, results, logs or errors.
type Credential struct {
	OwnerTenantID string
	CredentialID  string
	Version       string
	APIKey        Secret
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

// BindingResolver is implemented by the trusted host. It locates the binding
// by (scope.TenantID, bindingID) and verifies the scope may use it.
type BindingResolver interface {
	ResolveBinding(ctx context.Context, scope CallScope, bindingID string) (Binding, error)
}

// CredentialResolver is implemented by the trusted host. It returns the
// credential snapshot the binding references. The library does not cache it.
type CredentialResolver interface {
	ResolveCredential(ctx context.Context, scope CallScope, binding Binding) (Credential, error)
}
