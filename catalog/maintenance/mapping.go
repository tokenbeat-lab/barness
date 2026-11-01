package maintenance

import (
	"encoding/json"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func mapNewEntry(request ChangeRequest) (Entry, []Diagnostic, error) {
	offering, canonical, err := sourceOffering(request)
	if err != nil {
		return Entry{}, nil, err
	}
	provider, ok := sourceProviders[request.SourceProvider]
	if !ok {
		return Entry{}, nil, failure("unsupported_definition", "source Provider has no reviewed barness mapping")
	}
	if request.Identity.Provider != "" && string(request.Identity.Provider) != provider {
		return Entry{}, nil, failure("mapping_conflict", "target Provider differs from the reviewed source Provider mapping")
	}
	request.Identity.Provider = ai.ProviderID(provider)
	var id, name, kind string
	if err := json.Unmarshal(offering["id"], &id); err != nil || strings.TrimSpace(id) == "" {
		return Entry{}, nil, failure("invalid_source", "offering requires its wire model ID")
	}
	if err := json.Unmarshal(offering["name"], &name); err != nil || strings.TrimSpace(name) == "" {
		return Entry{}, nil, failure("invalid_source", "offering requires a name")
	}
	if raw, exists := offering["type"]; exists {
		if err := json.Unmarshal(raw, &kind); err != nil {
			return Entry{}, nil, failure("invalid_source", "type must be a string")
		}
	}
	operation := ai.OperationChat
	switch kind {
	case "", "chat":
	case "image":
		operation = ai.OperationImage
	case "decision":
		operation = ai.OperationClassifier
	default:
		return Entry{}, nil, failure("unsupported_definition", "source type %q is not implemented", kind)
	}
	if request.Identity.Operation == "" || request.Identity.API == "" {
		return Entry{}, nil, failure("confirmation_required", "an explicit target operation and API must be confirmed")
	}
	if request.Identity.Operation != operation {
		return Entry{}, nil, failure("mapping_conflict", "source type and selected operation differ")
	}
	var modalities struct {
		Input  []string `json:"input"`
		Output []string `json:"output"`
	}
	if err := json.Unmarshal(offering["modalities"], &modalities); err != nil {
		return Entry{}, nil, failure("invalid_source", "modalities are required")
	}
	for _, modality := range append(append([]string(nil), modalities.Input...), modalities.Output...) {
		if modality != "text" && modality != "image" {
			return Entry{}, nil, failure("unsupported_definition", "unmapped modality %q", modality)
		}
	}
	definition := map[string]any{"provider": provider, "api": request.Identity.API, "id": id, "name": name, "input": modalities.Input}
	diagnostics := []Diagnostic{{Kind: "confirmation", Field: "/api", Message: "API association and barness protocol-specific constraints require explicit confirmation"}}
	if kind == "" {
		diagnostics = append(diagnostics, Diagnostic{Kind: "confirmation", Field: "/operation", Message: "missing source type creates only a chat candidate; legacy audio compatibility must be reviewed"})
	}
	fields := map[string]FieldDecision{}
	for _, key := range []string{"provider", "id", "name", "input"} {
		fields["/"+key] = FieldDecision{Ownership: OwnershipSource, Reason: "reviewed models.dev offering mapping"}
	}
	fields["/api"] = FieldDecision{Ownership: OwnershipManual, Reason: request.Reason}
	if operation == ai.OperationChat || operation == ai.OperationClassifier {
		var limits struct {
			Context int `json:"context"`
			Output  int `json:"output"`
		}
		if err := json.Unmarshal(offering["limit"], &limits); err != nil || limits.Context < 1 || (operation == ai.OperationChat && limits.Output < 1) {
			return Entry{}, nil, failure("invalid_source", "context/output limits must be positive representable integers")
		}
		definition["contextWindow"] = limits.Context
		fields["/contextWindow"] = FieldDecision{Ownership: OwnershipSource, Reason: "offering context token limit"}
		if operation == ai.OperationChat {
			definition["maxTokens"] = limits.Output
			fields["/maxTokens"] = FieldDecision{Ownership: OwnershipSource, Reason: "offering output token limit"}
			var reasoning bool
			if err := json.Unmarshal(offering["reasoning"], &reasoning); err != nil {
				return Entry{}, nil, failure("invalid_source", "reasoning must be explicitly provided")
			}
			definition["reasoning"] = reasoning
			fields["/reasoning"] = FieldDecision{Ownership: OwnershipSource, Reason: "offering reasoning declaration"}
			if reasoning {
				diagnostics = append(diagnostics, Diagnostic{Kind: "confirmation", Field: "/thinkingLevelMap", Message: "reasoning options, defaults and compat wiring are manual decisions; source null is not an unsupported target effort"})
			}
		}
		cost, priceDiagnostics := mapBaseCost(offering)
		definition["cost"] = cost
		diagnostics = append(diagnostics, priceDiagnostics...)
		for field := range cost {
			fields["/cost/"+field] = FieldDecision{Ownership: OwnershipSource, Reason: "explicit offering USD per million token base rate, confirmed by maintainer"}
		}
	} else {
		definition["output"] = modalities.Output
		diagnostics = append(diagnostics, Diagnostic{Kind: "confirmation", Field: "/pricing", Message: "image per-token/modal/per-image pricing and parameters require an explicit manual definition"})
	}
	if raw, exists := offering["status"]; exists && string(raw) == `"deprecated"` {
		diagnostics = append(diagnostics, Diagnostic{Kind: "lifecycle_hint", Field: "/status", Message: "upstream deprecated is a review hint and does not withdraw this identity"})
	}
	if raw, exists := offering["reasoning_options"]; exists {
		_ = raw
		diagnostics = append(diagnostics, Diagnostic{Kind: "lossy", Field: "/reasoning_options", Message: "source reasoning options are retained in input; target effort wiring is not inferred"})
	}
	if err := overlayDefinition(definition, request.Definition, fields, request.Reason); err != nil {
		return Entry{}, nil, err
	}
	for pointer, decision := range request.Decisions {
		fields[pointer] = decision
	}
	// Identity always belongs to the serving offering; manual overlays cannot
	// turn it into another Provider, operation, protocol or wire model.
	for key, expected := range map[string]string{"provider": provider, "api": string(request.Identity.API), "id": id} {
		if actual, ok := definition[key].(string); ok && actual != expected {
			return Entry{}, nil, failure("mapping_conflict", "manual definition changes serving identity %s", key)
		}
	}
	diagnostics = resolveRequiredFields(definition, operation, diagnostics)
	raw, err := json.Marshal(definition)
	if err != nil {
		return Entry{}, nil, err
	}
	entry := Entry{Operation: operation, Definition: raw, Lifecycle: Active, Fields: fields, Source: &SourceLink{Kind: "models.dev", Provider: request.SourceProvider, ModelID: request.SourceModel, CanonicalID: canonical, Digest: request.Source.Digest, URL: request.Source.URL, FetchedAt: request.Source.FetchedAt}}
	if entry.Source.ModelID == "" {
		entry.Source.ModelID = request.Identity.ID
	}
	return entry, diagnostics, nil
}
