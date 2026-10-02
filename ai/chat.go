package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
)

// chatAdapter implements the OpenAI Chat Completions protocol on the OpenAI
// Go SDK's Chat service: the SDK sends the adapter's own request body with
// its retries off and decodes the SSE framing; the adapter reads each event
// as openai-node does for pi, since the SDK's typed stream stops reading at
// [DONE], fails on any chunk with an error field (null included) or a
// non-object chunk, and drops the reasoning fields and stream indexes pi
// depends on (ADR-0013). Like the Responses adapter it avoids
// openai.NewClient, whose defaults read OPENAI_API_KEY, OPENAI_BASE_URL,
// OPENAI_ORG_ID, OPENAI_PROJECT_ID and OPENAI_CUSTOM_HEADERS; a service is
// built per call from explicit options only.
//
// It speaks pi's openai-completions compat for the binding's provider
// (chatCompatOf): the standard OpenAI compat, or DeepSeek's. pi detects
// other vendors' quirks too; their Chat models are not served here.
type chatAdapter struct{}

// headers are the bearer authentication and barness-ai's User-Agent; pi
// sends session affinity headers on this protocol only to OpenRouter.
func (chatAdapter) headers(ac adapterCall) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+ac.apiKey.reveal())
	h.Set("User-Agent", userAgent)
	return h
}

func (chatAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	opts, _ := ac.options.(ChatOptions) // nil means protocol defaults
	cacheKey := opts.cacheKey(ac.cache, ac.endpoint, ac.model)
	body, err := buildChatBody(ac.model, ac.history, opts, cacheKey)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	// As in pi, the payload callback runs once per logical call, outside
	// any retry of the initial request. The SDK's streaming call sets
	// stream back to true after it, which pi's does not.
	body, failure := ac.hooks.payload(ctx, body, authorizeChatPayload(ac.model, cacheKey, ac.hostedTools))
	if failure != nil {
		return failure
	}
	if failure := ac.limits.checkRequestBody(body); failure != nil {
		return failure
	}
	var delivered deliveredBody
	reqOpts := openAIRequestOptions(ac, body, opts.TimeoutMs, opts.requestTimeout(), &delivered)
	svc := openai.NewChatCompletionService(reqOpts...)
	failures := newChatFailures(ac.apiKey, ac.initial)
	ctx = withProtocolTimeout(ctx, opts.requestTimeout())
	var stream *ssestream.Stream[openai.ChatCompletionChunk]
	var res *http.Response
	failure = ac.initial.send(ctx, func(ctx context.Context) attemptOutcome {
		res = nil
		// The request body is ours; the empty params are replaced by it.
		stream = svc.NewStreaming(ctx, openai.ChatCompletionNewParams{}, option.WithResponseInto(&res))
		if err := stream.Err(); err != nil {
			// The SDK has already closed the body of an error status, but
			// not of a response it dropped as ctx ended.
			_ = stream.Close()
			o := failures.request(ctx, err, res)
			delivered.close()
			return o
		}
		if res.StatusCode >= 300 {
			o := failures.status(res)
			_ = stream.Close()
			return o
		}
		return attemptOutcome{status: res.StatusCode, header: res.Header, providerRequestID: res.Header.Get(openAIRequestIDHeader)}
	})
	if failure != nil {
		return failure
	}
	// Close releases the response body on every later path. The typed
	// stream is never read: the adapter decodes the same body itself.
	defer stream.Close()
	requestID := res.Header.Get(openAIRequestIDHeader)
	if failure := ac.hooks.response(ctx, res); failure != nil {
		failure.ProviderRequestID = requestID
		return failure
	}

	out.start()
	p := newChatParser(out, ac.model, failures, ac.initial)
	failure = p.read(ctx, ssestream.NewDecoder(res))
	if failure != nil {
		failure.ProviderRequestID = requestID
	}
	return failure
}

// chatFailures classifies one Chat call's failures: the shared
// initial-request classification, with pi's Chat error texts, which carry
// no "API error" prefix, plus errors the provider reports in the stream.
type chatFailures struct{ httpFailures }

func newChatFailures(apiKey Secret, initial *initialRequest) chatFailures {
	return chatFailures{httpFailures{apiKey: apiKey, clock: initial.clock, requestIDHeader: openAIRequestIDHeader, describe: chatHTTPError}}
}

// inBand is a chunk's truthy error field: openai-node raises it as an
// APIError without a status, whose message pi reports, followed by
// OpenRouter's raw provider metadata when the message lacks it.
func (f chatFailures) inBand(raw json.RawMessage) *Error {
	return f.upstream(withRawMetadata(apiErrorText(raw), raw))
}

// chatHTTPError is how pi's Chat path describes a non-2xx response: pi's
// formatProviderError without a prefix over openai-node's APIError. A
// non-empty error object that the SDK's message does not already quote
// replaces it as "<status>: <object>" (cut to 4000 characters).
func chatHTTPError(res *http.Response, body []byte) (msg, sdk string) {
	_, sdk = describeHTTPError("", res.StatusCode, body)
	msg = sdk
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(body, &fields)
	errRaw := fields["error"]
	var obj map[string]any
	if json.Unmarshal(errRaw, &obj) == nil && len(obj) > 0 {
		if shown := truncateUTF16(compactJSON(errRaw), maxErrorBodyChars); !strings.Contains(sdk, shown) {
			msg = fmt.Sprintf("%d: %s", res.StatusCode, shown)
		}
	}
	return withRawMetadata(msg, errRaw), sdk
}

// withRawMetadata appends the error object's metadata.raw (OpenRouter's
// upstream provider error) to msg on its own line, unless msg already
// holds it. Only a string is appended; pi would append JavaScript's String()
// of any other truthy value.
func withRawMetadata(msg string, errRaw json.RawMessage) string {
	var e struct {
		Metadata struct {
			Raw json.RawMessage `json:"raw"`
		} `json:"metadata"`
	}
	if json.Unmarshal(errRaw, &e) != nil {
		return msg
	}
	if raw, ok := rawString(e.Metadata.Raw); ok && raw != "" && !strings.Contains(msg, raw) {
		return msg + "\n" + raw
	}
	return msg
}
