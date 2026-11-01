package maintenance

import (
	"encoding/json"
	"strings"
)

func reviseEntry(state *State, request ChangeRequest) ([]Diagnostic, error) {
	if len(request.Definition) == 0 && len(request.Decisions) == 0 && request.Source == nil {
		return nil, failure("invalid_argument", "revise requires an explicit definition, field decisions or source association")
	}
	for pi := range state.Providers {
		for ei := range state.Providers[pi].Entries {
			entry := &state.Providers[pi].Entries[ei]
			if entryIdentity(*entry) != request.Identity {
				continue
			}
			var definition map[string]any
			if err := json.Unmarshal(entry.Definition, &definition); err != nil {
				return nil, err
			}
			if entry.Fields == nil {
				entry.Fields = map[string]FieldDecision{}
			}
			if err := overlayDefinition(definition, request.Definition, entry.Fields, request.Reason); err != nil {
				return nil, err
			}
			for pointer, decision := range request.Decisions {
				if !hasField(definition, pointer) {
					return nil, failure("invalid_definition", "field decision %s does not identify a defined field", pointer)
				}
				if decision.Ownership == OwnershipSource && request.Source == nil && (entry.Source == nil || entry.Source.Kind != "models.dev") {
					return nil, failure("confirmation_required", "source ownership requires an explicit pinned serving association")
				}
				entry.Fields[pointer] = decision
			}
			raw, err := json.Marshal(definition)
			if err != nil {
				return nil, err
			}
			before := entryIdentity(*entry)
			entry.Definition = raw
			if entryIdentity(*entry) != before {
				return nil, failure("mapping_conflict", "revise cannot change the complete serving identity")
			}
			if request.Source != nil {
				offering, canonical, err := sourceOffering(request)
				if err != nil {
					return nil, err
				}
				var sourceID string
				if err := json.Unmarshal(offering["id"], &sourceID); err != nil || sourceID != before.ID || sourceProviders[request.SourceProvider] != string(before.Provider) {
					return nil, failure("mapping_conflict", "source association differs from the maintained serving identity")
				}
				model := request.SourceModel
				if model == "" {
					model = request.Identity.ID
				}
				entry.Source = &SourceLink{Kind: "models.dev", Provider: request.SourceProvider, ModelID: model, CanonicalID: canonical, Digest: request.Source.Digest, URL: request.Source.URL, FetchedAt: request.Source.FetchedAt}
			}
			return []Diagnostic{{Kind: "confirmation", Message: "existing human decisions are replaced only at explicitly supplied fields"}}, nil
		}
	}
	return nil, failure("not_found", "model identity is not maintained")
}

func hasField(definition map[string]any, pointer string) bool {
	if pointer == "/" {
		return true
	}
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	var value any = definition
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		object, ok := value.(map[string]any)
		if !ok {
			return false
		}
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		value, ok = object[part]
		if !ok {
			return false
		}
	}
	return true
}
