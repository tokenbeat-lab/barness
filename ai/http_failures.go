package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/clock"
)

// msgConnection is the Stainless SDKs' APIConnectionError text (openai-node
// and the Anthropic TypeScript SDK alike), which pi reports when no response
// arrived. An interrupted request is classified by requestInterrupted.
const msgConnection = "Connection error."

// httpFailures classifies the failures of one call's initial request, the
// part every protocol on a Stainless SDK shares: interruption, time and byte
// limits, connection failures and non-2xx responses. Every text that came
// from the provider passes through redact before it reaches a message or
// error. The protocol supplies how its SDK describes an HTTP error and which
// header carries the vendor's request id.
type httpFailures struct {
	apiKey Secret
	clock  *clock.Clock // reads a retry-after date; nil is the system clock
	// requestIDHeader names the vendor's request id response header.
	requestIDHeader string
	// describe is the protocol's text for a non-2xx response res, whose body
	// was read: msg for the message, sdk for the SDK's own error text, which
	// pi quotes when it refuses a requested retry delay.
	describe func(res *http.Response, body []byte) (msg, sdk string)
	// noResponse is the protocol's text for a request that got no response
	// at all; empty is the Stainless SDKs' msgConnection.
	noResponse string
}

// request classifies an attempt at the initial request that failed before a
// response was obtained: an interruption, a non-2xx response, or a
// connection that never produced one.
func (f httpFailures) request(ctx context.Context, err error, res *http.Response) attemptOutcome {
	if ctx.Err() != nil {
		return attemptOutcome{failure: requestInterrupted(ctx)}
	}
	// The SDK reads an error body itself and fails with the limit's error.
	var limit *limitExceeded
	if errors.As(err, &limit) && res != nil {
		return f.unread(res, limit.failure(PhaseRequest))
	}
	// An attempt that ran out of time before its response is, as the SDKs'
	// APIConnectionTimeoutError, a connection failure the retry rules retry.
	var expired *timeLimitExpired
	if errors.As(err, &expired) {
		if res != nil {
			// The error body stalled: what the headers tell is kept.
			return f.unread(res, expired.failure(PhaseRequest))
		}
		return attemptOutcome{failure: expired.failure(PhaseRequest), connection: true}
	}
	if res != nil && res.StatusCode >= 300 {
		return f.status(res)
	}
	var conn *connectionFailure
	msg := msgConnection
	if f.noResponse != "" {
		msg = f.noResponse
	}
	return attemptOutcome{failure: newError(CodeTransport, PhaseRequest, msg), connection: errors.As(err, &conn)}
}

// status classifies a non-2xx response. The SDK has already read the body of
// an error status and left a copy in res.Body; a redirect's body is unread.
// Either was read through the MaxErrorBodyBytes limit (limitBodies).
func (f httpFailures) status(res *http.Response) attemptOutcome {
	body, err := io.ReadAll(res.Body)
	var limit *limitExceeded
	var expired *timeLimitExpired
	switch {
	case errors.As(err, &limit):
		// The body is not described, and the attempt is never retried (see
		// attemptOutcome.retryable).
		return f.unread(res, limit.failure(PhaseRequest))
	case errors.As(err, &expired):
		return f.unread(res, expired.failure(PhaseRequest))
	}
	msg, sdkMsg := f.describe(res, body)
	e := newError(codeForStatus(res.StatusCode), PhaseRequest, f.redact(msg))
	o := f.unread(res, e)
	o.sdkMessage = f.redact(sdkMsg)
	return o
}

// unread is a non-2xx response whose body could not be read for e's
// reason, or was described into e. It keeps what the headers tell.
func (f httpFailures) unread(res *http.Response, e *Error) attemptOutcome {
	e.HTTPStatus = res.StatusCode
	e.ProviderRequestID = res.Header.Get(f.requestIDHeader)
	e.RetryAfter = retryAfter(res.Header, f.clock.Now())
	return attemptOutcome{status: res.StatusCode, header: res.Header, providerRequestID: e.ProviderRequestID, failure: e}
}

// upstream is a failure the provider reported inside the stream.
func (f httpFailures) upstream(msg string) *Error {
	return newError(CodeUpstreamError, PhaseStream, f.redact(msg))
}

// keyLike matches OpenAI- and Anthropic-style API keys ("sk-…", "sk-ant-…"),
// masked ones included ("sk-proj-****abcd"), and Google API keys ("AIza" and
// 35 more characters): vendors echo them in auth errors. The word boundary
// keeps ordinary words such as "task-runner" intact.
var keyLike = regexp.MustCompile(`\bsk-[A-Za-z0-9_*\-]{3,}|\bAIza[A-Za-z0-9_\-]{35}`)

// redact removes the call's key and anything key-shaped from provider text.
// pi passes such text through; this is a recorded security difference (spec
// I9: secrets never reach messages or errors).
func (f httpFailures) redact(text string) string {
	text = strings.ReplaceAll(text, f.apiKey.reveal(), "[REDACTED]")
	return keyLike.ReplaceAllString(text, "[REDACTED]")
}

// jsTruthy is JavaScript truthiness for a decoded JSON value.
func jsTruthy(v any) bool {
	switch v := v.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	}
	return true
}

// compactJSON approximates JSON.stringify of a value received as JSON: key
// order and content are kept, whitespace removed. Escapes and number
// spellings stay as sent, where JSON.stringify would normalize them.
func compactJSON(raw json.RawMessage) string {
	var buf bytes.Buffer
	if json.Compact(&buf, raw) != nil {
		return string(raw)
	}
	return buf.String()
}

// bodyReadFailure classifies an error that ended a streamed response body
// early: a byte or time limit of the policy, else the connection was lost,
// which lost describes in the protocol's words.
func bodyReadFailure(err error, lost string) *Error {
	var limit *limitExceeded
	var expired *timeLimitExpired
	switch {
	case errors.As(err, &limit):
		return limit.failure(PhaseStream)
	case errors.As(err, &expired):
		return expired.failure(PhaseStream)
	}
	return newError(CodeTransport, PhaseStream, lost)
}
