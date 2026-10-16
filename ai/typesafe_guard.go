package ai

import (
	"bytes"
	"encoding/json"
	"maps"
	"slices"
)

func (c *Client) freezeTypeSafeRequest(body []byte, model ClassifierModel) (typeSafeRequest, *Error) {
	var req typeSafeRequest
	if err := decodeClassifierJSON(body, &req); err != nil {
		return req, typeSafeRequestFailure()
	}
	if req.Model != model.ID {
		return req, newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	p := c.policy.Classifier
	if len(req.Questions) == 0 || len(req.Questions) > p.MaxQuestions || int64(len(req.State)) > p.MaxStateBytes || !classifierState(req.State) {
		return req, typeSafeRequestFailure()
	}
	for _, name := range slices.Sorted(maps.Keys(req.Questions)) {
		q := req.Questions[name]
		if name == "" || !q.withinSize(p.MaxQuestionBytes) || !q.valid() {
			return req, typeSafeRequestFailure()
		}
		// Preserve the old canonical per-question budget, including HTML
		// escaping. Native noul and public bool have equal encoded lengths;
		// using the received raw length would silently widen Hook limits.
		raw, err := json.Marshal(q)
		if err != nil || int64(len(raw)) > p.MaxQuestionBytes {
			return req, typeSafeRequestFailure()
		}
	}
	// Shape and sub-policy failures historically precede model capabilities.
	// Keep that priority and their callback_failed/request classification.
	for _, q := range req.Questions {
		if !q.supported(model.Capabilities) {
			return req, newError(CodeCallbackFailed, PhaseRequest, "payload callback exceeded classifier model capabilities")
		}
	}
	return req, nil
}

func typeSafeRequestFailure() *Error {
	return newError(CodeCallbackFailed, PhaseRequest, "payload callback produced an invalid classifier request")
}

// Native criteria evolve independently of Host question variants. Account
// their raw lengths before canonical encoding rather than constructing a
// public question to reuse its validation or capacity rules.
func (q typeSafeQuestion) withinSize(limit int64) bool {
	take := func(n int) bool {
		if int64(n) > limit {
			return false
		}
		limit -= int64(n)
		return true
	}
	if !take(len(q.instructions)) {
		return false
	}
	for key, value := range q.choices {
		if !take(len(key)) || !take(len(value)) {
			return false
		}
	}
	for _, value := range q.levels {
		if !take(len(value)) {
			return false
		}
	}
	return q.boolean == nil || (take(len(q.boolean.True)) && take(len(q.boolean.False)))
}

func (q typeSafeQuestion) valid() bool {
	if !classifierDescription(q.instructions, true) {
		return false
	}
	switch q.kind {
	case "choice":
		if len(q.choices) < 1 || len(q.choices) > 255 {
			return false
		}
		for _, value := range q.choices {
			if !bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !classifierDescription(value, false) {
				return false
			}
		}
	case "score":
		if len(q.levels) < 2 || len(q.levels) > 10 {
			return false
		}
		for _, value := range q.levels {
			if !classifierDescription(value, false) {
				return false
			}
		}
	case "noul":
		return q.boolean == nil || (classifierDescription(q.boolean.True, false) && classifierDescription(q.boolean.False, false))
	default:
		return false
	}
	return true
}

func (q typeSafeQuestion) supported(c ClassifierCapabilities) bool {
	var kind ClassifierQuestionKind
	switch q.kind {
	case "choice":
		kind = ClassifierQuestionChoice
		if len(q.choices) < max(1, c.MinChoices) || len(q.choices) > c.MaxChoices {
			return false
		}
	case "score":
		kind = ClassifierQuestionScore
		if len(q.levels) < c.MinScoreLevels || len(q.levels) > c.MaxScoreLevels {
			return false
		}
	case "noul":
		kind = ClassifierQuestionBool
	}
	if !slices.Contains(c.Kinds, kind) || !classifierDescriptionAllowed(q.instructions, false, c.DescriptionFormats) {
		return false
	}
	for _, value := range q.choices {
		if !classifierDescriptionAllowed(value, true, c.DescriptionFormats) {
			return false
		}
	}
	for _, value := range q.levels {
		if !classifierDescriptionAllowed(value, false, c.DescriptionFormats) {
			return false
		}
	}
	return q.boolean == nil || (classifierDescriptionAllowed(q.boolean.True, false, c.DescriptionFormats) && classifierDescriptionAllowed(q.boolean.False, false, c.DescriptionFormats))
}
