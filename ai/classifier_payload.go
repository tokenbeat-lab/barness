package ai

import (
	"encoding"
	"encoding/json"
	"reflect"
)

// Check borrowed callback values before encoding or converting them into
// independent question records. Exact encoded sizes and schema are checked
// afterwards; custom marshalers remain responsible for their own execution.
func (p *ClassifierPolicy) checkPayload(body map[string]any, requestLimit int64) *Error {
	invalid := func(name string, limit int64) *Error {
		if _, images := body["assets"]; images {
			return limitFailure(PhaseRequest, name, limit)
		}
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback exceeded classifier input limits")
	}
	if _, ok := jsonPayloadSize(body["state"], p.MaxStateBytes); !ok {
		return invalid("Classifier.MaxStateBytes", p.MaxStateBytes)
	}
	questions := reflect.ValueOf(body["questions"])
	_, jsonEncoding := body["questions"].(json.Marshaler)
	_, textEncoding := body["questions"].(encoding.TextMarshaler)
	if !jsonEncoding && !textEncoding && questions.IsValid() && questions.Kind() == reflect.Map {
		if questions.Len() > p.MaxQuestions {
			return invalid("Classifier.MaxQuestions", int64(p.MaxQuestions))
		}
		entries := questions.MapRange()
		for entries.Next() {
			if _, ok := jsonPayloadSize(entries.Value().Interface(), p.MaxQuestionBytes); !ok {
				return invalid("Classifier.MaxQuestionBytes", p.MaxQuestionBytes)
			}
		}
	}
	if _, ok := jsonPayloadSize(body, requestLimit); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", requestLimit)
	}
	return nil
}
