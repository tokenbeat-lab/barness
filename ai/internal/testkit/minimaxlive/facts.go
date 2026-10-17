package minimaxlive

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
)

// RawFacts preserves absence versus explicit null/zero usage and captures
// actual model values from protocol messages, independent of catalog identity.
func RawFacts(sse string) ([]string, []json.RawMessage) {
	var models []string
	var usages []json.RawMessage
	var inspect func(map[string]json.RawMessage)
	inspect = func(obj map[string]json.RawMessage) {
		var model string
		if json.Unmarshal(obj["model"], &model) == nil && model != "" && !slices.Contains(models, model) {
			models = append(models, model)
		}
		if u, ok := obj["usage"]; ok {
			usages = append(usages, u)
		}
		for _, name := range []string{"response", "message"} {
			var sub map[string]json.RawMessage
			if json.Unmarshal(obj[name], &sub) == nil {
				inspect(sub)
			}
		}
	}
	for _, line := range strings.Split(sse, "\n") {
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var obj map[string]json.RawMessage
		if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &obj) == nil {
			inspect(obj)
		}
	}
	return models, usages
}

// JSONEqual compares evidence values after JSON whitespace normalization.
func JSONEqual(a, b any) bool {
	aa, e := json.Marshal(a)
	if e != nil {
		return false
	}
	bb, e := json.Marshal(b)
	if e != nil {
		return false
	}
	var x, y any
	if json.Unmarshal(aa, &x) != nil || json.Unmarshal(bb, &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
