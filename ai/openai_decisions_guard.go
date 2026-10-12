package ai

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
)

func decisionsSchemaFailure() *Error {
	return newError(CodeCallbackFailed, PhaseRequest, "invalid Decisions request")
}

// Check borrowed hook shapes and sizes before allocating the encoded body.
// Exact schema and per-question encoded sizes are checked again on its bytes.
func (c *Client) checkDecisionsPayload(body map[string]any, model ClassifierModel) *Error {
	if body["model"] != model.ID {
		return newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	for _, key := range []string{"endpoint", "api_key", "headers", "provider", "api", "operation", "credential", "tools", "allowClassifierImages", "allow_classifier_images", "image_url", "file_id"} {
		if _, ok := body[key]; ok {
			return newError(CodeTenantDenied, PhaseRequest, "payload callback may not add authority fields")
		}
	}
	for key := range body {
		if key != "model" && key != "input" && key != "questions" {
			return decisionsSchemaFailure()
		}
	}
	p := c.policy.Classifier
	if text, ok := body["input"].(string); ok {
		if int64(len(text)) > p.MaxStateBytes {
			return limitFailure(PhaseRequest, "Classifier.MaxStateBytes", p.MaxStateBytes)
		}
	} else {
		if _, ok := jsonPayloadSize(body["input"], c.policy.MaxRequestBytes); !ok {
			return limitFailure(PhaseRequest, "MaxRequestBytes", c.policy.MaxRequestBytes)
		}
		surround, ok := jsonPayloadSize(map[string]any{"model": body["model"], "input": nil, "questions": body["questions"]}, c.policy.MaxRequestBytes)
		if !ok {
			return limitFailure(PhaseRequest, "MaxRequestBytes", c.policy.MaxRequestBytes)
		}
		// The null placeholder occupies four bytes; reserve everything else.
		if failure := c.checkDecisionsInputPayload(body["input"], c.policy.MaxRequestBytes-surround+4); failure != nil {
			return failure
		}
	}
	questions := reflect.ValueOf(body["questions"])
	if questions.IsValid() && (questions.Kind() == reflect.Slice || questions.Kind() == reflect.Array) {
		if questions.Len() > p.MaxQuestions {
			return limitFailure(PhaseRequest, "Classifier.MaxQuestions", int64(p.MaxQuestions))
		}
		for i := 0; i < questions.Len(); i++ {
			if _, ok := jsonPayloadSize(questions.Index(i).Interface(), p.MaxQuestionBytes); !ok {
				return limitFailure(PhaseRequest, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
			}
		}
	}
	if _, ok := jsonPayloadSize(body, c.policy.MaxRequestBytes); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", c.policy.MaxRequestBytes)
	}
	return nil
}

func (c *Client) freezeDecisionsRequest(ctx context.Context, body []byte, model ClassifierModel, allowImages bool) (decisionsRequest, *Error) {
	var req decisionsRequest
	if failure := contextError(ctx, PhaseRequest); failure != nil {
		return req, failure
	}
	if failure := c.checkDecisionsRawPayload(body); failure != nil {
		return req, failure
	}
	if err := decodeClassifierJSON(body, &req); err != nil {
		return req, decisionsSchemaFailure()
	}
	if req.Model != model.ID {
		return req, newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	p := c.policy.Classifier
	if len(req.Input) == 0 || len(req.Questions) == 0 {
		return req, decisionsSchemaFailure()
	}
	if failure := c.checkDecisionsInput(ctx, req.Input, model, allowImages); failure != nil {
		return req, failure
	}
	if len(req.Questions) > p.MaxQuestions {
		return req, limitFailure(PhaseRequest, "Classifier.MaxQuestions", int64(p.MaxQuestions))
	}
	var raw struct {
		Questions []json.RawMessage `json:"questions"`
	}
	if json.Unmarshal(body, &raw) != nil {
		return req, decisionsSchemaFailure()
	}
	seen := map[string]bool{}
	for i, q := range req.Questions {
		if int64(len(raw.Questions[i])) > p.MaxQuestionBytes {
			return req, limitFailure(PhaseRequest, "Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
		}
		if q.Name == "" || seen[q.Name] || q.Instructions == nil || *q.Instructions == "" || !validDecisionsQuestion(raw.Questions[i], q, model.Capabilities) {
			return req, decisionsSchemaFailure()
		}
		seen[q.Name] = true
	}
	return req, nil
}

// The union record stores frozen values, but each native kind has a closed
// schema. Even empty/null fields from another kind must fail the final guard.
func validDecisionsQuestion(raw []byte, q decisionsQuestion, caps ClassifierCapabilities) bool {
	switch q.Type {
	case "choice":
		var wire struct {
			Type         string            `json:"type"`
			Name         string            `json:"name"`
			Instructions *string           `json:"instructions"`
			Choices      []json.RawMessage `json:"choices"`
		}
		if !slices.Contains(caps.Kinds, ClassifierQuestionChoice) || decodeClassifierJSON(raw, &wire) != nil || len(q.Choices) < max(2, caps.MinChoices) || len(q.Choices) > min(255, caps.MaxChoices) {
			return false
		}
		values := map[string]bool{}
		for i, choice := range q.Choices {
			var checked decisionsChoice
			if decodeClassifierJSON(wire.Choices[i], &checked) != nil || choice.Value == nil || values[*choice.Value] || !decisionsDescriptionValid(wire.Choices[i], choice.Description) {
				return false
			}
			values[*choice.Value] = true
		}
	case "score":
		var wire struct {
			Type         string            `json:"type"`
			Name         string            `json:"name"`
			Instructions *string           `json:"instructions"`
			Levels       []json.RawMessage `json:"levels"`
		}
		if !slices.Contains(caps.Kinds, ClassifierQuestionScore) || decodeClassifierJSON(raw, &wire) != nil || len(q.Levels) < max(2, caps.MinScoreLevels) || len(q.Levels) > min(10, caps.MaxScoreLevels) {
			return false
		}
		for i, level := range q.Levels {
			var checked decisionsLevel
			if decodeClassifierJSON(wire.Levels[i], &checked) != nil || level.Label == nil || !decisionsDescriptionValid(wire.Levels[i], level.Description) {
				return false
			}
		}
	case "predicate":
		var wire struct {
			Type         string  `json:"type"`
			Name         string  `json:"name"`
			Instructions *string `json:"instructions"`
		}
		return slices.Contains(caps.Kinds, ClassifierQuestionBool) && decodeClassifierJSON(raw, &wire) == nil
	default:
		return false
	}
	return true
}

func decisionsDescriptionValid(raw []byte, value *string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return false
	}
	_, present := fields["description"]
	return !present || value != nil
}
