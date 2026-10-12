package ai

import (
	"encoding/json"
	"reflect"
	"strings"
)

func decisionsSchemaFailure() *Error {
	return newError(CodeCallbackFailed, PhaseRequest, "invalid Decisions text choice request")
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
		if raw, err := marshalJS(body["input"]); err == nil && decisionsExternalInput(raw) {
			return newError(CodeTenantDenied, PhaseRequest, "payload callback may not reference external or vendor resources")
		}
		return decisionsSchemaFailure()
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

func (c *Client) freezeDecisionsRequest(body []byte, model ClassifierModel) (decisionsRequest, *Error) {
	var req decisionsRequest
	if err := decodeClassifierJSON(body, &req); err != nil {
		return req, decisionsSchemaFailure()
	}
	if req.Model != model.ID {
		return req, newError(CodeTenantDenied, PhaseRequest, "payload callback may not change the authorized model")
	}
	p := c.policy.Classifier
	if req.Input == nil || len(req.Questions) == 0 {
		return req, decisionsSchemaFailure()
	}
	if int64(len(*req.Input)) > p.MaxStateBytes {
		return req, limitFailure(PhaseRequest, "Classifier.MaxStateBytes", p.MaxStateBytes)
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
		if decodeClassifierJSON(raw.Questions[i], &req.Questions[i]) != nil || q.Type != "choice" || q.Name == "" || seen[q.Name] || q.Instructions == nil || *q.Instructions == "" || len(q.Choices) < max(2, model.Capabilities.MinChoices) || len(q.Choices) > min(255, model.Capabilities.MaxChoices) {
			return req, decisionsSchemaFailure()
		}
		seen[q.Name] = true
		var record struct {
			Choices []json.RawMessage `json:"choices"`
		}
		if json.Unmarshal(raw.Questions[i], &record) != nil {
			return req, decisionsSchemaFailure()
		}
		values := map[string]bool{}
		for j, choice := range q.Choices {
			if decodeClassifierJSON(record.Choices[j], &req.Questions[i].Choices[j]) != nil || choice.Value == nil || values[*choice.Value] {
				return req, decisionsSchemaFailure()
			}
			var fields map[string]json.RawMessage
			if json.Unmarshal(record.Choices[j], &fields) != nil {
				return req, decisionsSchemaFailure()
			}
			if _, present := fields["description"]; present && choice.Description == nil {
				return req, decisionsSchemaFailure()
			}
			values[*choice.Value] = true
		}
	}
	return req, nil
}

// Unsupported messages cannot enable image input in this slice. External
// references are still an authority denial, including behind custom JSON
// marshalers; inspect a bounded independent value without borrowing it.
func decisionsExternalInput(raw []byte) bool {
	var root any
	if json.Unmarshal(raw, &root) != nil {
		return false
	}
	pending := []any{root}
	for len(pending) > 0 {
		last := len(pending) - 1
		value := pending[last]
		pending = pending[:last]
		switch v := value.(type) {
		case map[string]any:
			for key, child := range v {
				if key == "file_id" || key == "item_reference" {
					return true
				}
				if key == "image_url" {
					if url, ok := child.(string); ok && url != "" && !strings.HasPrefix(url, "data:") {
						return true
					}
				}
				pending = append(pending, child)
			}
		case []any:
			pending = append(pending, v...)
		}
	}
	return false
}
