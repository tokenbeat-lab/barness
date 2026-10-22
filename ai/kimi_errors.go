package ai

import (
	"encoding/json"
	"strconv"
)

// Only documented categories may enter diagnostics. Provider messages and
// unknown type strings are untrusted text, even when the HTTP status is known.
func kimiErrorType(raw json.RawMessage) string {
	var object struct {
		Type  string          `json:"type"`
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &object) != nil {
		return "unknown"
	}
	if len(object.Error) > 0 && string(object.Error) != "null" {
		var nested struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(object.Error, &nested) == nil {
			object.Type = nested.Type
		}
	}
	switch object.Type {
	case "content_filter", "invalid_request_error", "invalid_authentication_error", "incorrect_api_key_error", "permission_denied_error", "resource_not_found_error", "engine_overloaded_error", "exceeded_current_quota_error", "rate_limit_reached_error", "client_closed_request", "internal_server_error", "server_overloaded_error", "gateway_timeout_error":
		return object.Type
	}
	return "unknown"
}

func kimiErrorMessage(status int, kind string) string {
	if status == 0 {
		return "Kimi stream error (" + kind + ")"
	}
	return "Kimi request failed (HTTP " + strconv.Itoa(status) + ", " + kind + ")"
}

func kimiPermanent(status int, kind string) bool {
	switch kind {
	case "exceeded_current_quota_error", "invalid_request_error", "content_filter", "invalid_authentication_error", "incorrect_api_key_error", "permission_denied_error", "resource_not_found_error":
		return true
	}
	if status == 429 {
		return kind != "engine_overloaded_error" && kind != "rate_limit_reached_error"
	}
	return status >= 400 && status < 408 || status > 409 && status < 429 || status > 429 && status < 500
}

// kimiProviderError recognizes an error envelope inside a successful HTTP
// stream. It never includes the provider's free text or inferred request IDs.
func (f httpFailures) kimiProviderError(raw json.RawMessage) *Error {
	var object struct {
		Type  string          `json:"type"`
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &object) != nil {
		return nil
	}
	if (len(object.Error) == 0 || string(object.Error) == "null") && object.Type != "error" {
		return nil
	}
	return newError(CodeUpstreamError, PhaseStream, kimiErrorMessage(0, kimiErrorType(raw)))
}
