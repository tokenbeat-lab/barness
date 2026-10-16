package ai

import "bytes"

// Bound projected text without copying it. JSON string decoding and document
// compaction can shrink raw input, so raw byte lengths cannot be combined with
// native image expansion. Framing/escaping can only add to this lower bound.
func classifierJSONTextLowerBound(raw []byte) int64 {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0
	}
	size := int64(0)
	if raw[0] == '"' {
		for i := 1; i < len(raw)-1; i++ {
			size++
			if raw[i] == '\\' && i+1 < len(raw)-1 {
				i++
				if raw[i] == 'u' && i+4 < len(raw)-1 {
					i += 4
				}
			}
		}
		// A Unicode escape contributes at least one byte; paired surrogate escapes
		// contribute four, so counting each as one remains conservative.
		return size
	}
	if raw[0] != '{' && raw[0] != '[' {
		return 0
	}
	quoted, escaped := false, false
	for _, ch := range raw {
		if quoted {
			size++
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				quoted = false
			}
		} else {
			switch ch {
			case ' ', '\t', '\r', '\n':
				continue
			case '"':
				quoted = true
			}
			size++
		}
	}
	return size
}

// This traversal deliberately differs from the raw source budget: each JSON
// description may shrink when projected to native text. Keep literal source
// accounting in classifierQuestionSize so it never inherits that reduction.
func classifierQuestionTextLowerBound(question ClassifierQuestion, limit int64) (int64, bool) {
	size := int64(0)
	take := func(raw []byte) bool {
		n := classifierJSONTextLowerBound(raw)
		if n > limit-size {
			return false
		}
		size += n
		return true
	}
	switch q := question.(type) {
	case ChoiceQuestion:
		if !take(q.Instructions) {
			return size, false
		}
		for key, value := range q.Criteria {
			if int64(len(key)) > limit-size {
				return size, false
			}
			size += int64(len(key))
			if !take(value) {
				return size, false
			}
		}
	case ScoreQuestion:
		if !take(q.Instructions) {
			return size, false
		}
		for _, value := range q.Criteria {
			if !take(value) {
				return size, false
			}
		}
	case BoxQuestion:
		if !take(q.Instructions) {
			return size, false
		}
	case BoolQuestion:
		if !take(q.Instructions) {
			return size, false
		}
		if q.Criteria != nil && (!take(q.Criteria.True) || !take(q.Criteria.False)) {
			return size, false
		}
	}
	return size, true
}
