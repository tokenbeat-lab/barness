package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"unicode/utf16"

	"github.com/openai/openai-go/v3/packages/respjson"
	"github.com/openai/openai-go/v3/packages/ssestream"

	"github.com/tokenbeat-lab/barness/ai/internal/clock"
)

// Error texts barness-ai shares with pi-ai's Responses path, so a caller sees
// the same errorMessage for the same failure.
const (
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

// responsesFailures classifies one Responses call's failures: the shared
// initial-request classification, plus the protocol's stream errors.
type responsesFailures struct{ httpFailures }

func newResponsesFailures(provider ProviderID, apiKey Secret, c *clock.Clock) responsesFailures {
	prefix := string(provider) + " API error"
	if provider == ProviderOpenAI {
		prefix = "OpenAI API error"
	}
	return responsesFailures{httpFailures{apiKey: apiKey, clock: c, requestIDHeader: openAIRequestIDHeader,
		describe: func(res *http.Response, body []byte) (string, string) {
			return describeHTTPError(prefix, res.StatusCode, body)
		}}}
}

// stream classifies an error that ended the SSE stream early.
func (f responsesFailures) stream(err error) *Error {
	var streamErr *ssestream.StreamError
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var limit *limitExceeded
	var expired *timeLimitExpired
	switch {
	case errors.As(err, &limit):
		return limit.failure(PhaseStream)
	case errors.As(err, &expired):
		return expired.failure(PhaseStream)
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

// describeHTTPError ports how pi-ai describes a non-2xx response: openai-node
// builds the APIError message (sdk) from the body's "error" field (or the
// raw text of a non-JSON body), and pi's formatProviderError shows a
// non-empty error object as the body instead (msg).
func describeHTTPError(prefix string, status int, body []byte) (msg, sdk string) {
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
	// openai-node's APIError.makeMessage.
	obj, isObject := errVal.(map[string]any)
	var text string
	switch {
	case isObject:
		text = apiErrorText(errRaw)
	case jsTruthy(errVal):
		text = compactJSON(errRaw)
	case !isJSON || !jsTruthy(parsed):
		text = string(body)
	}
	sdk = fmt.Sprintf("%d %s", status, text)
	if text == "" {
		sdk = fmt.Sprintf("%d status code (no body)", status)
	}
	if len(obj) > 0 {
		return fmt.Sprintf("%s (%d): %s", prefix, status, truncateUTF16(compactJSON(errRaw), maxErrorBodyChars)), sdk
	}
	return fmt.Sprintf("%s (%d): %s", prefix, status, sdk), sdk
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

// truncateUTF16 is pi's truncateErrorText, which counts JavaScript string
// (UTF-16) units.
func truncateUTF16(s string, max int) string {
	units := utf16.Encode([]rune(s))
	if len(units) <= max {
		return s
	}
	return fmt.Sprintf("%s... [truncated %d chars]", string(utf16.Decode(units[:max])), len(units)-max)
}
