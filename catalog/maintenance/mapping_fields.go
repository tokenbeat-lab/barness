package maintenance

import (
	"bytes"
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func mapBaseCost(offering map[string]json.RawMessage) (map[string]any, []Diagnostic) {
	cost := map[string]any{}
	diagnostics := []Diagnostic{}
	var source map[string]json.RawMessage
	_ = json.Unmarshal(offering["cost"], &source)
	for _, pair := range [][2]string{{"input", "input"}, {"output", "output"}, {"cache_read", "cacheRead"}, {"cache_write", "cacheWrite"}} {
		sourceKey, targetKey := pair[0], pair[1]
		raw, exists := source[sourceKey]
		if !exists || bytes.Equal(raw, []byte("null")) {
			continue
		}
		var rate float64
		if err := json.Unmarshal(raw, &rate); err != nil || rate < 0 {
			diagnostics = append(diagnostics, Diagnostic{Kind: "invalid_definition", Field: "/cost/" + targetKey, Message: "source rate must be a nonnegative number", Blocking: true})
			continue
		}
		cost[targetKey] = rate
	}
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	tierSeen := false
	for _, key := range keys {
		switch key {
		case "input", "output", "cache_read", "cache_write":
		case "tiers", "context_over_200k":
			if !tierSeen {
				diagnostics = append(diagnostics, Diagnostic{Kind: "lossy", Field: "/cost/tiers", Message: "projected context size loses gt/gte; define complete target tiers and explicitly confirm strict greater-than boundaries", Blocking: true})
				tierSeen = true
			}
		default:
			diagnostics = append(diagnostics, Diagnostic{Kind: "unsupported_definition", Field: "/cost/" + key, Message: "conditional or modality rate cannot be silently collapsed into base token prices", Blocking: true})
		}
	}
	return cost, diagnostics
}

func overlayDefinition(target map[string]any, raw json.RawMessage, fields map[string]FieldDecision, reason string, operation ai.Operation) error {
	if len(raw) == 0 {
		return nil
	}
	if err := validateDefinitionJSON(raw, operation); err != nil {
		return failure("invalid_definition", "manual definition: %v", err)
	}
	var overlay map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&overlay); err != nil || overlay == nil {
		return failure("invalid_definition", "manual definition must be an object")
	}
	mergeDefinition(target, overlay, "", fields, reason)
	return nil
}

func mergeDefinition(target, overlay map[string]any, prefix string, fields map[string]FieldDecision, reason string) {
	for key, value := range overlay {
		pointer := prefix + "/" + strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
		if object, ok := value.(map[string]any); ok {
			current, ok := target[key].(map[string]any)
			if !ok {
				current = map[string]any{}
			}
			mergeDefinition(current, object, pointer, fields, reason)
			target[key] = current
			if len(object) == 0 {
				fields[pointer] = FieldDecision{Ownership: OwnershipManual, Reason: reason}
			}
		} else {
			target[key] = value
			fields[pointer] = FieldDecision{Ownership: OwnershipManual, Reason: reason}
		}
	}
}

func resolveRequiredFields(definition map[string]any, operation ai.Operation, diagnostics []Diagnostic, fields map[string]FieldDecision) []Diagnostic {
	for i := range diagnostics {
		diagnostic := &diagnostics[i]
		if diagnostic.Kind == "lossy" && diagnostic.Field == "/cost/tiers" && fields["/cost/tiers"].Ownership == OwnershipManual {
			cost, _ := definition["cost"].(map[string]any)
			if tiers, ok := cost["tiers"].([]any); ok && len(tiers) > 0 {
				diagnostic.Blocking = false
			}
		}
		if diagnostic.Kind == "invalid_definition" && strings.HasPrefix(diagnostic.Field, "/cost/") && fields[diagnostic.Field].Ownership == OwnershipManual {
			diagnostic.Blocking = false
		}
	}
	diagnostics = append(diagnostics, requiredCostFields(definition, operation)...)
	if operation == ai.OperationChat {
		if reasoning, _ := definition["reasoning"].(bool); reasoning {
			if _, provided := definition["thinkingLevelMap"]; !provided {
				diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: "/thinkingLevelMap", Message: "reasoning requires reviewed target effort wiring", Blocking: true})
			}
		}
	}
	if operation == ai.OperationImage {
		for _, field := range []string{"capabilities", "pricing"} {
			if _, provided := definition[field]; !provided {
				diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: "/" + field, Message: "image definition requires explicit confirmed " + field, Blocking: true})
			}
		}
	}
	if operation == ai.OperationClassifier {
		if _, provided := definition["capabilities"]; !provided {
			diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: "/capabilities", Message: "classifier capabilities require explicit confirmation", Blocking: true})
		}
	}
	return diagnostics
}

// Rate presence is shared by additions and revisions. Add-only capability and
// reasoning wiring policy must not be reimposed on accepted legacy defaults.
func requiredCostFields(definition map[string]any, operation ai.Operation) []Diagnostic {
	diagnostics := []Diagnostic{}
	if operation == ai.OperationChat || operation == ai.OperationClassifier {
		cost, _ := definition["cost"].(map[string]any)
		diagnostics = append(diagnostics, missingTierFields(cost)...)
		for _, field := range []string{"input", "output", "cacheRead", "cacheWrite"} {
			value, provided := cost[field]
			if !provided || value == nil {
				diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: "/cost/" + field, Message: "rate is unknown; provide an explicit confirmed value (including zero)", Blocking: true})
			}
		}
	}
	return diagnostics
}

// Every target tier, including manual/official additions, needs explicit
// threshold and complete rates. Typed zero defaults cannot prove presence.
func missingTierFields(cost map[string]any) []Diagnostic {
	tiers, _ := cost["tiers"].([]any)
	diagnostics := []Diagnostic{}
	for index, tier := range tiers {
		prefix := "/cost/tiers/" + strconv.Itoa(index)
		values, ok := tier.(map[string]any)
		if !ok {
			diagnostics = append(diagnostics, Diagnostic{Kind: "invalid_definition", Field: prefix, Message: "target tier must be an object", Blocking: true})
			continue
		}
		for _, field := range []string{"inputTokensAbove", "input", "output", "cacheRead", "cacheWrite"} {
			if value, provided := values[field]; !provided || value == nil {
				diagnostics = append(diagnostics, Diagnostic{Kind: "missing", Field: prefix + "/" + field, Message: "each target tier requires its explicit threshold and all four rates", Blocking: true})
			}
		}
	}
	return diagnostics
}
