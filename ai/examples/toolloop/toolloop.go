// Package toolloop is barness-ai's minimal tool round trip (spec I6, User
// Stories 15–17, 49): the host, not the library, executes tools and starts
// the next turn.
//
// One generation turn is one logical call. When a turn ends with
// StopReasonToolUse, the host validates every call with ai.ValidateToolCall,
// runs the valid ones itself, answers each call with a ToolResultMessage and
// starts the next turn as a new logical call with a new RequestID. A turn
// that ended any other way — length, error, aborted — is never executed,
// however complete its calls look. The loop is bounded by MaxTurns.
package toolloop

import (
	"context"
	"errors"
	"fmt"

	"github.com/tokenbeat-lab/barness/ai"
)

// Tool is a function the host offers the model and runs itself.
type Tool struct {
	Declaration ai.Tool
	// Run executes a validated call. Its text answers the call; an error
	// answers it as a failed execution (isError) with the error's text, and
	// the loop goes on so the model can react.
	Run func(ctx context.Context, call ai.ValidToolCall) (string, error)
}

// Loop runs turns until the model answers without tools.
type Loop struct {
	Client *ai.Client
	Target ai.Target
	// Scope returns the trusted scope of each new logical call: the same
	// tenant and actor, a fresh RequestID every time.
	Scope func() ai.CallScope
	// MaxTurns bounds the turns of one Run; a model that keeps calling
	// tools past it ends with ErrTooManyTurns.
	MaxTurns int
}

// ErrTooManyTurns ends a Run whose model still asked for tools after
// MaxTurns turns.
var ErrTooManyTurns = errors.New("toolloop: too many turns")

// Turn is one logical call of a Run.
type Turn struct {
	Result ai.Result
	// Results are the host's answers to the turn's tool calls, in order.
	Results []ai.ToolResultMessage
}

// Run starts from req and returns every turn. The last turn's message is
// the model's final answer, or the failure that ended the run with err.
func (l Loop) Run(ctx context.Context, req ai.Request, tools []Tool) ([]Turn, error) {
	declarations := make([]ai.Tool, len(tools))
	byName := make(map[string]Tool, len(tools))
	for i, t := range tools {
		if _, dup := byName[t.Declaration.Name]; dup || t.Run == nil {
			return nil, fmt.Errorf("toolloop: tool %q is declared twice or has no Run", t.Declaration.Name)
		}
		declarations[i] = t.Declaration
		byName[t.Declaration.Name] = t
	}
	req.Tools = declarations
	history := append([]ai.Message(nil), req.Messages...)
	var turns []Turn
	for range l.MaxTurns {
		req.Messages = history
		res, err := l.Client.Complete(ctx, l.Scope(), l.Target, req, nil)
		turn := Turn{Result: res}
		if err != nil || res.Message.StopReason != ai.StopReasonToolUse {
			return append(turns, turn), err
		}
		history = append(history, res.Message)
		for i, c := range res.Message.Content {
			call, ok := c.(ai.ToolCall)
			if !ok {
				continue
			}
			result := execute(ctx, res.Message, i, call, declarations, byName)
			turn.Results = append(turn.Results, result)
			history = append(history, result)
		}
		turns = append(turns, turn)
	}
	return turns, fmt.Errorf("%w: %d", ErrTooManyTurns, l.MaxTurns)
}

// execute answers one call: refused calls and failed runs become isError
// results the model sees; only a validated call is run.
func execute(ctx context.Context, msg ai.AssistantMessage, index int, call ai.ToolCall, declarations []ai.Tool, byName map[string]Tool) ai.ToolResultMessage {
	valid, err := ai.ValidateToolCall(msg, index, declarations)
	if err != nil {
		// pi-ai's wording, written for the model to correct its call.
		text := err.Error()
		var refused *ai.ToolCallError
		if errors.As(err, &refused) {
			text = refused.Message
		}
		return ai.ToolResultText(call, text, true)
	}
	out, err := byName[valid.Name()].Run(ctx, valid)
	if err != nil {
		return ai.ToolResultText(call, err.Error(), true)
	}
	return ai.ToolResultText(call, out, false)
}
