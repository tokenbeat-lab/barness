package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// geminiAdapter implements the Gemini Developer API (generateContent,
// streamed as SSE). It speaks HTTP directly instead of through the Google
// GenAI Go SDK, which spec I5 allows for a combination whose required
// fields the SDK cannot carry losslessly (ADR-0012): the SDK decodes a
// thoughtSignature into bytes and a function call's arguments into a map,
// exposes no raw response, and fails on SSE lines pi's SDK skips. Its
// client constructor also reads GOOGLE_API_KEY, GEMINI_API_KEY and
// GOOGLE_GEMINI_BASE_URL, which barness never does.
//
// The stream is decoded as @google/genai, which pi uses, decodes it, and
// read as pi reads it: only a finish reason ends the call cleanly. Like pi,
// the adapter never runs the response callback; and the baseline's refusal
// of a custom fetch has nothing to refuse, as barness has no per-call
// transport.
type geminiAdapter struct{}

// msgGeminiFetchFailed is the text pi reports for a request that got no
// response: Node's fetch (undici) fails with it and the Google SDK passes
// it on unchanged.
const msgGeminiFetchFailed = "fetch failed"

// headers are the API key, barness-ai's User-Agent and the body's media
// type, as pi passes headers to the Google SDK.
func (geminiAdapter) headers(ac adapterCall) http.Header {
	h := http.Header{}
	h.Set("X-Goog-Api-Key", ac.apiKey.reveal())
	h.Set("User-Agent", userAgent)
	h.Set("Content-Type", "application/json")
	return h
}

func (geminiAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	opts, _ := ac.options.(GeminiOptions) // nil means protocol defaults
	body, failure := buildGeminiBody(ac.model, ac.history, opts)
	if failure != nil {
		return failure
	}
	// As in pi, the payload callback runs once per logical call, outside
	// any retry of the initial request. It sees the REST body (ADR-0012).
	body, failure = ac.hooks.payload(ctx, body, authorizeGeminiPayload(ac.model, ac.hostedTools))
	if failure != nil {
		return failure
	}
	if failure := ac.limits.checkRequestBody(body); failure != nil {
		return failure
	}

	target := geminiStreamURL(ac.endpoint, ac.model.ID)
	failures := httpFailures{apiKey: ac.apiKey, clock: ac.initial.clock, describe: geminiHTTPError, noResponse: msgGeminiFetchFailed}
	var res *http.Response
	failure = ac.initial.send(ctx, func(ctx context.Context) attemptOutcome {
		res = nil
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
		if err != nil {
			return attemptOutcome{failure: newError(CodeInvalidRequest, PhaseRequest, "request could not be built")}
		}
		for name, values := range ac.header {
			// A name without values was removed by the header transform;
			// present but empty, net/http neither writes it nor puts its
			// own User-Agent in its place.
			req.Header[name] = slices.Clone(values)
		}
		r, err := ac.http.Do(req)
		if err != nil {
			o := failures.request(ctx, err, nil)
			// pi retries only failures with an HTTP status (google-shared
			// retryGoogleRequest): a request that got no response is not.
			o.connection = false
			return o
		}
		ac.limits.limitBody(r)
		if r.StatusCode >= 300 {
			o := failures.status(r)
			closeBody(r)
			// pi's retry sees the status of a Google SDK error but none of
			// its headers: neither x-should-retry nor a requested delay.
			o.header = nil
			return o
		}
		res = r
		return attemptOutcome{status: r.StatusCode, header: r.Header}
	})
	if failure != nil {
		return failure
	}
	// Closing releases the response body on every later path.
	defer closeBody(res)

	out.start()
	p := &geminiParser{out: out, model: ac.model, attempt: ac.initial, failures: failures}
	return p.read(ctx, newGeminiDecoder(res.Body, failures.upstream))
}

// geminiStreamURL is where the Google SDK sends a streamed generateContent
// for pi: the model's base URL (which carries the API version) without a
// trailing slash, then models/{id}:streamGenerateContent?alt=sse.
func geminiStreamURL(endpoint, modelID string) string {
	return strings.TrimSuffix(endpoint, "/") + "/models/" + url.PathEscape(modelID) + ":streamGenerateContent?alt=sse"
}

// geminiHTTPError is the Google SDK's ApiError message for a non-2xx
// response, which pi reports as is: the JSON error body, compacted, or for
// any other body a JSON error object holding its text, the status and the
// status text. An error body declared JSON that does not parse is described
// as text; the SDK fails on it with JavaScript's parse error (ADR-0012).
func geminiHTTPError(res *http.Response, body []byte) (msg, sdk string) {
	if strings.Contains(res.Header.Get("Content-Type"), "application/json") && json.Valid(body) {
		msg = compactJSON(body)
		return msg, msg
	}
	type errorObject struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
	}
	statusText := strings.TrimPrefix(res.Status, strconv.Itoa(res.StatusCode)+" ")
	wrapped, err := marshalJS(map[string]errorObject{"error": {Message: string(body), Code: res.StatusCode, Status: statusText}})
	if err != nil {
		wrapped = []byte(strconv.Itoa(res.StatusCode))
	}
	msg = string(wrapped)
	return msg, msg
}
