// Package hostintegration is barness-ai's minimal cloud host integration (spec
// I3, Testing Decisions §1, User Story 49). It shows what a host does around
// one generation turn and what it leaves to the library:
//
//   - Identity: the scope comes from the authentication result (Principal),
//     never from the request. A request claiming another tenant is refused
//     here, before barness-ai is called (ErrForbidden).
//   - History: the host reads it by (tenant, session) from its own store. A
//     session id that is not the caller's tenant's is refused here; the
//     library cannot tell whose text a message is.
//   - Native state: messages read back from storage carry no trust. The host
//     vouches for the stored provenance with ai.TrustNativeState after its
//     store's own checks; the library then replays native state only where
//     that provenance matches the call, and downgrades it otherwise.
//   - Cancellation: the turn runs under the downstream request's context. A
//     client that disconnects, or a send that fails, cancels this turn only;
//     every other turn keeps running.
//   - Attribution: every event goes downstream in its ai.EventEnvelope, so a
//     consumer of several turns (Merge) routes by the envelope, never by
//     arrival order. Which envelope fields leave the host is its decision.
//
// What the library itself refuses — a scope without TenantID or RequestID, a
// binding the tenant does not own, a model the binding does not allow — it
// refuses whatever the host does; the host boundary is an additional,
// separate responsibility.
package hostintegration

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokenbeat-lab/barness/ai"
)

// Principal is an authentication result: who is calling, as the host's
// authentication established it (a verified token, a session cookie). How
// it is established is out of scope here.
type Principal struct {
	TenantID string
	ActorID  string
}

// TurnRequest is a client's request for one turn, decoded from untrusted
// JSON. Nothing in it is an identity: TenantID is at most a claim, and
// SessionID only a reference the host resolves within the caller's tenant.
type TurnRequest struct {
	// TenantID, when set, must be the caller's own tenant.
	TenantID  string `json:"tenantId,omitempty"`
	SessionID string `json:"sessionId"`
	BindingID string `json:"bindingId"`
	ModelID   string `json:"modelId"`
	Prompt    string `json:"prompt"`
}

// Sink is the downstream connection of one turn, such as a server-sent
// events response. A Send error means the client is gone.
type Sink interface {
	Send(ai.EventEnvelope) error
}

var (
	// ErrForbidden is the host boundary's refusal of a request that claims
	// another tenant or references a session the caller's tenant does not
	// own. Nothing reaches barness-ai.
	ErrForbidden = errors.New("hostintegration: forbidden")
	// ErrBadRequest refuses a malformed request before barness-ai is called.
	ErrBadRequest = errors.New("hostintegration: bad request")
	// ErrDownstream reports a turn canceled because its downstream send
	// failed.
	ErrDownstream = errors.New("hostintegration: downstream send failed")
)

// Host runs turns for authenticated callers.
type Host struct {
	client   *ai.Client
	sessions Store
	// newRequestID mints a globally unique RequestID per logical call
	// (requestid.New in production).
	newRequestID func() string
}

// New returns a Host using client, sessions and newRequestID.
func New(client *ai.Client, sessions Store, newRequestID func() string) *Host {
	return &Host{client: client, sessions: sessions, newRequestID: newRequestID}
}

// OpenSession starts a session owned by who's tenant.
func (h *Host) OpenSession(ctx context.Context, who Principal) (string, error) {
	return h.sessions.Create(ctx, who.TenantID)
}

// Turn runs one generation turn of in's session for who, sending every
// event to sink, and stores the turn when it succeeded. ctx is the
// downstream request's context: when the client disconnects, the turn is
// canceled. The returned Result is barness-ai's, also for failed turns; it is
// the zero Result when the host boundary refused the request.
func (h *Host) Turn(ctx context.Context, who Principal, in TurnRequest, sink Sink) (ai.Result, error) {
	scope, err := h.scope(who, in)
	if err != nil {
		return ai.Result{}, err
	}
	history, err := h.history(ctx, scope, in.SessionID)
	if err != nil {
		return ai.Result{}, err
	}
	prompt := ai.UserText(in.Prompt)
	req := ai.Request{Messages: append(history, prompt)}
	s := h.client.Stream(ctx, scope, ai.Target{BindingID: in.BindingID, ModelID: in.ModelID}, req, nil)
	defer s.Close()
	for s.Next() {
		if err := sink.Send(s.Envelope()); err != nil {
			// Only this turn's generation ends; Close returns its permits and
			// connection at once.
			_ = s.Close()
			res, _ := s.Result()
			return res, fmt.Errorf("%w: %w", ErrDownstream, err)
		}
	}
	res, err := s.Result()
	if err != nil {
		return res, err
	}
	user, err := encodeUser(in.Prompt)
	if err != nil {
		return res, err
	}
	assistant, err := encodeAssistant(res.Message)
	if err != nil {
		return res, err
	}
	return res, h.sessions.Append(ctx, scope.TenantID, in.SessionID, user, assistant)
}

// scope builds the call's trusted scope from the authentication result. The
// request's claimed tenant is checked, never used.
func (h *Host) scope(who Principal, in TurnRequest) (ai.CallScope, error) {
	if who.TenantID == "" {
		return ai.CallScope{}, fmt.Errorf("%w: unauthenticated", ErrForbidden)
	}
	if in.TenantID != "" && in.TenantID != who.TenantID {
		return ai.CallScope{}, fmt.Errorf("%w: the request claims another tenant", ErrForbidden)
	}
	if in.SessionID == "" || in.BindingID == "" || in.ModelID == "" || in.Prompt == "" {
		return ai.CallScope{}, fmt.Errorf("%w: sessionId, bindingId, modelId and prompt are required", ErrBadRequest)
	}
	return ai.CallScope{TenantID: who.TenantID, ActorID: who.ActorID, RequestID: h.newRequestID()}, nil
}

// history reads the session of scope's tenant and turns its records back
// into barness-ai messages. Stored native state is vouched for with the
// scope; a record whose provenance cannot be vouched for stays untrusted and
// is downgraded by the library rather than refused.
func (h *Host) history(ctx context.Context, scope ai.CallScope, sessionID string) ([]ai.Message, error) {
	records, err := h.sessions.Load(ctx, scope.TenantID, sessionID)
	if errors.Is(err, ErrNoSession) {
		return nil, fmt.Errorf("%w: unknown session", ErrForbidden)
	}
	if err != nil {
		return nil, err
	}
	out := make([]ai.Message, 0, len(records))
	for _, rec := range records {
		m, err := decode(scope, rec)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
