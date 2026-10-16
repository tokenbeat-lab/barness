package ai

import (
	"bytes"
	"encoding/json"
	"maps"
	"slices"
	"strconv"
)

// Only called after the entire native legend has matched the final frozen
// rubric. Never normalize an unvalidated echo into an apparently valid one.
func restoreTypeSafeLegend(answer *ScoreAnswer, indices map[string]int) *Error {
	if len(indices) == 0 || len(answer.Legend) == 0 {
		return nil
	}
	var walk func(any, string) any
	walk = func(value any, pointer string) any {
		switch node := value.(type) {
		case string:
			if index, ok := indices[node]; ok {
				answer.LegendImageReferences = append(answer.LegendImageReferences, ImageReference{ImageIndex: index, Placement: ImagePlacementValue, Pointer: pointer})
				return nil
			}
		case map[string]any:
			for _, key := range slices.Sorted(maps.Keys(node)) {
				node[key] = walk(node[key], pointer+"/"+imagePointerKey(key))
			}
		case []any:
			for i, item := range node {
				node[i] = walk(item, pointer+"/"+strconv.Itoa(i))
			}
		}
		return value
	}
	for _, level := range slices.Sorted(maps.Keys(answer.Legend)) {
		d := json.NewDecoder(bytes.NewReader(answer.Legend[level]))
		d.UseNumber()
		var value any
		if d.Decode(&value) != nil {
			return classifierProtocolFailure()
		}
		before := len(answer.LegendImageReferences)
		value = walk(value, "/"+strconv.Itoa(level))
		if before == len(answer.LegendImageReferences) {
			continue // Preserve image-free legend JSON exactly as received.
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return classifierProtocolFailure()
		}
		answer.Legend[level] = raw
	}
	return nil
}
