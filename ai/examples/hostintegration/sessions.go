package hostintegration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/tokenbeat-lab/barness/ai"
)

// Store is the host's session storage. Every operation is keyed by tenant:
// a session is only ever found under the tenant that created it, which is the
// host's authorization of history (spec I3: the host reads history by
// (tenant_id, session_id)). Records are the host's own persistence model,
// opaque bytes to the store.
//
// A durable store must also guarantee each record's integrity — database
// access control, or a MAC over (tenant, session, record) — because Turn
// vouches for the native state provenance it reads back.
type Store interface {
	// Create starts an empty session owned by tenant.
	Create(ctx context.Context, tenant string) (sessionID string, err error)
	// Load returns the session's records in order, or ErrNoSession when
	// tenant owns no session with that id.
	Load(ctx context.Context, tenant, sessionID string) ([][]byte, error)
	// Append adds records to tenant's session, or fails with ErrNoSession.
	Append(ctx context.Context, tenant, sessionID string, records ...[]byte) error
}

// ErrNoSession: the tenant owns no session with that id, whether or not
// another tenant does.
var ErrNoSession = errors.New("hostintegration: no such session")

// MemoryStore is an in-process Store. It is safe for concurrent use.
type MemoryStore struct {
	mu       sync.Mutex
	sessions map[sessionKey][][]byte
}

type sessionKey struct{ tenant, session string }

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: map[sessionKey][][]byte{}}
}

// Create mints a random session id. It is not requestid.New on purpose: a
// session id is a tenant-scoped storage key that lives across many logical
// calls, a RequestID names exactly one call; they must never be confused.
func (s *MemoryStore) Create(_ context.Context, tenant string) (string, error) {
	var b [12]byte
	_, _ = rand.Read(b[:])
	id := "sess-" + hex.EncodeToString(b[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionKey{tenant, id}] = nil
	return id, nil
}

func (s *MemoryStore) Load(_ context.Context, tenant, sessionID string) ([][]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	records, ok := s.sessions[sessionKey{tenant, sessionID}]
	if !ok {
		return nil, ErrNoSession
	}
	out := make([][]byte, len(records))
	for i, r := range records {
		out[i] = append([]byte(nil), r...)
	}
	return out, nil
}

func (s *MemoryStore) Append(_ context.Context, tenant, sessionID string, records ...[]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := sessionKey{tenant, sessionID}
	existing, ok := s.sessions[key]
	if !ok {
		return ErrNoSession
	}
	for _, r := range records {
		existing = append(existing, append([]byte(nil), r...))
	}
	s.sessions[key] = existing
	return nil
}

// record is one stored history message, the host's persistence model. It is
// converted to and from barness-ai's types at this boundary only (spec I1).
type record struct {
	Role string `json:"role"` // "user" or "assistant"
	Text string `json:"text,omitempty"`
	// Assistant fields.
	Blocks     []block       `json:"blocks,omitempty"`
	API        ai.API        `json:"api,omitempty"`
	Provider   ai.ProviderID `json:"provider,omitempty"`
	Model      string        `json:"model,omitempty"`
	ResponseID string        `json:"responseId,omitempty"`
	StopReason ai.StopReason `json:"stopReason,omitempty"`
	Usage      ai.Usage      `json:"usage,omitzero"`
	Timestamp  int64         `json:"timestamp,omitempty"`
	// Envelope is the provenance barness-ai vouched for when the turn was
	// produced. Stored as data; trusted again only through TrustNativeState.
	Envelope *ai.NativeStateEnvelope `json:"envelope,omitempty"`
}

// block is an assistant content block with its kind.
type block struct {
	Kind string          `json:"kind"` // "text", "thinking" or "toolCall"
	Data json.RawMessage `json:"data"`
}

func encodeUser(prompt string) ([]byte, error) {
	return json.Marshal(record{Role: "user", Text: prompt})
}

// encodeAssistant stores m with the provenance its native state came with.
func encodeAssistant(m ai.AssistantMessage) ([]byte, error) {
	rec := record{Role: "assistant", API: m.API, Provider: m.Provider, Model: m.Model, ResponseID: m.ResponseID,
		StopReason: m.StopReason, Usage: m.Usage, Timestamp: m.Timestamp}
	for _, c := range m.Content {
		var kind string
		switch c.(type) {
		case ai.Text:
			kind = "text"
		case ai.Thinking:
			kind = "thinking"
		case ai.ToolCall:
			kind = "toolCall"
		default:
			return nil, fmt.Errorf("hostintegration: cannot store content %T", c)
		}
		data, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		rec.Blocks = append(rec.Blocks, block{Kind: kind, Data: data})
	}
	if env, ok := m.NativeState.Envelope(); ok {
		rec.Envelope = &env
	}
	return json.Marshal(rec)
}

// decode turns a stored record back into a message of scope's call. A stored
// envelope is vouched for with scope: TrustNativeState refuses one naming
// another tenant or missing fields, and the message then stays untrusted —
// its native state is downgraded by the library, the turn goes ahead.
func decode(scope ai.CallScope, data []byte) (ai.Message, error) {
	var rec record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("hostintegration: corrupt history record: %w", err)
	}
	switch rec.Role {
	case "user":
		return ai.UserText(rec.Text), nil
	case "assistant":
	default:
		return nil, fmt.Errorf("hostintegration: corrupt history record: role %q", rec.Role)
	}
	m := ai.AssistantMessage{API: rec.API, Provider: rec.Provider, Model: rec.Model, ResponseID: rec.ResponseID,
		StopReason: rec.StopReason, Usage: rec.Usage, Timestamp: rec.Timestamp}
	for _, b := range rec.Blocks {
		c, err := decodeBlock(b)
		if err != nil {
			return nil, fmt.Errorf("hostintegration: corrupt history record: %w", err)
		}
		m.Content = append(m.Content, c)
	}
	if rec.Envelope != nil {
		if trusted, err := ai.TrustNativeState(scope, *rec.Envelope); err == nil {
			m.NativeState = trusted
		}
	}
	return m, nil
}

func decodeBlock(b block) (ai.AssistantContent, error) {
	switch b.Kind {
	case "text":
		var v ai.Text
		err := json.Unmarshal(b.Data, &v)
		return v, err
	case "thinking":
		var v ai.Thinking
		err := json.Unmarshal(b.Data, &v)
		return v, err
	case "toolCall":
		var v ai.ToolCall
		err := json.Unmarshal(b.Data, &v)
		return v, err
	}
	return nil, fmt.Errorf("block kind %q", b.Kind)
}
