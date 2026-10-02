package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// anthropicAdapter implements the Anthropic Messages protocol. The Anthropic
// Go SDK sends the request: it applies the protocol's headers and reads an
// error response, with its own retries off. It deliberately avoids
// anthropic.NewClient, whose defaults read ANTHROPIC_API_KEY,
// ANTHROPIC_AUTH_TOKEN, ANTHROPIC_BASE_URL, ANTHROPIC_PROFILE,
// ANTHROPIC_CUSTOM_HEADERS and credential files from the environment; a
// fresh service is built per call from explicit options only.
//
// The event stream is decoded here, as pi-ai decodes it itself instead of
// through its SDK: by SSE event name, with JSON repair, an "error" event
// failing the call and only message_stop after message_start ending it
// cleanly.
type anthropicAdapter struct{}

// anthropicRequestIDHeader is where Anthropic returns its request id.
const anthropicRequestIDHeader = "request-id"

// Error texts barness-ai shares with pi-ai's Anthropic path.
const (
	// pi's check after the stream once its signal was aborted.
	msgAnthropicAborted        = "Request was aborted"
	msgAnthropicNoMessageStop  = "Anthropic stream ended before message_stop"
	msgAnthropicNoStopReason   = "Anthropic stream ended without a stop reason"
	msgAnthropicFallback       = "Anthropic performed an unsupported mid-output model fallback"
	msgAnthropicRefusalDefault = "The model refused to complete the request"
	msgAnthropicSensitive      = "Provider stopped with: sensitive"
)

// Error texts pi takes from its JavaScript runtime (JSON.parse, undici),
// which barness cannot reproduce; recorded as approved differences in the pi
// ledger.
const (
	msgAnthropicConnectionLost = "Connection lost while reading the Anthropic Messages stream"
)

// msgAnthropicInvalidJSON is pi's "Could not parse Anthropic SSE event …"
// without the JavaScript parser's message and the raw frame, which may
// repeat content.
func msgAnthropicInvalidJSON(event string) string {
	return "Could not parse Anthropic SSE event " + event + ": invalid JSON"
}

// headers are the API key, barness-ai's User-Agent, the browser-access
// header pi sends and the beta features the call needs. The header
// transform may change anthropic-beta, which then replaces those features,
// as a configured anthropic-beta header does in pi.
func (anthropicAdapter) headers(ac adapterCall) http.Header {
	h := http.Header{}
	h.Set("X-Api-Key", ac.apiKey.reveal())
	h.Set("User-Agent", userAgent)
	h.Set("Anthropic-Dangerous-Direct-Browser-Access", "true")
	opts, _ := ac.options.(AnthropicOptions) // nil means protocol defaults
	if betas := anthropicBetas(ac.model, opts); len(betas) > 0 {
		h.Set("Anthropic-Beta", strings.Join(betas, ","))
	}
	return h
}

func (anthropicAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	opts, _ := ac.options.(AnthropicOptions) // nil means protocol defaults
	header := ac.header.Clone()
	betas := normalizeBetas(header.Values("Anthropic-Beta"))
	// As in pi, a payload callback sees the beta features as the betas
	// parameter and may change them; they are sent as the header.
	hooked := ac.hooks.hooks.OnPayload != nil
	var bodyBetas []string
	if hooked {
		bodyBetas = betas
	}
	body, err := buildAnthropicBody(ac.model, ac.history, opts, bodyBetas)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	// As in pi, the payload callback runs once per logical call, outside
	// any retry of the initial request. pi forces stream back on after a
	// replacement; the SDK's streaming call always does.
	body, failure := ac.hooks.payload(ctx, body, authorizeAnthropicPayload(ac.model, ac.hostedTools))
	if failure != nil {
		return failure
	}
	beta := strings.Join(betas, ",")
	if hooked {
		if body, beta, failure = takeBetas(body); failure != nil {
			return failure
		}
	}
	if failure := ac.limits.checkRequestBody(body); failure != nil {
		return failure
	}
	if beta != "" {
		header.Set("Anthropic-Beta", beta)
	} else if _, ok := header["Anthropic-Beta"]; ok {
		header["Anthropic-Beta"] = nil
	}

	// pi announces the SDK's request timeout, its 10 minute default included.
	announce := announceTimeout(opts.requestTimeout())
	var delivered deliveredBody
	reqOpts := []option.RequestOption{
		option.WithHTTPClient(ac.http),
		option.WithBaseURL(ac.endpoint),
		// Retries are barness-ai's (ac.initial), never the SDK's.
		option.WithMaxRetries(0),
		option.WithRequestBody("application/json", body),
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			return markConnectionFailures(r, next)
		}),
		// Every response body is read through the policy's limits, the
		// SDK's own read of an error body included.
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			return ac.limits.limitBodies(r, next)
		}),
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			return announce(r, next)
		}),
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			return delivered.keep(r, next)
		}),
	}
	// Authentication travels in header, so the SDK's own API key setting
	// stays empty and adds no second credential.
	reqOpts = append(reqOpts, anthropicHeaderOptions(header)...)
	svc := anthropic.NewBetaMessageService(reqOpts...)
	failures := httpFailures{apiKey: ac.apiKey, clock: ac.initial.clock, requestIDHeader: anthropicRequestIDHeader, describe: anthropicHTTPError}
	ctx = withProtocolTimeout(ctx, opts.requestTimeout())
	var res *http.Response
	failure = ac.initial.send(ctx, func(ctx context.Context) attemptOutcome {
		res = nil
		// The request body is ours; the empty params are replaced by it.
		stream := svc.NewStreaming(ctx, anthropic.BetaMessageNewParams{}, option.WithResponseInto(&res))
		if err := stream.Err(); err != nil {
			// res is nil when the SDK dropped the response as ctx ended.
			o := failures.request(ctx, err, res)
			closeBody(res)
			delivered.close()
			return o
		}
		if res.StatusCode >= 300 {
			o := failures.status(res)
			closeBody(res)
			return o
		}
		return attemptOutcome{status: res.StatusCode, header: res.Header, providerRequestID: res.Header.Get(anthropicRequestIDHeader)}
	})
	if failure != nil {
		return failure
	}
	// Closing releases the response body on every later path.
	defer closeBody(res)
	requestID := res.Header.Get(anthropicRequestIDHeader)
	if failure := ac.hooks.response(ctx, res); failure != nil {
		failure.ProviderRequestID = requestID
		return failure
	}

	out.start()
	p := &anthropicParser{out: out, failures: failures, model: ac.model, attempt: ac.initial,
		open: map[int64][]anthropicSlot{}, usage: anthropicUsage{model: ac.model}}
	failure = p.read(ctx, newSSEDecoder(res.Body, ac.limits.frame))
	if failure != nil {
		failure.ProviderRequestID = requestID
	}
	return failure
}

// anthropicHeaderOptions sends header exactly: each name with its values in
// order, and a name with no values suppressed. It repeats Responses'
// headerOptions because the two SDKs' option types differ.
func anthropicHeaderOptions(header http.Header) []option.RequestOption {
	var opts []option.RequestOption
	for name, values := range header {
		if len(values) == 0 {
			opts = append(opts, option.WithHeaderDel(name))
			continue
		}
		opts = append(opts, option.WithHeader(name, values[0]))
		for _, v := range values[1:] {
			opts = append(opts, option.WithHeaderAdd(name, v))
		}
	}
	return opts
}

// takeBetas moves the betas parameter of a body a payload callback produced
// out of the body, as pi's SDK does: the header carries it as JavaScript's
// toString() would, the strings joined by commas as given (a configured
// anthropic-beta header is normalized, a callback's betas are not).
func takeBetas(body []byte) ([]byte, string, *Error) {
	var decoded map[string]any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&decoded); err != nil {
		return nil, "", newError(CodeCallbackFailed, PhaseRequest, "payload callback produced a body that cannot be encoded")
	}
	var betas string
	switch v := decoded["betas"].(type) {
	case nil:
	case string:
		betas = v
	case []any:
		parts := make([]string, len(v))
		for i, b := range v {
			s, ok := b.(string)
			if !ok {
				return nil, "", newError(CodeCallbackFailed, PhaseRequest, "payload callback produced malformed betas")
			}
			parts[i] = s
		}
		betas = strings.Join(parts, ",")
	default:
		return nil, "", newError(CodeCallbackFailed, PhaseRequest, "payload callback produced malformed betas")
	}
	delete(decoded, "betas")
	out, err := marshalJS(decoded)
	if err != nil {
		return nil, "", newError(CodeCallbackFailed, PhaseRequest, "payload callback produced a body that cannot be encoded")
	}
	return out, betas, nil
}

// anthropicHTTPError is the Anthropic TypeScript SDK's APIError message for
// a non-2xx response, which pi reports as is: the status and the parsed
// body's message, else the body as JSON, else its text.
func anthropicHTTPError(res *http.Response, body []byte) (msg, sdk string) {
	status := res.StatusCode
	var parsed any
	var text string
	if json.Unmarshal(body, &parsed) == nil && jsTruthy(parsed) {
		var fields struct {
			Message json.RawMessage `json:"message"`
		}
		_ = json.Unmarshal(body, &fields)
		var message any
		_ = json.Unmarshal(fields.Message, &message)
		switch m, isString := message.(string); {
		case isString && m != "":
			text = m
		case jsTruthy(message):
			text = compactJSON(fields.Message)
		default:
			text = compactJSON(body)
		}
	} else {
		text = string(body)
	}
	if text == "" {
		msg = fmt.Sprintf("%d status code (no body)", status)
	} else {
		msg = fmt.Sprintf("%d %s", status, text)
	}
	return msg, msg
}
