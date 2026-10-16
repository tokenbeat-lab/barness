package ai

import "encoding/json"

// Reserve complete canonical questions, including escaped private keys, before
// expanding their content slots. Text container formats stay independent
// of the image-location capability; only explicit root images bypass them.
func (c *Client) projectTypeSafeQuestions(native *typeSafeRequest, original map[string]any, plan classifierImagePlan, keys []string) *Error {
	remaining := make(map[string]int64)
	for _, ref := range plan {
		if ref.Placement != ImagePlacementValue {
			continue
		}
		parts, _ := imagePointerParts(ref.Pointer)
		if parts[0] != "questions" {
			continue
		}
		name := parts[1]
		budget, exists := remaining[name]
		if !exists {
			raw, err := json.Marshal(native.Questions[name])
			if err != nil {
				return typeSafeRequestFailure()
			}
			budget = c.policy.Classifier.MaxQuestionBytes - int64(len(raw))
		}
		encoded, _ := json.Marshal(keys[ref.ImageIndex])
		growth := int64(len(encoded)) - 4
		if growth > budget {
			return limitFailure(PhaseScope, "Classifier.MaxQuestionBytes", c.policy.Classifier.MaxQuestionBytes)
		}
		remaining[name] = budget - growth
	}
	questions := original["questions"].(map[string]any)
	for _, ref := range plan {
		if ref.Placement != ImagePlacementValue {
			continue
		}
		parts, _ := imagePointerParts(ref.Pointer)
		if parts[0] == "questions" {
			q := questions[parts[1]].(map[string]any)
			q[parts[2]] = projectTypeSafeValue(q[parts[2]], parts[3:], keys[ref.ImageIndex])
		}
	}
	for name := range remaining {
		raw, err := json.Marshal(questions[name])
		if err != nil {
			return typeSafeRequestFailure()
		}
		var q typeSafeQuestion
		if json.Unmarshal(raw, &q) != nil {
			return typeSafeRequestFailure()
		}
		native.Questions[name] = q
	}
	return nil
}

func typeSafeRootImage(raw json.RawMessage, assets map[string]typeSafeAsset) bool {
	var key string
	if json.Unmarshal(raw, &key) != nil {
		return false
	}
	_, exists := assets[key]
	return exists
}
