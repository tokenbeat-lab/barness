package ai

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Resolve only content values. Protocol fields and collection containers never
// enter this traversal. No path is created and borrowed JSON is never modified.
func classifierReferenceTarget(req ClassifierRequest, pointer string) bool {
	parts, ok := imagePointerParts(pointer)
	if !ok {
		return false
	}
	var raw json.RawMessage
	var tail []string
	switch parts[0] {
	case "state":
		raw, tail = req.State, parts[1:]
	case "questions":
		if len(parts) < 3 {
			return false
		}
		q, exists := req.Questions[parts[1]]
		if !exists {
			return false
		}
		if parts[2] == "instructions" {
			switch q := q.(type) {
			case ChoiceQuestion:
				raw = q.Instructions
			case ScoreQuestion:
				raw = q.Instructions
			case BoxQuestion:
				raw = q.Instructions
			case BoolQuestion:
				raw = q.Instructions
			}
			tail = parts[3:]
		} else if parts[2] == "criteria" && len(parts) >= 4 {
			switch q := q.(type) {
			case ChoiceQuestion:
				raw = q.Criteria[parts[3]]
			case ScoreQuestion:
				i, ok := imageArrayIndex(parts[3], len(q.Criteria))
				if !ok {
					return false
				}
				raw = q.Criteria[i]
			case BoolQuestion:
				if q.Criteria == nil {
					return false
				}
				switch parts[3] {
				case "true":
					raw = q.Criteria.True
				case "false":
					raw = q.Criteria.False
				}
			}
			tail = parts[4:]
		} else {
			return false
		}
	default:
		return false
	}
	for _, part := range tail {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			return false
		}
		switch raw[0] {
		case '{':
			var values map[string]json.RawMessage
			if json.Unmarshal(raw, &values) != nil {
				return false
			}
			raw = values[part]
		case '[':
			var values []json.RawMessage
			if json.Unmarshal(raw, &values) != nil {
				return false
			}
			i, ok := imageArrayIndex(part, len(values))
			if !ok {
				return false
			}
			raw = values[i]
		default:
			return false
		}
	}
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func imagePointerParts(pointer string) ([]string, bool) {
	if !strings.HasPrefix(pointer, "/") {
		return nil, false
	}
	parts := strings.Split(pointer[1:], "/")
	for i, part := range parts {
		for j := 0; j < len(part); j++ {
			if part[j] == '~' {
				j++
				if j >= len(part) || (part[j] != '0' && part[j] != '1') {
					return nil, false
				}
			}
		}
		// RFC 6901 requires ~1 before ~0; ~01 denotes the literal key ~1.
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
	}
	return parts, true
}

func imageArrayIndex(part string, length int) (int, bool) {
	i, err := strconv.Atoi(part)
	return i, err == nil && i >= 0 && i < length && strconv.Itoa(i) == part
}

func imagePointerKey(key string) string {
	return strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
}

func (p classifierImagePlan) hasValue(pointer string) bool {
	for _, ref := range p {
		if ref.Placement == ImagePlacementValue && ref.Pointer == pointer {
			return true
		}
	}
	return false
}

// Document validation bounds and checks the JSON before traversal can allocate
// any subtrees. Roots marked by the plan are allowed to be null during that
// validation; they must still resolve to an allowed target here.
func (p classifierImagePlan) checkValues(req ClassifierRequest) *Error {
	targets := map[string]bool{}
	for _, ref := range p {
		if ref.Placement != ImagePlacementValue {
			continue
		}
		if targets[ref.Pointer] || !classifierReferenceTarget(req, ref.Pointer) {
			return newError(CodeInvalidRequest, PhaseScope, "image reference must target a unique existing null content value")
		}
		targets[ref.Pointer] = true
	}
	return nil
}
