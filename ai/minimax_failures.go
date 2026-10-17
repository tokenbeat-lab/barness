package ai

import (
	"encoding/json"
	"fmt"
	"slices"
)

// MiniMax Messages ErrorResponse independently documents these compatibility
// types (official OpenAPI snapshot 2026-10-18):
// https://platform.minimax.cn/docs/api-reference/text/api/openapi-chat-anthropic.json
// Chat/Responses numeric base_resp
// codes and message words are not evidence of compatible quota semantics.
func (f httpFailures) miniMaxErrorType(raw json.RawMessage) string {
	if f.minimax != APIAnthropicMessages {
		return ""
	}
	var body struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &body) != nil {
		return ""
	}
	if slices.Contains([]string{"invalid_request_error", "authentication_error", "permission_error", "not_found_error", "request_too_large", "rate_limit_error", "api_error", "overloaded_error"}, body.Error.Type) {
		return body.Error.Type
	}
	return ""
}
func miniMaxPermanentErrorType(typ string) bool {
	return slices.Contains([]string{"invalid_request_error", "authentication_error", "permission_error", "not_found_error", "request_too_large"}, typ)
}

// Provider bodies can echo credentials, prompts, reasoning or tool arguments.
// Diagnostics retain only HTTP facts and a documented finite type. Compatible
// billing/plan codes and trace priority await own live capture in ticket 10.
func miniMaxErrorMessage(status int, typ string) string {
	msg := "MiniMax API error"
	if status != 0 {
		msg += fmt.Sprintf(" (HTTP %d)", status)
	}
	if typ != "" {
		msg += "; type " + typ
	}
	return msg
}

// HTTP itself establishes these permanent request/auth/permission refusals.
// A retry hint cannot turn them into a replayable account operation. Unknown
// 429 and transient408/409/5xx keep the existing explicit bounded retry policy.
func miniMaxPermanentHTTPStatus(status int) bool {
	return status == 400 || status == 401 || status == 403 || status == 404 || status == 413 || status == 422
}
