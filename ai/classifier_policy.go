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
	if !classifierState(req.State) {
		return newError(CodeInvalidRequest, PhaseScope, "classifier state must be a JSON string, object or array")
	}
	if len(req.Questions) == 0 {
		return newError(CodeInvalidRequest, PhaseScope, "classifier requires at least one question")
	}
	for _, key := range slices.Sorted(maps.Keys(req.Questions)) {
		if key == "" {
			return newError(CodeInvalidRequest, PhaseScope, "classifier question key must not be empty")
		}
		q, ok := req.Questions[key].(ChoiceQuestion)
		if !ok {
			return newError(CodeInvalidRequest, PhaseScope, "classifier requires a choice question")
		}
		if len(q.Criteria) < 1 || len(q.Criteria) > 255 {
			return newError(CodeInvalidRequest, PhaseScope, "choice requires 1 to 255 options")
		}
		// Summing known lengths is bounded by the policy on each iteration; no
		// encoding or deep copy of an oversized question occurs first.
		size := int64(len(q.Instructions))
		if size > p.MaxQuestionBytes {
			return limitFailure(PhaseScope, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
		}
		for k, v := range q.Criteria {
			n := int64(len(k)) + int64(len(v))
			if n > p.MaxQuestionBytes-size {
				return limitFailure(PhaseScope, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
			}
			size += n
		}
		instructions, ok := classifierString(q.Instructions)
		if !ok || instructions == "" {
			return newError(CodeInvalidRequest, PhaseScope, "choice instructions must be a nonempty JSON string")
		}
		for _, v := range q.Criteria {
			if _, ok := classifierString(v); !ok && string(bytes.TrimSpace(v)) != "null" {
				return newError(CodeInvalidRequest, PhaseScope, "choice criteria must be JSON strings or null")
			}
		}
		// Count the complete encoded question too, including JSON key escaping
		// and punctuation, after the cheap check bounded the allocation.
		body, err := json.Marshal(q)
		if err != nil {
			return newError(CodeInvalidRequest, PhaseScope, "choice question could not be encoded")
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
	if !slices.Contains(c.Kinds, ClassifierQuestionChoice) {
		return "model does not support choice questions"
	}
	for _, q := range r.Questions {
		if len(q.(ChoiceQuestion).Criteria) > c.MaxChoices {
			return "choice options exceed model capability"
		}
	}
	return ""
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
		if q, ok := q.(ChoiceQuestion); ok {
			if !take(len(q.Instructions)) {
				return limitFailure(PhaseScope, "MaxRequestBytes", l.request)
			}
			for key, value := range q.Criteria {
				if !take(len(key)) || !take(len(value)) {
					return limitFailure(PhaseScope, "MaxRequestBytes", l.request)
				}
			}
		}
	}
	return nil
}

func classifierString(raw json.RawMessage) (string, bool) { return rawString(bytes.TrimSpace(raw)) }
