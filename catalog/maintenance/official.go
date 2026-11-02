package maintenance

import (
	"encoding/json"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func selectedSourceKind(request ChangeRequest) string {
	if request.SourceKind != "" {
		return request.SourceKind
	}
	if request.Source != nil && request.Source.Kind != "" {
		return request.Source.Kind
	}
	return "models.dev"
}

func validateOfficialSource(request ChangeRequest) error {
	if request.Source == nil || len(request.Source.Bytes) == 0 || request.Source.URL == "" || digest(request.Source.Bytes) != request.Source.Digest {
		return failure("source_unavailable", "official supplement requires actual captured source bytes and URL")
	}
	if _, err := time.Parse(time.RFC3339Nano, request.Source.FetchedAt); err != nil {
		return failure("invalid_source", "official source requires its actual fetch time")
	}
	return nil
}

// Official additions carry a full manually reviewed runtime definition. The
// capture is provenance, not a second automatic mapping or protocol registry.
func officialEntry(request ChangeRequest) (Entry, []Diagnostic, error) {
	if err := validateOfficialSource(request); err != nil {
		return Entry{}, nil, err
	}
	if request.Identity.Operation == "" || request.Identity.Provider == "" || request.Identity.API == "" || request.Identity.ID == "" {
		return Entry{}, nil, failure("confirmation_required", "official addition requires explicit operation, Provider, API and wire ID")
	}
	if err := validateDefinitionJSON(request.Definition, request.Identity.Operation); err != nil {
		return Entry{}, nil, failure("invalid_definition", "official definition: %v", err)
	}
	definition := map[string]any{}
	fields := map[string]FieldDecision{}
	if err := overlayDefinition(definition, request.Definition, fields, request.Reason, request.Identity.Operation); err != nil {
		return Entry{}, nil, err
	}
	required := []string{"provider", "api", "id", "name"}
	switch request.Identity.Operation {
	case ai.OperationChat:
		required = append(required, "reasoning", "input", "contextWindow", "maxTokens", "cost")
	case ai.OperationClassifier:
		required = append(required, "contextWindow", "capabilities", "cost")
	case ai.OperationImage:
		required = append(required, "input", "output", "capabilities", "pricing")
	}
	for _, field := range required {
		if _, provided := definition[field]; !provided {
			return Entry{}, nil, failure("invalid_definition", "full official definition is missing %s", field)
		}
	}
	raw, err := json.Marshal(definition)
	if err != nil {
		return Entry{}, nil, err
	}
	entry := Entry{Operation: request.Identity.Operation, Definition: raw, Lifecycle: Active, Fields: fields, Source: &SourceLink{Kind: "official", Provider: string(request.Identity.Provider), ModelID: request.Identity.ID, Digest: request.Source.Digest, URL: request.Source.URL, FetchedAt: request.Source.FetchedAt}}
	if entryIdentity(entry) != request.Identity {
		return Entry{}, nil, failure("mapping_conflict", "official definition differs from explicitly selected identity")
	}
	for pointer, decision := range request.Decisions {
		if decision.Ownership != OwnershipManual {
			return Entry{}, nil, failure("invalid_definition", "official fields require manual ownership")
		}
		entry.Fields[pointer] = decision
	}
	for pointer, decision := range entry.Fields {
		if decision.Evidence == "" {
			decision.Evidence = "input:" + request.Source.Digest
			entry.Fields[pointer] = decision
		}
	}
	diagnostics := resolveRequiredFields(definition, entry.Operation, []Diagnostic{{Kind: "confirmation", Message: "full official definition, prices and protocol constraints require explicit manual confirmation"}}, entry.Fields)
	return entry, diagnostics, nil
}
