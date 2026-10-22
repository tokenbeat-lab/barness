package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiMessagesRequiredTools(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, mode := range []string{"no-tools", "hook-remove-tools", "plain-text", "empty-turn", "tool-call", "hook-any", "hook-auto", "length"} {
				t.Run(string(market)+"/"+entry+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K05-required-"+entry+"-"+mode)
					w := newKimiMessagesWorld(t, market, nil)
					req := miniMaxToolRequest()
					if mode == "no-tools" {
						req.Tools = nil
					}
					script := kimiMessagesScript(`{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}`)
					if mode == "tool-call" {
						script = string(miniMaxMessagesFixture(t, "tools.sse", "kimi-k3"))
					}
					if mode == "empty-turn" {
						frames := strings.Split(strings.TrimSpace(script), "\n\n")
						script = strings.Join(append(frames[:1], frames[3:]...), "\n\n") + "\n\n"
					}
					if mode == "length" {
						script = strings.ReplaceAll(script, `"stop_reason":"end_turn"`, `"stop_reason":"max_tokens"`)
					}
					ev.Fixture("synthetic.sse", []byte(script))
					enqueue(ev, w, provider.SSE([][]byte{[]byte(script)}))
					hook := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						// Simple entries obtain forced choice at the same final public hook seam.
						choice := "any"
						if mode == "hook-auto" {
							choice = "auto"
						}
						p.Body["tool_choice"] = map[string]any{"type": choice}
						if mode == "hook-remove-tools" {
							delete(p.Body, "tools")
						}
						return ai.KeepPayload(), nil
					}}
					client := w.client.WithHooks(hook)
					opts := ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "any"})}
					if mode == "hook-any" {
						opts.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "auto"})
					}
					var o outcome
					scope := textScope(t.Name())
					switch entry {
					case "stream":
						o = drain(client.Stream(ctxFor(t), scope, kimiMessagesTarget, req, opts))
					case "complete":
						o.result, o.err = client.Complete(ctxFor(t), scope, kimiMessagesTarget, req, opts)
					case "stream-simple":
						o = drain(client.StreamSimple(ctxFor(t), scope, kimiMessagesTarget, req, ai.SimpleOptions{}))
					case "complete-simple":
						o.result, o.err = client.CompleteSimple(ctxFor(t), scope, kimiMessagesTarget, req, ai.SimpleOptions{})
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					switch mode {
					case "no-tools", "hook-remove-tools":
						checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
						ev.Check("forced final payload without tools never sends", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
					case "tool-call", "hook-auto", "length":
						ev.Check("valid final choice completes", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
					default:
						ev.Check("forced normal turn requires a call", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
					}
				})
			}
		}
	}
}
