package ai

import (
	"context"
	"net/http"
)

// Hooks are a trusted host's request callbacks for one logical call. They are
// Go functions, kept apart from the deserializable Request and Options, so
// cloud JSON can never carry them. Each is optional and there is at most one
// of each per call; there is no handler registry.
//
// Every callback runs synchronously on the call's path, receives the call's
// context and its scope explicitly, and must honor the context. A callback
// error fails the call with CodeCallbackFailed in PhaseRequest, and a context
// that ended while it ran fails it as canceled or deadline_exceeded. The
// callback's error text is never put in the message, so host internals
// cannot reach it; errors.Is/As on the call's error still reach the
// callback's own error for the host's diagnosis.
//
// Unlike an observer, a callback can fail the call, but it cannot widen it:
// the final authentication, target, model and native references must still
// be the ones the binding authorized.
type Hooks struct {
	// TransformHeaders edits the request headers in place once the
	// authentication and request headers are merged, before the call is
	// handed to the protocol adapter. Credential values are shown as
	// [REDACTED]; authentication, target and account headers must be left
	// as they are.
	TransformHeaders func(ctx context.Context, scope CallScope, header http.Header) error
	// OnPayload runs once per logical call after the adapter built the
	// native request body, outside any retry of the initial request. It may
	// observe the body, change payload.Body in place, or return a
	// replacement (see PayloadDecision).
	OnPayload func(ctx context.Context, scope CallScope, payload *Payload) (PayloadDecision, error)
	// OnResponse runs once, after the initial response was received
	// successfully and before the start event. It reads HTTP metadata only;
	// it never sees or replaces the final message.
	OnResponse func(ctx context.Context, scope CallScope, response ResponseInfo) error
}

// Payload is the native request body the adapter built, for OnPayload.
type Payload struct {
	ProviderID ProviderID
	API        API
	ModelID    string
	// Body is the decoded JSON object; numbers are json.Number so they are
	// sent as received. Changes made in place are sent unless replaced.
	Body map[string]any
}

// PayloadDecision is what OnPayload decided. The zero value, like
// KeepPayload, sends Payload.Body including any in-place changes;
// ReplacePayload sends another body instead.
type PayloadDecision struct {
	replace bool
	body    map[string]any
}

// KeepPayload sends the payload's Body, with whatever was changed in place.
func KeepPayload() PayloadDecision { return PayloadDecision{} }

// ReplacePayload sends body instead of the payload's Body. A nil body fails
// the call; it never means "keep".
func ReplacePayload(body map[string]any) PayloadDecision {
	return PayloadDecision{replace: true, body: body}
}

// ResponseInfo is the initial HTTP response's metadata, for OnResponse.
type ResponseInfo struct {
	Status int
	// Header is a copy; changing it has no effect.
	Header     http.Header
	ProviderID ProviderID
	API        API
	ModelID    string
}

// HookedClient runs calls with a fixed set of Hooks. It shares its Client's
// configuration, transport and resolvers; every call it starts pins the
// Hooks it was created with, so concurrent calls of different tenants each
// use only their own.
type HookedClient struct {
	client *Client
	hooks  Hooks
}

// WithHooks returns a HookedClient whose calls run with hooks. Networking is
// not part of Hooks: the transport is only ever assembled in Config.
func (c *Client) WithHooks(hooks Hooks) *HookedClient {
	return &HookedClient{client: c, hooks: hooks}
}

// Stream is Client.Stream with the hooks.
func (h *HookedClient) Stream(ctx context.Context, scope CallScope, target Target, req Request, opts Options) *Stream {
	return h.client.startStream(ctx, newCall(scope, target, req, opts, nil, h.hooks))
}

// StreamSimple is Client.StreamSimple with the hooks.
func (h *HookedClient) StreamSimple(ctx context.Context, scope CallScope, target Target, req Request, opts SimpleOptions) *Stream {
	return h.client.startStream(ctx, newCall(scope, target, req, nil, &opts, h.hooks))
}

// Complete is Client.Complete with the hooks.
func (h *HookedClient) Complete(ctx context.Context, scope CallScope, target Target, req Request, opts Options) (Result, error) {
	return h.client.run(ctx, newCall(scope, target, req, opts, nil, h.hooks), nil)
}

// CompleteSimple is Client.CompleteSimple with the hooks.
func (h *HookedClient) CompleteSimple(ctx context.Context, scope CallScope, target Target, req Request, opts SimpleOptions) (Result, error) {
	return h.client.run(ctx, newCall(scope, target, req, nil, &opts, h.hooks), nil)
}
