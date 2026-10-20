// Package glmhost demonstrates three explicit standard GLM routes for the
// chosen China or international service.
// The Host owns authentication, versioned resolvers, key purpose, policy and
// tool execution. No key is read here and no fallback route is selected.
package glmhost

import (
	"context"
	"errors"

	"github.com/tokenbeat-lab/barness/ai"
)

// Entry chooses an existing public Client entry; every call is one turn.
type Entry string

const (
	FullStream     Entry = "full-stream"
	FullComplete   Entry = "full-complete"
	SimpleStream   Entry = "simple-stream"
	SimpleComplete Entry = "simple-complete"
)

var ErrEntry = errors.New("glmhost: unknown entry or protocol")

// ObservedTurn retains event attribution for this small example. A UI should
// consume envelopes incrementally instead of accumulating a long answer.
type ObservedTurn struct {
	Result    ai.Result
	Events    []ai.Event
	Envelopes []ai.EventEnvelope
}

// Turn uses explicit full options or neutral simple options, preserving the
// server's default thinking. The caller supplies an authenticated Scope with
// a fresh RequestID and only trusted hooks. It must not log content/Results to
// shared operational logs. Cancellation is passed through to the Client.
func Turn(ctx context.Context, client *ai.Client, scope ai.CallScope, target ai.Target, req ai.Request, api ai.API, entry Entry, hooks ai.Hooks) (ObservedTurn, error) {
	var opts ai.Options
	switch api {
	case ai.APIOpenAICompletions:
		opts = ai.ChatOptions{MaxTokens: ai.Value(4096)}
	case ai.APIOpenAIResponses:
		opts = ai.ResponsesOptions{MaxTokens: ai.Value(4096)}
	case ai.APIAnthropicMessages:
		opts = ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
	default:
		return ObservedTurn{}, ErrEntry
	}
	if entry != FullStream && entry != FullComplete && entry != SimpleStream && entry != SimpleComplete {
		return ObservedTurn{}, ErrEntry
	}
	c := client.WithHooks(hooks)
	var out ObservedTurn
	var err error
	simple := ai.SimpleOptions{MaxTokens: ai.Value(4096)}
	switch entry {
	case FullComplete:
		out.Result, err = c.Complete(ctx, scope, target, req, opts)
		return out, err
	case SimpleComplete:
		out.Result, err = c.CompleteSimple(ctx, scope, target, req, simple)
		return out, err
	}
	var stream *ai.Stream
	if entry == FullStream {
		stream = c.Stream(ctx, scope, target, req, opts)
	} else {
		stream = c.StreamSimple(ctx, scope, target, req, simple)
	}
	defer stream.Close()
	for stream.Next() {
		out.Events = append(out.Events, stream.Event())
		out.Envelopes = append(out.Envelopes, stream.Envelope())
	}
	out.Result, err = stream.Result()
	return out, err
}
