package ai

import (
	"cmp"
	"encoding/json"
	"maps"
	"slices"
)

// Sampling follows the supported canonical level, independently of the
// adapter's wire reasoning. Apply it once after named fields, before Hooks;
// reasoning overrides in sampling/callbacks never feed back into selection.
func applyModelSamplingParams(body []byte, model Model, input ThinkingLevel, overrides map[string]json.RawMessage) ([]byte, error) {
	level := clampThinkingLevel(model, cmp.Or(input, thinkingOff))
	return applySamplingParams(body, model.SamplingParams, model.SamplingParamsByThinkingLevel[level], overrides)
}

// applySamplingParams sets each sampling parameter on the encoded body,
// replacing a named field of the same name (pi's Object.assign after
// buildParams). Keys are applied in sorted order; the body's own key order
// is kept and new keys are appended.
func applySamplingParams(body []byte, sources ...map[string]json.RawMessage) ([]byte, error) {
	count := 0
	for _, params := range sources {
		count += len(params)
	}
	if count == 0 {
		return body, nil
	}
	v, err := parseJSON(string(body))
	if err != nil {
		return nil, err
	}
	obj := v.(*jsonObject)
	for _, params := range sources {
		for _, k := range slices.Sorted(maps.Keys(params)) {
			value, err := parseJSON(string(params[k]))
			if err != nil {
				return nil, err
			}
			obj.set(k, value)
		}
	}
	return []byte(stringifyJSON(obj)), nil
}
