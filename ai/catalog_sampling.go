package ai

import (
	"encoding/json"
	"maps"
	"slices"
)

// All profiles are checked at the catalog boundary, including inactive ones
// and metadata on protocols that ignore sampling. Reserved fields belong to
// the consuming API; native protocols only validate the new JSON structure.
func validateThinkingSampling(profiles map[ThinkingLevel]map[string]json.RawMessage, reserved []string) string {
	for _, level := range slices.Sorted(maps.Keys(profiles)) {
		if !slices.Contains(modelThinkingLevels, level) {
			return "samplingParamsByThinkingLevel has a non-canonical key"
		}
	}
	for _, level := range modelThinkingLevels {
		params, exists := profiles[level]
		if !exists {
			continue
		}
		field := "samplingParamsByThinkingLevel." + string(level)
		if params == nil {
			return field + " must be an object"
		}
		if problem := validateJSONValues(field, params); problem != "" {
			return problem
		}
		if problem := checkReservedKeys(params, reserved); problem != "" {
			return field + ": " + problem
		}
	}
	return ""
}
