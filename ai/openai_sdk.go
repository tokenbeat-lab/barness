package ai

import (
	"net/http"
	"time"

	"github.com/openai/openai-go/v3/option"
)

// requestIDHeaderOf is the response header where provider p returns its
// own request id on the OpenAI protocols: OpenAI's x-request-id, or
// DeepSeek's x-ds-trace-id (DeepSeek sends no x-request-id; observed live
// 2026-10-02, maintainer decision 2026-10-03, issue 33). It is keyed by the
// provider like responsesCapabilitiesOf and chatCompatOf. Error.
// ProviderRequestID is a barness-ai extension; pi reads no such header.
func requestIDHeaderOf(p ProviderID) string {
	if p == ProviderDeepSeek {
		return "x-ds-trace-id"
	}
	return "x-request-id"
}

// openAIRequestOptions are the OpenAI Go SDK options both OpenAI protocols
// (Responses, Chat Completions) send their own body with: the Client's
// HTTP client and the binding's endpoint, no SDK retries, the policy's
// byte limits on every response body, openai-node's timeout announcement
// for an explicit timeoutMs, the call's exact headers without the SDK's host
// description, and delivered remembering the body of a response the SDK may
// drop.
func openAIRequestOptions(ac adapterCall, body []byte, timeoutMs int, timeout time.Duration, delivered *deliveredBody) []option.RequestOption {
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
		// SDK's own reads of an error body included. The SDK's SSE decoder
		// also refuses a line over 32 MiB on its own; a MaxFrameBytes above
		// that cannot be reached and ends as a lost connection instead.
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			return ac.limits.limitBodies(r, next)
		}),
	}
	if timeoutMs > 0 {
		// openai-node announces only an explicit timeoutMs.
		announce := announceTimeout(timeout)
		reqOpts = append(reqOpts, option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			return announce(r, next)
		}))
	}
	reqOpts = append(reqOpts, option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		return delivered.keep(r, next)
	}))
	// Authentication travels in ac.header, so the SDK's own API key setting
	// stays empty and adds no second Authorization.
	reqOpts = append(reqOpts, headerOptions(ac.header)...)
	keep := keepRemoved(ac.header)
	return append(reqOpts, option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		return keep(r, next)
	}), option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		return withoutHostDescription(r, next)
	}))
}

// headerOptions sends header exactly: each name with its values in order,
// and a name with no values suppressed.
func headerOptions(header http.Header) []option.RequestOption {
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
