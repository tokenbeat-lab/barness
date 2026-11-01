package maintenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// Identity always includes the operation and protocol; names alone are not
// catalog keys or permission grants.
type Identity struct {
	Operation ai.Operation  `json:"operation"`
	Provider  ai.ProviderID `json:"provider"`
	API       ai.API        `json:"api"`
	ID        string        `json:"id"`
}

type FrozenSource struct {
	Kind      string `json:"kind,omitempty"`
	URL       string `json:"url"`
	FetchedAt string `json:"fetchedAt,omitempty"`
	Digest    string `json:"digest"`
	Bytes     []byte `json:"bytes"`
}

// LocalInput preserves exactly what the maintainer supplied. Unlike a fetched
// source it has a role and filename, with no invented URL or retrieval time.
type LocalInput struct {
	Role   string `json:"role"`
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Bytes  []byte `json:"bytes"`
}

func PinLocalInput(role, name string, raw []byte) LocalInput {
	return LocalInput{Role: role, Name: name, Digest: digest(raw), Bytes: append([]byte(nil), raw...)}
}

type Change struct {
	Action    string   `json:"action"`
	Identity  Identity `json:"identity"`
	Reason    string   `json:"reason"`
	Reference string   `json:"reference,omitempty"`
}

// Diagnostic separates missing data, conversion loss and explicit human
// decisions. Blocking diagnostics cannot be acknowledged away by --confirm.
type Diagnostic struct {
	Kind     string          `json:"kind"`
	Field    string          `json:"field,omitempty"`
	Message  string          `json:"message"`
	Blocking bool            `json:"blocking,omitempty"`
	Current  json.RawMessage `json:"current,omitempty"`
	Upstream json.RawMessage `json:"upstream,omitempty"`
}

type FieldChange struct {
	Identity Identity        `json:"identity"`
	Field    string          `json:"field"`
	Before   json.RawMessage `json:"before,omitempty"`
	After    json.RawMessage `json:"after,omitempty"`
}

// Candidate freezes final versions and exact runtime bytes before checking.
// It is immutable; independent check attempts never rewrite this object.
type Candidate struct {
	ID          string         `json:"id"`
	BaseDigest  string         `json:"baseDigest"`
	State       State          `json:"state"`
	Snapshot    []byte         `json:"snapshot"`
	Change      Change         `json:"change"`
	Inputs      []FrozenSource `json:"inputs,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Changes     []FieldChange  `json:"changes"`
	LocalInputs []LocalInput   `json:"localInputs,omitempty"`
}

type Revision struct {
	Adoption       *AdoptedReview `json:"adoption,omitempty"`
	ID             string         `json:"id"`
	Parent         string         `json:"parent,omitempty"`
	CandidateID    string         `json:"candidateId,omitempty"`
	State          State          `json:"state"`
	Snapshot       []byte         `json:"snapshot"`
	Change         Change         `json:"change"`
	CheckReference string         `json:"checkReference,omitempty"`
}

type ChangeRequest struct {
	Action         string
	Identity       Identity
	Reason         string
	Source         *FrozenSource
	SourceProvider string
	SourceKind     string
	SourceModel    string
	Definition     json.RawMessage
	Decisions      map[string]FieldDecision
	Revision       string
	LocalInputs    []LocalInput
}

type Proposal struct {
	Status      string        `json:"status"`
	Diagnostics []Diagnostic  `json:"diagnostics,omitempty"`
	Candidate   *Candidate    `json:"candidate,omitempty"`
	Changes     []FieldChange `json:"changes,omitempty"`
}

// Error is the stable CLI category. Messages contain maintenance diagnostics,
// never provider credentials or raw remote response bodies.
type Error struct {
	Category    string       `json:"category"`
	Message     string       `json:"message"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func failure(category, format string, args ...any) error {
	return &Error{Category: category, Message: fmt.Sprintf(format, args...)}
}

func candidateDigest(candidate Candidate) (string, error) {
	candidate.ID = ""
	raw, err := json.Marshal(candidate)
	if err != nil {
		return "", err
	}
	return digest(raw), nil
}

func validateCandidate(candidate Candidate) error {
	id, err := candidateDigest(candidate)
	if err != nil || id != candidate.ID {
		return failure("candidate_changed", "candidate digest does not match frozen content")
	}
	if strings.TrimSpace(candidate.Change.Reason) == "" {
		return failure("invalid_definition", "change reason is required")
	}
	for _, diagnostic := range candidate.Diagnostics {
		if diagnostic.Blocking {
			return failure("unresolved_definition", "candidate has unresolved %s", diagnostic.Kind)
		}
	}
	raw, err := generateCandidate(candidate.State)
	if err != nil {
		return failure("invalid_definition", "candidate state: %v", err)
	}
	if !bytes.Equal(raw, candidate.Snapshot) {
		return failure("candidate_changed", "candidate snapshot differs from frozen state")
	}
	for _, input := range candidate.Inputs {
		if digest(input.Bytes) != input.Digest || strings.TrimSpace(input.URL) == "" {
			return failure("source_changed", "frozen source is missing provenance or has changed")
		}
	}
	for _, input := range candidate.LocalInputs {
		if (input.Role != "definition" && input.Role != "decisions") || input.Name == "" || digest(input.Bytes) != input.Digest {
			return failure("input_changed", "frozen local input has invalid role, name or bytes")
		}
	}
	return nil
}
