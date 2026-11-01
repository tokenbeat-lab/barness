// Package maintenance owns the reviewed source of barness model definitions.
// Its provenance and decisions never enter ai's runtime model types.
package maintenance

import (
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

const (
	Active              = "active"
	Withdrawn           = "withdrawn"
	OwnershipManual     = "manual"
	OwnershipSource     = "source"
	OwnershipHistorical = "historical"
)

// State is the complete selected maintenance input. Provider versions advance
// only when an explicitly confirmed candidate is successfully applied.
type State struct {
	RulesRevision string           `json:"rulesRevision"`
	Providers     []ProviderRecord `json:"providers"`
}

type ProviderRecord struct {
	ID      ai.ProviderID `json:"id"`
	Version uint64        `json:"version"`
	Entries []Entry       `json:"entries"`
}

// Entry retains withdrawn definitions for recovery and append-only history.
// Position is the stable order within its operation's runtime collection.
type Entry struct {
	Operation  ai.Operation             `json:"operation"`
	Position   int                      `json:"position"`
	Definition json.RawMessage          `json:"definition"`
	Lifecycle  string                   `json:"lifecycle"`
	Source     *SourceLink              `json:"source,omitempty"`
	Fields     map[string]FieldDecision `json:"fields,omitempty"`
}

// SourceLink identifies retained input bytes, not an inferred upstream match.
// Path is repository/store-relative; Digest binds those original bytes or the
// explicitly identified historical catalog. MissingServing records source gaps.
type SourceLink struct {
	Kind           string `json:"kind"`
	Provider       string `json:"provider"`
	ModelID        string `json:"modelId"`
	CanonicalID    string `json:"canonicalId,omitempty"`
	Digest         string `json:"digest"`
	Path           string `json:"path,omitempty"`
	URL            string `json:"url,omitempty"`
	FetchedAt      string `json:"fetchedAt,omitempty"`
	MissingServing bool   `json:"missingServing,omitempty"`
}

// FieldDecision uses a JSON pointer to name a field in Definition. A manual
// decision remains authoritative until another explicit decision replaces it.
type FieldDecision struct {
	Ownership string `json:"ownership"`
	Reason    string `json:"reason"`
	Evidence  string `json:"evidence,omitempty"`
}
