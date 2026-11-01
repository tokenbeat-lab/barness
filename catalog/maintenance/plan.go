package maintenance

import (
	"encoding/json"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// Plan computes a deterministic preview without creating files or advancing
// versions. Proposed versions are part of the candidate tested later.
func Plan(store string, request ChangeRequest) (Proposal, error) {
	current, err := ReadCurrent(store)
	if err != nil {
		return Proposal{}, err
	}
	if strings.TrimSpace(request.Reason) == "" {
		return Proposal{}, failure("invalid_argument", "a change reason is required")
	}
	stateRaw, err := json.Marshal(current.State)
	if err != nil {
		return Proposal{}, err
	}
	state, err := DecodeState(stateRaw)
	if err != nil {
		return Proposal{}, err
	}
	diagnostics := []Diagnostic{}
	switch request.Action {
	case "withdraw", "restore":
		if err := changeLifecycle(&state, request); err != nil {
			return Proposal{}, err
		}
	case "add":
		entry, mapped, err := mapNewEntry(request)
		if err != nil {
			return Proposal{}, err
		}
		diagnostics = mapped
		for _, diagnostic := range diagnostics {
			if diagnostic.Blocking {
				return Proposal{Status: "definition_failed", Diagnostics: diagnostics}, nil
			}
		}
		if err := insertEntry(&state, entry); err != nil {
			return Proposal{}, err
		}
		request.Identity = entryIdentity(entry)
	case "revise":
		var err error
		diagnostics, err = reviseEntry(&state, request)
		if err != nil {
			return Proposal{}, err
		}
	default:
		return Proposal{}, failure("invalid_argument", "unknown maintenance action %q", request.Action)
	}
	changed, err := advanceVersions(current.State, &state)
	if err != nil {
		return Proposal{}, failure("invalid_definition", "%v", err)
	}
	nextDigest, err := stateDigest(state)
	if err != nil {
		return Proposal{}, err
	}
	if !changed && nextDigest == current.ID {
		return Proposal{Status: "unchanged"}, nil
	}
	snapshot, err := Generate(state)
	if err != nil {
		return Proposal{}, failure("invalid_definition", "%v", err)
	}
	candidate := Candidate{BaseDigest: current.ID, State: state, Snapshot: snapshot, Change: Change{request.Action, request.Identity, request.Reason}, Diagnostics: diagnostics}
	candidate.Changes = stateChanges(current.State, state)
	if request.Source != nil {
		candidate.Inputs = []FrozenSource{*request.Source}
	}
	candidate.ID, err = candidateDigest(candidate)
	if err != nil {
		return Proposal{}, err
	}
	return Proposal{Status: "unconfirmed", Candidate: &candidate, Diagnostics: diagnostics, Changes: candidate.Changes}, nil
}

func insertEntry(state *State, entry Entry) error {
	identity := entryIdentity(entry)
	position := 0
	providerIndex := -1
	for pi, provider := range state.Providers {
		if provider.ID == identity.Provider {
			providerIndex = pi
		}
		for _, existing := range provider.Entries {
			if entryIdentity(existing) == identity {
				return failure("identity_exists", "add cannot overwrite an existing identity; use an explicitly reviewed revise or restore")
			}
			if existing.Operation == entry.Operation && existing.Position >= position {
				position = existing.Position + 1
			}
		}
	}
	entry.Position = position
	if providerIndex < 0 {
		state.Providers = append(state.Providers, ProviderRecord{ID: identity.Provider, Version: 1})
		providerIndex = len(state.Providers) - 1
	}
	state.Providers[providerIndex].Entries = append(state.Providers[providerIndex].Entries, entry)
	return nil
}

func changeLifecycle(state *State, request ChangeRequest) error {
	for pi := range state.Providers {
		for ei := range state.Providers[pi].Entries {
			entry := &state.Providers[pi].Entries[ei]
			if entryIdentity(*entry) != request.Identity {
				continue
			}
			entry.Lifecycle = Withdrawn
			if request.Action == "restore" {
				entry.Lifecycle = Active
			}
			return nil
		}
	}
	return failure("not_found", "model identity is not maintained")
}

func entryIdentity(entry Entry) Identity {
	var identity Identity
	_ = json.Unmarshal(entry.Definition, &identity)
	identity.Operation = entry.Operation
	return identity
}

func advanceVersions(base State, next *State) (bool, error) {
	before := map[ai.ProviderID]ProviderRecord{}
	for _, provider := range base.Providers {
		before[provider.ID] = provider
	}
	changed := false
	for i := range next.Providers {
		provider := &next.Providers[i]
		old, exists := before[provider.ID]
		if !exists {
			provider.Version = 1
			changed = true
			continue
		}
		a, err := ProviderHash(old)
		if err != nil {
			return false, err
		}
		b, err := ProviderHash(*provider)
		if err != nil {
			return false, err
		}
		provider.Version = old.Version
		if a != b {
			provider.Version++
			changed = true
		}
	}
	return changed, nil
}
