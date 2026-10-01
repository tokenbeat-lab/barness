package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/openai/openai-go/v3/packages/respjson"
	"github.com/openai/openai-go/v3/packages/ssestream"
)

// Error texts barness-ai shares with pi-ai's Responses path, so a caller sees
// the same errorMessage for the same failure.
const (
	// openai-node's APIUserAbortError / APIConnectionTimeoutError /
	// APIConnectionError: the request was interrupted before a response.
	msgRequestAborted  = "Request was aborted."
	msgRequestTimedOut = "Request timed out."
	msgConnection      = "Connection error."
	// pi's explicit check after a stream that reached its terminal.
	msgAbortedAfterTerminal = "Request was aborted"
	// pi's check for a stream without a terminal event. A stream interrupted
	// by cancellation ends this way in pi too: the SDK ends it quietly.
	msgNoTerminal = "OpenAI Responses stream ended before a terminal response event"
)

// Error texts pi takes from its JavaScript runtime (JSON.parse, undici), which
// barness cannot reproduce; recorded as approved differences in the pi ledger.
const (
	msgInvalidJSON    = "OpenAI Responses stream event is not valid JSON"
	msgUndecodable    = "OpenAI Responses stream event could not be decoded"
	msgConnectionLost = "Connection lost while reading the OpenAI Responses stream"
)

// maxErrorBodyChars is pi's MAX_PROVIDER_ERROR_BODY_CHARS (UTF-16 units).
const maxErrorBodyChars = 4000

// responsesFailures classifies one call's failures. Every text that came from
// the provider passes through redact before it reaches a message or error.
type responsesFailures struct {
	provider ProviderID
	apiKey   Secret
}

// request classifies a failure before the stream started: an interruption,
// a non-2xx response, or a connection that never produced one.
func (f responsesFailures) request(ctx context.Context, err error, res *http.Response) *Error {
	if ctx.Err() != nil {
		msg := msgRequestAborted
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			msg = msgRequestTimedOut
		}
		return interrupted(ctx, PhaseRequest, msg)
	}
	if res != nil && res.StatusCode >= 300 {
		return f.status(res)
	}
	return newError(CodeTransport, PhaseRequest, msgConnection)
}

// status classifies a non-2xx response. The SDK has already read the body of
// an error status and left a copy in res.Body; a redirect's body is unread.
func (f responsesFailures) status(res *http.Response) *Error {
	// Unbounded for now: the SDK has already read an error body whole, so this
	// adds no allocation, but a redirect's body is read here and a hostile
	// upstream can make either large. Ticket 12 bounds both by
	// MaxErrorBodyBytes; remove this note then.
	body, _ := io.ReadAll(res.Body)
	e := newError(codeForStatus(res.StatusCode), PhaseRequest, f.redact(describeHTTPError(f.prefix(), res.StatusCode, body)))
	e.HTTPStatus = res.StatusCode
	e.ProviderRequestID = res.Header.Get("x-request-id")
	e.RetryAfter = retryAfter(res.Header)
	return e
}

// stream classifies an error that ended the SSE stream early.
func (f responsesFailures) stream(err error) *Error {
	var streamErr *ssestream.StreamError
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &streamErr):
		// The SDK stops at any frame with a top-level "error"; openai-node
		// raises it as an APIError built from that object.
		var frame struct {
			Error json.RawMessage `json:"error"`
		}
		_ = json.Unmarshal(streamErr.Event.Data, &frame)
		return newError(CodeUpstreamError, PhaseStream, f.redact(apiErrorText(frame.Error)))
	case errors.As(err, &syntaxErr):
		return newError(CodeProtocol, PhaseStream, msgInvalidJSON)
	case errors.As(err, &typeErr):
		return newError(CodeProtocol, PhaseStream, msgUndecodable)
	default:
		return newError(CodeTransport, PhaseStream, msgConnectionLost)
	}
}

// upstream is a failure the provider reported inside the stream.
func (f responsesFailures) upstream(msg string) *Error {
	return newError(CodeUpstreamError, PhaseStream, f.redact(msg))
}

// prefix is pi's provider label in HTTP error messages.
func (f responsesFailures) prefix() string {
	if f.provider == ProviderOpenAI {
		return "OpenAI API error"
	}
	return string(f.provider) + " API error"
}

// keyLike matches OpenAI-style API keys, masked ones included
// ("sk-proj-****abcd"): vendors echo them in auth errors. The word boundary
// keeps ordinary words such as "task-runner" intact.
var keyLike = regexp.MustCompile(`\bsk-[A-Za-z0-9_*\-]{3,}`)

// redact removes the call's key and anything key-shaped from provider text.
// pi passes such text through; this is a recorded security difference (spec
// I9: secrets never reach messages or errors).
func (f responsesFailures) redact(text string) string {
	text = strings.ReplaceAll(text, f.apiKey.reveal(), "[REDACTED]")
	return keyLike.ReplaceAllString(text, "[REDACTED]")
}

// describeHTTPError ports how pi-ai describes a non-2xx response: openai-node
// builds the APIError message from the body's "error" field (or the raw text
// of a non-JSON body), and pi's formatProviderError shows a non-empty error
// object as the body instead.
func describeHTTPError(prefix string, status int, body []byte) string {
	var parsed any
	isJSON := json.Unmarshal(body, &parsed) == nil
	var errRaw json.RawMessage
	var errVal any
	if obj, ok := parsed.(map[string]any); ok {
		errVal = obj["error"]
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(body, &fields)
		errRaw = fields["error"]
	}
	if obj, ok := errVal.(map[string]any); ok && len(obj) > 0 {
		return fmt.Sprintf("%s (%d): %s", prefix, status, truncateUTF16(compactJSON(errRaw), maxErrorBodyChars))
	}
	var msg string
	switch {
	case jsTruthy(errVal):
		msg = compactJSON(errRaw)
	case !isJSON || !jsTruthy(parsed):
		msg = string(body)
	}
	if msg == "" {
		return fmt.Sprintf("%s (%d): %d status code (no body)", prefix, status, status)
	}
	return fmt.Sprintf("%s (%d): %d %s", prefix, status, status, msg)
}

// apiErrorText is openai-node's APIError message for an error object without
// an HTTP status: its message, else the object as JSON.
func apiErrorText(raw json.RawMessage) string {
	var e struct {
		Message any `json:"message"`
	}
	if json.Unmarshal(raw, &e) == nil && jsTruthy(e.Message) {
		if s, ok := e.Message.(string); ok {
			return s
		}
		b, _ := marshalJS(e.Message)
		return string(b)
	}
	return compactJSON(raw)
}

// jsText renders a string field as a JavaScript template literal would:
// "undefined" when absent and "null" when null.
func jsText(field respjson.Field, value string) string {
	switch raw := field.Raw(); {
	case field.Valid():
		return value
	case raw == respjson.Omitted:
		return "undefined"
	default: // null, or a value of another JSON type
		return raw
	}
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

// truncateUTF16 is pi's truncateErrorText, which counts JavaScript string
// (UTF-16) units.
func truncateUTF16(s string, max int) string {
	units := utf16.Encode([]rune(s))
	if len(units) <= max {
		return s
	}
	return fmt.Sprintf("%s... [truncated %d chars]", string(utf16.Decode(units[:max])), len(units)-max)
}
