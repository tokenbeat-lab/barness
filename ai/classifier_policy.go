package ai

import (
	"bytes"
	"encoding/json"
	"maps"
	"slices"
)

// ClassifierPolicy enables classification with explicit finite input bounds.
type ClassifierPolicy struct {
	MaxQuestions     int
	MaxStateBytes    int64
	MaxQuestionBytes int64
}

func (p *ClassifierPolicy) validate() error {
	if p == nil {
		return nil
	}
	for _, v := range []struct {
		name  string
		value int64
	}{{"MaxQuestions", int64(p.MaxQuestions)}, {"MaxStateBytes", p.MaxStateBytes}, {"MaxQuestionBytes", p.MaxQuestionBytes}} {
		if v.value <= 0 {
			return policyError("Classifier."+v.name, "must be positive; unlimited is not supported")
		}
	}
	return nil
}
func (p *ResourcePolicy) clone() ResourcePolicy {
	out := *p
	if p.Classifier != nil {
		v := *p.Classifier
		out.Classifier = &v
	}
	return out
}

func (p *ClassifierPolicy) check(req ClassifierRequest) *Error {
	if p == nil {
		return newError(CodeInvalidRequest, PhaseScope, "the client policy does not enable classification")
	}
	if len(req.Questions) > p.MaxQuestions {
		return limitFailure(PhaseScope, "Classifier.MaxQuestions", int64(p.MaxQuestions))
	}
	if int64(len(req.State)) > p.MaxStateBytes {
		return limitFailure(PhaseScope, "Classifier.MaxStateBytes", p.MaxStateBytes)
	}
	if !classifierState(req.State) || uniqueClassifierJSON(req.State) != nil {
		return newError(CodeInvalidRequest, PhaseScope, "classifier state must be a JSON string, object or array")
	}
	if len(req.Questions) == 0 {
		return newError(CodeInvalidRequest, PhaseScope, "classifier requires at least one question")
	}
	for _, key := range slices.Sorted(maps.Keys(req.Questions)) {
		if key == "" {
			return newError(CodeInvalidRequest, PhaseScope, "classifier question key must not be empty")
		}
		q := req.Questions[key]
		if size, ok := classifierQuestionSize(q, p.MaxQuestionBytes); !ok || size > p.MaxQuestionBytes {
			return limitFailure(PhaseScope, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
		}
		if problem := validateClassifierQuestion(q); problem != "" {
			return newError(CodeInvalidRequest, PhaseScope, problem)
		}
		// Count the complete encoded question too, including JSON key escaping
		// and punctuation, after the cheap check bounded the allocation.
		body, err := json.Marshal(q)
		if err != nil {
			return newError(CodeInvalidRequest, PhaseScope, "classifier question could not be encoded")
		}
		if int64(len(body)) > p.MaxQuestionBytes {
			return limitFailure(PhaseScope, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
		}
	}
	return nil
}

func classifierState(raw json.RawMessage) bool {
	if !json.Valid(raw) {
		return false
	}
	for _, b := range raw {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '"', '{', '[':
			return true
		default:
			return false
		}
	}
	return false
}

func (r ClassifierRequest) checkCapabilities(c ClassifierCapabilities) string {
	for _, question := range r.Questions {
		var kind ClassifierQuestionKind
		switch q := question.(type) {
		case ChoiceQuestion:
			kind = ClassifierQuestionChoice
			if len(q.Criteria) > c.MaxChoices {
				return "choice options exceed model capability"
			}
		case ScoreQuestion:
			kind = ClassifierQuestionScore
			if len(q.Criteria) < c.MinScoreLevels || len(q.Criteria) > c.MaxScoreLevels {
				return "score levels exceed model capability"
			}
		case BoolQuestion:
			kind = ClassifierQuestionBool
		default:
			return "unsupported classifier question"
		}
		if !slices.Contains(c.Kinds, kind) {
			return "model does not support the question kind"
		}
	}
	return ""
}

// classifierQuestionSize sums raw lengths without encoding or copying the
// caller's data. Both per-question and whole-request limits use this bound.
func classifierQuestionSize(question ClassifierQuestion, limit int64) (int64, bool) {
	size := int64(0)
	take := func(n int) bool {
		if int64(n) > limit-size {
			return false
		}
		size += int64(n)
		return true
	}
	switch q := question.(type) {
	case ChoiceQuestion:
		if !take(len(q.Instructions)) {
			return size, false
		}
		for key, value := range q.Criteria {
			if !take(len(key)) || !take(len(value)) {
				return size, false
			}
		}
	case ScoreQuestion:
		if !take(len(q.Instructions)) {
			return size, false
		}
		for _, value := range q.Criteria {
			if !take(len(value)) {
				return size, false
			}
		}
	case BoolQuestion:
		if !take(len(q.Instructions)) {
			return size, false
		}
		if q.Criteria != nil && (!take(len(q.Criteria.True)) || !take(len(q.Criteria.False))) {
			return size, false
		}
	}
	return size, true
}

func validateClassifierQuestion(question ClassifierQuestion) string {
	var instructions json.RawMessage
	switch q := question.(type) {
	case ChoiceQuestion:
		instructions = q.Instructions
		if len(q.Criteria) < 1 || len(q.Criteria) > 255 {
			return "choice requires 1 to 255 options"
		}
		for _, v := range q.Criteria {
			if !classifierDescription(v, false) && string(bytes.TrimSpace(v)) != "null" {
				return "invalid choice criterion"
			}
		}
	case ScoreQuestion:
		instructions = q.Instructions
		if len(q.Criteria) < 2 || len(q.Criteria) > 10 {
			return "score requires 2 to 10 ordered levels"
		}
		for _, v := range q.Criteria {
			if !classifierDescription(v, false) {
				return "invalid score criterion"
			}
		}
	case BoolQuestion:
		instructions = q.Instructions
		if q.Criteria != nil && (!classifierDescription(q.Criteria.True, false) || !classifierDescription(q.Criteria.False, false)) {
			return "bool criteria require true and false descriptions"
		}
	default:
		return "unsupported classifier question"
	}
	if !classifierDescription(instructions, true) {
		return "instructions must be a nonempty JSON string, object or array"
	}
	return ""
}

func classifierDescription(raw json.RawMessage, nonempty bool) bool {
	if !classifierState(raw) || uniqueClassifierJSON(raw) != nil {
		return false
	}
	if !nonempty {
		return true
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var value any
	if d.Decode(&value) != nil {
		return false
	}
	switch v := value.(type) {
	case string:
		return v != ""
	case map[string]any:
		return len(v) > 0
	case []any:
		return len(v) > 0
	}
	return false
}

// checkClassifierInput bounds the known total before copying the caller's
// maps and raw JSON. The final encoded-body check additionally counts escaping,
// punctuation and the model field.
func (l byteLimits) checkClassifierInput(req ClassifierRequest) *Error {
	remaining := l.request
	take := func(n int) bool {
		if int64(n) > remaining {
			return false
		}
		remaining -= int64(n)
		return true
	}
	if !take(len(req.State)) {
		return limitFailure(PhaseScope, "MaxRequestBytes", l.request)
	}
	for key, q := range req.Questions {
		if !take(len(key)) {
			return limitFailure(PhaseScope, "MaxRequestBytes", l.request)
		}
		size, ok := classifierQuestionSize(q, remaining)
		if !ok {
			return limitFailure(PhaseScope, "MaxRequestBytes", l.request)
		}
		remaining -= size
	}
	return nil
}
