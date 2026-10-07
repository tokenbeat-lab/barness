package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf16"

	"github.com/openai/openai-go/v3/packages/respjson"

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
	return responsesFailures{httpFailures{apiKey: apiKey, clock: c, requestIDHeader: requestIDHeaderOf(provider),
		describe: func(res *http.Response, body []byte) (string, string) {
			return describeHTTPError(prefix, res.StatusCode, body)
		}}}
}

// stream classifies an error that ended the SSE stream early.
func (f responsesFailures) stream(err error) *Error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var limit *limitExceeded
	var expired *timeLimitExpired
	switch {
	case errors.As(err, &limit):
		return limit.failure(PhaseStream)
	case errors.As(err, &expired):
		return expired.failure(PhaseStream)
	case errors.As(err, &syntaxErr):
		return newError(CodeProtocol, PhaseStream, msgInvalidJSON)
	case errors.As(err, &typeErr):
		return newError(CodeProtocol, PhaseStream, msgUndecodable)
	default:
		return newError(CodeTransport, PhaseStream, msgConnectionLost)
	}
}

// describeHTTPError ports how pi-ai describes a non-2xx response: openai-node
// builds the APIError message (sdk) from its normalized error body (or the
// raw text of a non-JSON body). pi shows a non-empty error object only when
// the SDK message does not already carry it.
func describeHTTPError(prefix string, status int, body []byte) (msg, sdk string) {
	var parsed any
	isJSON := json.Unmarshal(body, &parsed) == nil
	errRaw := openAIHTTPErrorBody(body)
	var errVal any
	_ = json.Unmarshal(errRaw, &errVal)
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
		shown := truncateUTF16(compactJSON(errRaw), maxErrorBodyChars)
		if !strings.Contains(sdk, shown) {
			return fmt.Sprintf("%s (%d): %s", prefix, status, shown), sdk
		}
	}
	return fmt.Sprintf("%s (%d): %s", prefix, status, sdk), sdk
}

// openAIHTTPErrorBody is openai-node 7.19.0's makeStatusError normalization:
// an object/array with missing or null error becomes the SDK's error body.
// Keep the raw JSON so pi's property order and truncation remain unchanged.
func openAIHTTPErrorBody(body []byte) json.RawMessage {
	var value any
	if json.Unmarshal(body, &value) != nil {
		return nil
	}
	switch obj := value.(type) {
	case map[string]any:
		if obj["error"] == nil {
			return body
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(body, &fields)
		return fields["error"]
	case []any:
		return body
	default:
		return nil
	}
}

// openAIStreamError applies the Node SDK's SSE rule before either adapter:
// named error uses data.error ?? data; other frames fail on truthy error.
// Callers validate JSON first so malformed error frames stay protocol errors.
func openAIStreamError(eventType string, data []byte) (json.RawMessage, bool) {
	var frame struct {
		Error json.RawMessage `json:"error"`
	}
	_ = json.Unmarshal(data, &frame)
	if eventType == "error" {
		if len(frame.Error) == 0 || string(frame.Error) == "null" {
			return data, true
		}
		return frame.Error, true
	}
	return frame.Error, rawTruthy(frame.Error)
}

// apiErrorText is openai-node's APIError message for an error object without
// an HTTP status: its message, else the object as JSON.
func apiErrorText(raw json.RawMessage) string {
	if !rawTruthy(raw) {
		return "(no status code or body)"
	}
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
