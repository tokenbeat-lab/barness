package ai

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Code is a stable error classification usable with errors.Is via Error.Is.
type Code string

const (
	CodeInvalidRequest        Code = "invalid_request"
	CodeTenantDenied          Code = "tenant_denied"
	CodeBindingNotFound       Code = "binding_not_found"
	CodeCredentialUnavailable Code = "credential_unavailable"
	CodeAdmissionDenied       Code = "admission_denied"
	// CodeUpstreamAuth: the provider refused the credential (HTTP 401/403).
	// barness-ai never switches keys or falls back to another identity.
	CodeUpstreamAuth Code = "upstream_auth"
	// CodeRateLimited: the provider rate-limited the request (HTTP 429).
	CodeRateLimited Code = "rate_limited"
	// CodeUpstreamError: the provider reported a failure of its own: an HTTP
	// 5xx, 408 or 409, an in-stream error, a failed response, or an incomplete response
	// for a reason other than the output limit.
	CodeUpstreamError    Code = "upstream_error"
	CodeTransport        Code = "transport"
	CodeProtocol         Code = "protocol"
	CodeCanceled         Code = "canceled"
	CodeDeadlineExceeded Code = "deadline_exceeded"
	CodeResourceLimit    Code = "resource_limit"
	// CodeCallbackFailed: a trusted host's request callback (Hooks) returned
	// an error or produced no usable request body. It is the host's own
	// fault, kept apart from an invalid request so the host can tell the two
	// apart; a callback's attempt to widen authorization is CodeTenantDenied.
	CodeCallbackFailed Code = "callback_failed"
)

// Phase is where in the call an error happened.
type Phase string

const (
	// PhaseScope is the first check of a call: its scope and the structure of
	// its request (for example, tool declarations), before anything resolves.
	PhaseScope      Phase = "scope"
	PhaseBinding    Phase = "binding"
	PhaseCapability Phase = "capability"
	PhaseCredential Phase = "credential"
	// PhaseConsistency is the check that the binding and credential snapshots
	// agree on tenant, account, reference and version (D2, ADR-0003).
	PhaseConsistency Phase = "consistency"
	PhaseRequest     Phase = "request"
	PhaseStream      Phase = "stream"
	// PhaseEventQueue is a Stream's event queue: its consumer fell behind the
	// policy's MaxQueuedEvents or MaxQueuedEventBytes.
	PhaseEventQueue Phase = "event_queue"
)

// Error is the classified error returned with a failed Result. It never
// carries secrets or secret-backend internals. Its Message is the message's
// ErrorMessage; Code and Phase classify it without replacing the message's
// StopReason.
type Error struct {
	Code    Code
	Phase   Phase
	Message string
	// HTTPStatus is the provider's response status; 0 when no HTTP response
	// was received.
	HTTPStatus int
	// ProviderRequestID is the vendor's own request id (for OpenAI the
	// x-request-id response header), when a response carried one.
	ProviderRequestID string
	// RetryAfter is the delay the provider asked for, from retry-after-ms or
	// retry-after (seconds or an HTTP date); 0 when it asked for none or for
	// a time already past.
	RetryAfter time.Duration
	// cause is a trusted host callback's own error, kept for the host's
	// diagnosis through Unwrap; it never enters Message.
	cause error
}

// Unwrap returns the trusted host callback error that caused this failure,
// or nil.
func (e *Error) Unwrap() error { return e.cause }

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

// contextError classifies a context termination during setup, or returns nil.
// pi has no setup stage of its own, so these texts are barness's.
func contextError(ctx context.Context, phase Phase) *Error {
	switch err := ctx.Err(); {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return interrupted(ctx, phase, "Request timed out")
	default:
		return interrupted(ctx, phase, "Request was aborted")
	}
}

// interrupted classifies an ended context as canceled or deadline_exceeded.
// The phase decides whether the message ends as error or aborted (see
// Error.aborts); msg is the phase's own text.
func interrupted(ctx context.Context, phase Phase, msg string) *Error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return newError(CodeDeadlineExceeded, phase, msg)
	}
	return newError(CodeCanceled, phase, msg)
}

// codeForStatus classifies a provider's non-2xx HTTP status. An auth refusal
// is reported as such and never retried with another identity. 408 and 409
// are the provider's transient conditions (the baseline retries them), not
// an invalid request.
func codeForStatus(status int) Code {
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return CodeUpstreamAuth
	case status == http.StatusTooManyRequests:
		return CodeRateLimited
	case status >= 500 || status == http.StatusRequestTimeout || status == http.StatusConflict:
		return CodeUpstreamError
	case status >= 400:
		return CodeInvalidRequest
	default:
		// A redirect is never followed (spec I5) and nothing else is expected.
		return CodeProtocol
	}
}

// retryAfter reads the delay a provider asked for, as the retry rules read
// it (serverDelayMs), at now.
func retryAfter(h http.Header, now time.Time) time.Duration {
	ms, _ := serverDelayMs(h, now)
	return msDuration(ms)
}
