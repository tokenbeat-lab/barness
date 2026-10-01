package ai

import (
	"context"
	"net/http"
)

// adapter converts one call to a wire protocol and reports normalized progress
// to the assembler. It must reach a protocol terminal (setting the message's
// StopReason) or return the classified failure.
type adapter interface {
	// simpleOptions maps protocol-neutral options, already resolved for m
	// (SimpleOptions.resolve), onto this protocol.
	simpleOptions(m Model, o SimpleOptions) Options
	// historyRules are this protocol's inputs to the history transform for m.
	historyRules(m Model) historyRules
	// headers are the authentication and protocol request headers for the
	// call, merged; the header transform runs over them before stream.
	headers(call adapterCall) http.Header
	// stream sends call.header exactly and runs call.hooks' payload and
	// response callbacks at this protocol's points.
	stream(ctx context.Context, call adapterCall, out *assembler) *Error
}

// adapterCall is everything one call's adapter may use: an immutable snapshot
// of the resolved binding, credential and the history prepared for the
// model. Nothing in it is shared mutable tenant state.
type adapterCall struct {
	http     *http.Client
	endpoint string
	apiKey   Secret
	model    Model
	history  transcript
	options  Options // matches the adapter's API; nil means protocol defaults
	// cache scopes the cache and affinity identifiers the adapter derives.
	cache cacheScope
	// hostedTools are the binding's AllowedHostedTools.
	hostedTools []string
	// header is the final request header set. A name with no values was
	// removed by the header transform and must not be sent, not even as an
	// SDK default.
	header http.Header
	hooks  boundHooks
}

// registry is the explicit adapter table selected by Binding.API. It is built
// once per Client and never modified afterwards.
func registry() map[API]adapter {
	return map[API]adapter{
		APIOpenAIResponses: responsesAdapter{},
	}
}
