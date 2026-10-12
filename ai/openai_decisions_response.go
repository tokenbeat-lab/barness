package ai

import (
	"encoding/json"
	"strings"
)

func decodeDecisionsResponse(raw []byte, model ClassifierModel, initial *initialRequest, req decisionsRequest, out *ClassifierResult) *Error {
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil {
		return classifierProtocolFailure()
	}
	reporting, usage, failure := decisionsUsage(fields["usage"], model.Cost)
	out.Usage = usage
	initial.usage(reporting, usage)
	if failure != nil {
		return failure
	}
	if rawModel, ok := fields["model"]; ok {
		var value *string
		if json.Unmarshal(rawModel, &value) != nil || value == nil {
			return classifierProtocolFailure()
		}
		out.ResponseModel = Value(*value)
	}
	if duplicate || uniqueClassifierJSON(raw) != nil {
		return classifierProtocolFailure()
	}
	for key := range fields {
		for _, known := range []string{"answers", "model", "usage"} {
			if key != known && strings.EqualFold(key, known) {
				return classifierProtocolFailure()
			}
		}
	}
	var answers []json.RawMessage
	if json.Unmarshal(fields["answers"], &answers) != nil || len(answers) != len(req.Questions) {
		return classifierProtocolFailure()
	}
	questions := make(map[string]decisionsQuestion, len(req.Questions))
	for _, q := range req.Questions {
		questions[q.Name] = q
	}
	validated := make(map[string]ClassifierAnswer, len(answers))
	for _, raw := range answers {
		var tag struct {
			Type string  `json:"type"`
			Name *string `json:"name"`
		}
		if json.Unmarshal(raw, &tag) != nil || tag.Name == nil {
			return classifierProtocolFailure()
		}
		q, ok := questions[*tag.Name]
		if !ok || validated[*tag.Name] != nil {
			return classifierProtocolFailure()
		}
		if tag.Type == "refusal" {
			if !decisionsAnswerFields(raw, "type", "name") {
				return classifierProtocolFailure()
			}
			validated[*tag.Name] = RefusalAnswer{}
			continue
		}
		if tag.Type != q.Type {
			return classifierProtocolFailure()
		}
		var answer ClassifierAnswer
		var failure *Error
		switch q.Type {
		case "choice":
			answer, failure = decodeDecisionsChoice(raw, q)
		case "score":
			answer, failure = decodeDecisionsScore(raw, q)
		case "predicate":
			answer, failure = decodeDecisionsPredicate(raw)
		default:
			return classifierProtocolFailure()
		}
		if failure != nil {
			return failure
		}
		validated[*tag.Name] = answer
	}
	out.Answers = validated
	return nil
}
