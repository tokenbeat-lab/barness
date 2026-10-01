package ai

import (
	"context"
	"net/http"
)

// adapter converts one call to a wire protocol and reports normalized progress
// to the assembler. It must reach a protocol terminal (setting the message's
// StopReason) or return the classified failure.
type adapter interface {
	// simpleOptions maps protocol-neutral options onto this protocol for model.
	simpleOptions(m Model, o SimpleOptions) Options
	stream(ctx context.Context, call adapterCall, out *assembler) *Error
}

// adapterCall is everything one call's adapter may use: an immutable snapshot
// of the resolved binding, credential and request. Nothing in it is shared
// mutable tenant state.
type adapterCall struct {
	http     *http.Client
	endpoint string
	apiKey   Secret
	model    Model
	request  Request
	options  Options // matches the adapter's API; nil means protocol defaults
}

// registry is the explicit adapter table selected by Binding.API. It is built
// once per Client and never modified afterwards.
func registry() map[API]adapter {
	return map[API]adapter{
		APIOpenAIResponses: responsesAdapter{},
	}
}
