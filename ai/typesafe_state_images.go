package ai

import (
	"encoding/json"
	"slices"
	"strconv"
)

// Reserve the complete projected State before expanding any occurrences. The
// source JSON is already compact and HTML escaped; every target is a unique
// null checked at the public boundary. Numbers remain json.Number throughout.
func (c *Client) projectTypeSafeState(raw json.RawMessage, state any, plan classifierImagePlan, keys []string) (json.RawMessage, *Error) {
	limit := c.policy.Classifier.MaxStateBytes
	remaining := limit - int64(len(raw))
	contextCount := 0
	keySizes := make([]int64, len(keys))
	for i, key := range keys {
		encoded, _ := json.Marshal(key)
		keySizes[i] = int64(len(encoded))
	}
	for _, ref := range plan {
		size := keySizes[ref.ImageIndex]
		if ref.Placement == ImagePlacementValue {
			size -= 4 // replace only the literal null
		} else {
			if contextCount == 0 {
				size += int64(len(`{"state":,"images":[]}`))
			} else {
				size++
			}
			contextCount++
		}
		if size > remaining {
			return nil, limitFailure(PhaseScope, "Classifier.MaxStateBytes", limit)
		}
		remaining -= size
	}
	contextKeys := make([]string, 0, contextCount)
	for _, ref := range plan {
		key := keys[ref.ImageIndex]
		if ref.Placement == ImagePlacementContext {
			contextKeys = append(contextKeys, key)
			continue
		}
		parts, _ := imagePointerParts(ref.Pointer)
		if len(parts) == 1 {
			state = key
			continue
		}
		parent := state
		for _, part := range parts[1 : len(parts)-1] {
			switch node := parent.(type) {
			case map[string]any:
				parent = node[part]
			case []any:
				i, _ := strconv.Atoi(part)
				parent = node[i]
			}
		}
		last := parts[len(parts)-1]
		switch node := parent.(type) {
		case map[string]any:
			node[last] = key
		case []any:
			i, _ := strconv.Atoi(last)
			node[i] = key
		}
	}
	if contextCount > 0 {
		state = struct {
			State  any      `json:"state"`
			Images []string `json:"images"`
		}{state, contextKeys}
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return nil, typeSafeRequestFailure()
	}
	return encoded, nil
}

// Keep envelope provenance outside native user data: a business object named
// state/images is still business content. Hooks may change every value and asset,
// but cannot turn those coincidental field names into an implicit context scope.
func typeSafeContextProjection(req ClassifierRequest, final typeSafeRequest, locations []ClassifierImageLocation) bool {
	// With no builder layout, State-capable models scan all final State values
	// as business content. Only context-only models need envelope recognition;
	// their images must be actual whole-value asset uses, never template text.
	if len(req.Images) == 0 {
		if slices.Contains(locations, ClassifierImageLocationState) {
			return false
		}
		var envelope typeSafeContext
		if decodeClassifierJSON(final.State, &envelope) != nil || len(envelope.Images) == 0 {
			return false
		}
		for _, key := range envelope.Images {
			if _, exists := final.Assets[key]; !exists {
				return false
			}
		}
		return true
	}
	if len(req.ImageReferences) == 0 {
		return true
	}
	for _, ref := range req.ImageReferences {
		if ref.Placement == ImagePlacementContext {
			return true
		}
	}
	return false
}
