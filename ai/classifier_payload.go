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
	invalid := func() *Error {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback exceeded classifier input limits")
	}
	if _, ok := jsonPayloadSize(body["state"], p.MaxStateBytes); !ok {
		return invalid()
	}
	questions := reflect.ValueOf(body["questions"])
	_, jsonEncoding := body["questions"].(json.Marshaler)
	_, textEncoding := body["questions"].(encoding.TextMarshaler)
	if !jsonEncoding && !textEncoding && questions.IsValid() && questions.Kind() == reflect.Map {
		if questions.Len() > p.MaxQuestions {
			return invalid()
		}
		entries := questions.MapRange()
		for entries.Next() {
			if _, ok := jsonPayloadSize(entries.Value().Interface(), p.MaxQuestionBytes); !ok {
				return invalid()
			}
		}
	}
	if _, ok := jsonPayloadSize(body, requestLimit); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", requestLimit)
	}
	return nil
}
