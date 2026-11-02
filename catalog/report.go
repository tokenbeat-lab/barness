// Package catalog loads local model catalogs for Host assembly. It performs no
// network access and never merges a selected file with the built-in catalog.
package catalog

import (
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

// Status describes acceptance by this program, not vendor health or permission.
type Status string

const (
	Loaded Status = "loaded"
	Failed Status = "failed"
)

// Reason classifies a model failure without converting it into a runtime error.
type Reason string

const (
	UnknownField          Reason = "unknown_field"
	UnsupportedDefinition Reason = "unsupported_definition"
	InvalidDefinition     Reason = "invalid_definition"
	DuplicateIdentity     Reason = "duplicate_identity"
)

// Identity can be partial when the source has missing or ill-typed fields.
type Identity struct {
	Operation ai.Operation  `json:"operation"`
	Provider  ai.ProviderID `json:"provider,omitempty"`
	API       ai.API        `json:"api,omitempty"`
	ID        string        `json:"id,omitempty"`
}

// ModelStatus is one source position. Raw retains rejected definitions for a
// trusted Host to diagnose; it must not be copied automatically to public logs.
type ModelStatus struct {
	Identity   Identity        `json:"identity"`
	Collection string          `json:"collection"`
	Index      int             `json:"index"`
	Status     Status          `json:"status"`
	Reason     Reason          `json:"reason,omitempty"`
	Diagnostic string          `json:"diagnostic,omitempty"`
	Suggestion string          `json:"suggestion,omitempty"`
	Raw        json.RawMessage `json:"raw,omitempty"`
}

// Report is separate from the effective runtime catalog and owns its raw bytes.
// SourceHash signs the complete original catalog. EffectiveHash signs the typed
// accepted subset, including SourceVersion, and can legitimately differ.
type Report struct {
	SourceVersion string        `json:"sourceVersion"`
	SourceHash    string        `json:"sourceHash"`
	EffectiveHash string        `json:"effectiveHash"`
	Loaded        int           `json:"loaded"`
	Failed        int           `json:"failed"`
	Models        []ModelStatus `json:"models"`
}

// FileReason names failures that reject the entire file before model loading.
type FileReason string

const (
	FileReadFailed       FileReason = "read_failed"
	FileInvalidJSON      FileReason = "invalid_json"
	FileInvalidStructure FileReason = "invalid_structure"
	FileInvalidVersion   FileReason = "invalid_version"
	FileInvalidHash      FileReason = "invalid_hash"
	FileHashMismatch     FileReason = "hash_mismatch"
)

// FileError always rejects the whole source; no partial catalog is returned.
type FileError struct {
	Reason     FileReason `json:"reason"`
	Diagnostic string     `json:"diagnostic"`
}

func (e *FileError) Error() string { return "catalog: " + string(e.Reason) + ": " + e.Diagnostic }
