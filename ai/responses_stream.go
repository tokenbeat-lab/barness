package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
)

// readResponsesStream feeds the body to p until failure or EOF, including
// events after a protocol terminal. As in pi, a terminal response event is
// success: EOF alone is not a terminal. Read the SDK's raw SSE decoder so
// named event:error can be handled before typed events, as in openai-node.
func readResponsesStream(ctx context.Context, dec ssestream.Decoder, p responsesParser) *Error {
	for dec.Next() {
		event := dec.Event()
		if bytes.HasPrefix(event.Data, []byte("[DONE]")) {
			break
		}
		if json.Valid(event.Data) {
			// The SDK wraps thread.* frames; pi's Responses adapter ignores
			// that wrapper rather than treating its data as response events.
			if strings.HasPrefix(event.Type, "thread.") {
				continue
			}
			if raw, reported := openAIStreamError(event.Type, event.Data); reported {
				return p.failures.upstream(apiErrorText(raw))
			}
		}
		var decoded responses.ResponseStreamEventUnion
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			return p.failures.stream(err)
		}
		if failure := p.handle(decoded); failure != nil {
			return failure
		}
		if failure := p.out.failure(); failure != nil {
			return failure
		}
	}
	if ctx.Err() != nil {
		msg := msgNoTerminal
		if p.terminal {
			msg = msgAbortedAfterTerminal
		}
		return interrupted(ctx, PhaseStream, msg)
	}
	if err := dec.Err(); err != nil {
		return p.failures.stream(err)
	}
	if p.failure != nil {
		return p.failure
	}
	if !p.terminal {
		return newError(CodeProtocol, PhaseStream, msgNoTerminal)
	}
	// pi checks completion after the terminal/EOF checks. Track content
	// blocks independently of slots: a repeated output_index can overwrite
	// the only slot that still points at an unfinished block.
	if p.stop == StopReasonToolUse && len(p.unfinished) > 0 {
		i := slices.Min(slices.Collect(maps.Keys(p.unfinished)))
		ref := p.unfinished[i]
		return newError(CodeProtocol, PhaseStream, p.failures.redact("OpenAI Responses stream completed with an unfinished tool call: "+ref.name+" ("+ref.id+")"))
	}
	return nil
}
