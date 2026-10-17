package ai

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// GLM errors are deliberately described from documented codes, never provider
// messages: error bodies can echo prompts, reasoning or credentials. Responses
// has its own named code set; it does not use Chat's numeric business codes.
// Sources and synthetic Messages wire limits: ticket-08/sources and README.
func (f httpFailures) glmCode(raw json.RawMessage) string {
	var envelope struct {
		Error json.RawMessage `json:"error"`
		Code  json.RawMessage `json:"code"`
		Type  string          `json:"type"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return ""
	}
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		return f.glmCode(envelope.Error)
	}
	if f.glm == APIAnthropicMessages && !rawTruthy(envelope.Code) {
		switch envelope.Type {
		case "authentication_error", "permission_error", "invalid_request_error", "not_found_error", "rate_limit_error", "overloaded_error", "api_error":
			return envelope.Type
		}
	}
	s, ok := rawString(envelope.Code)
	if !ok {
		// Numeric business codes belong to Chat and the explicit Messages
		// compatibility fixture, never to the Responses string schema.
		if f.glm == APIOpenAIResponses {
			return ""
		}
		var n json.Number
		if json.Unmarshal(envelope.Code, &n) != nil {
			return ""
		}
		s = n.String()
	}
	if f.glm == APIOpenAIResponses {
		switch s {
		case "invalid_request", "model_not_found", "not_implemented", "request_too_large",
			"authentication_error", "invalid_api_key", "expired", "permission_denied",
			"context_length_exceeded", "insufficient_quota", "quota_exceeded",
			"usage_limit_reached", "usage_not_included", "server_error",
			"rate_limit_exceeded", "server_is_overloaded", "overloaded", "slow_down",
			"content_filter", "cyber_policy":
			return s
		}
	}
	// Keep unknown numeric codes for diagnosis without echoing arbitrary text.
	if len(s) > 0 && len(s) <= 9 {
		if n, err := strconv.ParseUint(s, 10, 32); err == nil && strconv.FormatUint(n, 10) == s {
			return s
		}
	}
	return ""
}

func (f httpFailures) glmPermanent(code string) bool {
	if f.glm == APIOpenAIResponses {
		switch code {
		case "invalid_request", "model_not_found", "not_implemented", "request_too_large",
			"authentication_error", "invalid_api_key", "expired", "permission_denied",
			"context_length_exceeded", "insufficient_quota", "quota_exceeded",
			"usage_limit_reached", "usage_not_included", "content_filter", "cyber_policy":
			return true
		}
		return false
	}
	switch code {
	case "authentication_error", "permission_error", "invalid_request_error", "not_found_error",
		"1000", "1001", "1003", "1005", "1113", "1210", "1211", "1212", "1213", "1214", "1215",
		"1220", "1221", "1222", "1261", "1301", "1308", "1309", "1310", "1311", "1313", "1314", "1315",
		"1316", "1317", "1318", "1319", "1320", "1321":
		return true
	}
	return false
}

func glmErrorMessage(status int, code string) string {
	message := "GLM API error"
	if status != 0 {
		message += fmt.Sprintf(" (HTTP %d)", status)
	}
	if code != "" {
		message += "; code " + code
	}
	return message
}

func (f httpFailures) providerError(raw json.RawMessage) *Error {
	if f.glm != "" {
		return newError(CodeUpstreamError, PhaseStream, glmErrorMessage(0, f.glmCode(raw)))
	}
	return f.upstream(apiErrorText(raw))
}

// providerCodeError keeps a typed Responses code on the same safe path as raw
// error objects. The provider's free-form message never enters diagnostics.
func (f httpFailures) providerCodeError(code string) *Error {
	raw, _ := json.Marshal(map[string]string{"code": code})
	return f.providerError(raw)
}
