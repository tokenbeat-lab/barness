package maintenance

import (
	"encoding/json"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// updateCosts only refreshes explicitly source-owned base token rates. Price
// forms whose units/conditions cannot be matched remain unchanged for review.
func updateCosts(state *State, request ChangeRequest) ([]Diagnostic, error) {
	source, err := sourceDocument(request)
	if err != nil {
		return nil, err
	}
	target, ok := sourceProviders[request.SourceProvider]
	if !ok || (request.Identity.Provider != "" && string(request.Identity.Provider) != target) {
		return nil, failure("mapping_conflict", "cost refresh requires the matching reviewed Provider association")
	}
	diagnostics := []Diagnostic{}
	for pi := range state.Providers {
		if string(state.Providers[pi].ID) != target {
			continue
		}
		for ei := range state.Providers[pi].Entries {
			entry := &state.Providers[pi].Entries[ei]
			identity := entryIdentity(*entry)
			if entry.Lifecycle != Active || !matchesScope(identity, request.Identity) {
				continue
			}
			if entry.Source == nil || entry.Source.Kind != "models.dev" || entry.Source.Provider != request.SourceProvider {
				continue
			}
			if entry.Operation == ai.OperationImage {
				diagnostics = append(diagnostics, Diagnostic{Kind: "manual_confirmation", Field: identityKey(identity) + "/pricing", Message: "image token/modal/per-image prices require an explicitly reviewed compatible pricing definition; kept unchanged"})
				continue
			}
			lookup := request
			lookup.SourceModel = entry.Source.ModelID
			offering, _, err := offeringFrom(source, lookup)
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: identityKey(identity), Message: "offering missing: retained existing definition and lifecycle"})
				continue
			}
			var sourceID string
			if json.Unmarshal(offering["id"], &sourceID) != nil || sourceID != identity.ID {
				return nil, failure("mapping_conflict", "source wire identity differs from maintained association")
			}
			rates, conditions := mapBaseCost(offering)
			if len(conditions) > 0 {
				for _, condition := range conditions {
					condition.Blocking = false
					condition.Message += "; existing model prices kept unchanged"
					diagnostics = append(diagnostics, condition)
				}
				continue
			}
			var definition map[string]any
			if err := json.Unmarshal(entry.Definition, &definition); err != nil {
				return nil, err
			}
			cost, _ := definition["cost"].(map[string]any)
			changed := false
			for _, field := range []string{"input", "output", "cacheRead", "cacheWrite"} {
				pointer := "/cost/" + field
				value, provided := rates[field]
				if !provided {
					diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: identityKey(identity) + pointer, Message: "source rate absent: retained existing value without substituting zero"})
					continue
				}
				old, _ := json.Marshal(cost[field])
				next, _ := json.Marshal(value)
				if string(old) == string(next) {
					continue
				}
				if fieldOwnership(entry.Fields, pointer) != OwnershipSource {
					diagnostics = append(diagnostics, Diagnostic{Kind: "manual_override", Field: identityKey(identity) + pointer, Message: "upstream price differs; retained explicit manual/historical field", Current: old, Upstream: next})
					continue
				}
				cost[field] = value
				changed = true
			}
			if changed {
				entry.Definition, err = json.Marshal(definition)
				if err != nil {
					return nil, err
				}
				entry.Source.Digest, entry.Source.URL, entry.Source.FetchedAt = request.Source.Digest, request.Source.URL, request.Source.FetchedAt
			}
			if string(offering["status"]) == `"deprecated"` {
				diagnostics = append(diagnostics, Diagnostic{Kind: "lifecycle_hint", Field: identityKey(identity), Message: "upstream deprecated is a review hint; model remains active"})
			}
		}
	}
	return diagnostics, nil
}

func fieldOwnership(fields map[string]FieldDecision, pointer string) string {
	for {
		if decision, exists := fields[pointer]; exists {
			return decision.Ownership
		}
		index := strings.LastIndex(pointer, "/")
		if index <= 0 {
			break
		}
		pointer = pointer[:index]
	}
	if decision, exists := fields["/"]; exists {
		return decision.Ownership
	}
	return OwnershipManual
}

func matchesScope(identity, scope Identity) bool {
	return (scope.Operation == "" || scope.Operation == identity.Operation) && (scope.Provider == "" || scope.Provider == identity.Provider) && (scope.API == "" || scope.API == identity.API) && (scope.ID == "" || scope.ID == identity.ID)
}
