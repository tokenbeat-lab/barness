package ai

import (
	"context"
	"errors"
)

// Code is a stable error classification usable with errors.Is via Error.Is.
type Code string

const (
	CodeInvalidRequest        Code = "invalid_request"
	CodeTenantDenied          Code = "tenant_denied"
	CodeBindingNotFound       Code = "binding_not_found"
	CodeCredentialUnavailable Code = "credential_unavailable"
	CodeAdmissionDenied       Code = "admission_denied"
	CodeUpstreamAuth          Code = "upstream_auth"
	CodeRateLimited           Code = "rate_limited"
	CodeTransport             Code = "transport"
	CodeProtocol              Code = "protocol"
	CodeCanceled              Code = "canceled"
	CodeDeadlineExceeded      Code = "deadline_exceeded"
	CodeResourceLimit         Code = "resource_limit"
)

// Phase is where in the call an error happened.
type Phase string

const (
	PhaseScope      Phase = "scope"
	PhaseBinding    Phase = "binding"
	PhaseCapability Phase = "capability"
	PhaseCredential Phase = "credential"
	PhaseRequest    Phase = "request"
	PhaseStream     Phase = "stream"
)

// Error is the classified error returned with a failed Result. It never
// carries secrets or secret-backend internals.
type Error struct {
	Code       Code
	Phase      Phase
	Message    string
	HTTPStatus int // 0 when no HTTP response was received
}

func (e *Error) Error() string {
	return "ai: " + string(e.Code) + " (" + string(e.Phase) + "): " + e.Message
}

// Is matches another *Error by Code, so errors.Is(err, &ai.Error{Code: ...})
// tests the classification.
func (e *Error) Is(target error) bool {
	var t *Error
	return errors.As(target, &t) && t.Code == e.Code && (t.Phase == "" || t.Phase == e.Phase)
}

func newError(code Code, phase Phase, msg string) *Error {
	return &Error{Code: code, Phase: phase, Message: msg}
}

// contextError classifies a context termination, or returns nil.
func contextError(ctx context.Context, phase Phase) *Error {
	switch err := ctx.Err(); {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return newError(CodeDeadlineExceeded, phase, "Request timed out")
	default:
		return newError(CodeCanceled, phase, "Request was aborted")
	}
}
