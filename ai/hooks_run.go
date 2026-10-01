package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
)

// boundHooks are one call's Hooks bound to its scope and authorized model.
// The adapter invokes payload and response at its protocol's points; the
// core invokes headers before dispatching to the adapter.
type boundHooks struct {
	hooks Hooks
	scope CallScope
	model Model
}

// guardedHeaders are the headers that carry authentication, the target,
// the vendor account/project or the tenant-scoped cache identity. The list
// is deliberately wider than the protocols implemented so far: an alternate
// credential header is honored by some gateway or vendor even where the
// current adapter never sends it. A header transform must leave each exactly as the adapter produced
// it — present with the same values, or absent — so it cannot replace the
// binding's credential, add a second one, redirect the request or act as
// another account.
var guardedHeaders = []string{
	"Authorization", "Proxy-Authorization", "Cookie", "Host",
	"X-Api-Key", "Api-Key", "X-Goog-Api-Key", "Cf-Aig-Authorization",
	"Openai-Organization", "Openai-Project", "X-Goog-User-Project",
	"Session_id", "X-Client-Request-Id", "X-Session-Id",
}

const redactedHeaderValue = "[REDACTED]"

// headers runs TransformHeaders over the adapter's merged headers. A name
// the transform removed maps to no values in the result, so the adapter
// suppresses it rather than letting an SDK default take its place.
func (b boundHooks) headers(ctx context.Context, merged http.Header) (http.Header, *Error) {
	if b.hooks.TransformHeaders == nil {
		return merged, nil
	}
	view := merged.Clone()
	for _, name := range guardedHeaders {
		if vs := view.Values(name); len(vs) > 0 {
			masked := make([]string, len(vs))
			for i := range masked {
				masked[i] = redactedHeaderValue
			}
			view[name] = masked
		}
	}
	err := b.hooks.TransformHeaders(ctx, b.scope, view)
	if failure := callbackFailure(ctx, err, "header transform failed"); failure != nil {
		return nil, failure
	}
	// The transform may have written non-canonical keys directly.
	final := http.Header{}
	for k, vs := range view {
		ck := http.CanonicalHeaderKey(k)
		final[ck] = append(final[ck], vs...)
	}
	for _, name := range guardedHeaders {
		want := merged.Values(name)
		if len(want) > 0 {
			want = slices.Repeat([]string{redactedHeaderValue}, len(want))
		}
		if !slices.Equal(final.Values(name), want) {
			return nil, newError(CodeTenantDenied, PhaseRequest, "header transform may not change the "+name+" header")
		}
		if len(want) > 0 {
			final[name] = merged.Values(name)
		}
	}
	for k := range merged {
		if _, kept := final[k]; !kept {
			final[k] = nil
		}
	}
	return final, nil
}

// payload runs OnPayload over the native body and returns what to send.
// authorize is the adapter's check that the final body still names only the
// authorized model and no native reference the library cannot authorize; it
// returns the refusal or "".
func (b boundHooks) payload(ctx context.Context, body []byte, authorize func(map[string]any) string) ([]byte, *Error) {
	if b.hooks.OnPayload == nil {
		return body, nil
	}
	var decoded map[string]any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&decoded); err != nil {
		return nil, newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	p := &Payload{ProviderID: b.model.Provider, API: b.model.API, ModelID: b.model.ID, Body: decoded}
	decision, err := b.hooks.OnPayload(ctx, b.scope, p)
	if failure := callbackFailure(ctx, err, "payload callback failed"); failure != nil {
		return nil, failure
	}
	final := p.Body
	if decision.replace {
		final = decision.body
	}
	if final == nil {
		return nil, newError(CodeInvalidRequest, PhaseRequest, "payload callback produced no request body")
	}
	if problem := authorize(final); problem != "" {
		return nil, newError(CodeTenantDenied, PhaseRequest, problem)
	}
	out, err := marshalJS(final)
	if err != nil {
		return nil, newError(CodeInvalidRequest, PhaseRequest, "payload callback produced a body that cannot be encoded")
	}
	return out, nil
}

// response runs OnResponse with the initial response's metadata.
func (b boundHooks) response(ctx context.Context, res *http.Response) *Error {
	if b.hooks.OnResponse == nil {
		return nil
	}
	info := ResponseInfo{
		Status:     res.StatusCode,
		Header:     res.Header.Clone(),
		ProviderID: b.model.Provider,
		API:        b.model.API,
		ModelID:    b.model.ID,
	}
	err := b.hooks.OnResponse(ctx, b.scope, info)
	return callbackFailure(ctx, err, "response callback failed")
}

// callbackFailure classifies a callback's outcome. An ended context wins
// over the callback's own result, so a callback that ignores cancellation
// still does not let the call proceed. The callback's error text never
// reaches the message; the error itself stays reachable through
// errors.Is/As on the returned *Error for the host's own diagnosis.
func callbackFailure(ctx context.Context, err error, msg string) *Error {
	if failure := contextError(ctx, PhaseRequest); failure != nil {
		failure.cause = err
		return failure
	}
	if err != nil {
		failure := newError(CodeInvalidRequest, PhaseRequest, msg)
		failure.cause = err
		return failure
	}
	return nil
}
