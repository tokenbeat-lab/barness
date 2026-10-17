package e2e

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxResponsesEffort(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, source := range []string{"full", "simple", "sampling", "hook"} {
			for _, level := range []string{"", "none", "minimal", "low", "medium", "high", "xhigh", "max"} {
				if (source == "full" || source == "simple") && level == "none" {
					continue
				}
				t.Run(model+"/"+source+"/"+level, func(t *testing.T) {
					ev := run.Case(t, "MM03-effort-"+t.Name())
					w := newMiniMaxResponsesWorld(t, nil)
					enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
					opts := ai.ResponsesOptions{}
					hooks := ai.Hooks{}
					switch source {
					case "full":
						opts.ReasoningEffort = ai.ThinkingLevel(level)
					case "sampling":
						if level != "" {
							opts.SamplingParams = map[string]json.RawMessage{"reasoning": mustMarshal(t, map[string]any{"effort": level})}
						}
					case "hook":
						if level != "" {
							hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
								p.Body["reasoning"] = map[string]any{"effort": level}
								return ai.KeepPayload(), nil
							}
						}
					}
					client := w.client.WithHooks(hooks)
					scope := textScope(t.Name())
					var o outcome
					if source == "simple" {
						o = drain(client.StreamSimple(ctxFor(t), scope, minimaxResponsesTarget(model), minimaxResponsesRequest(), ai.SimpleOptions{Reasoning: ai.ThinkingLevel(level)}))
					} else {
						o = drain(client.Stream(ctxFor(t), scope, minimaxResponsesTarget(model), minimaxResponsesRequest(), opts))
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
					valid := level == "" || source == "simple" || (model == "MiniMax-M3" && slices.Contains([]string{"high", "none"}, level)) || (model != "MiniMax-M3" && slices.Contains([]string{"low", "medium", "high", "xhigh", "max"}, level))
					if !valid {
						phase := ai.PhaseRequest
						if source == "full" {
							phase = ai.PhaseCapability
						}
						checkRejected(ev, o, ai.CodeInvalidRequest, phase)
						ev.Check("zero attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
						if source == "full" {
							ev.Check("unsupported full effort precedes credential", w.host.ReadsFor(scope.RequestID).Credentials == 0, "%+v", w.host.ReadsFor(scope.RequestID))
						}
						return
					}
					if !ev.Check("supported effort succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
						return
					}
					fields := bodyFields(t, w.provider.Requests()[0].Body)
					if level == "" {
						_, has := fields["reasoning"]
						ev.Check("omission retains service default", !has, "%s", w.provider.Requests()[0].Body)
						return
					}
					want := level
					if source == "simple" {
						if model == "MiniMax-M3" {
							want = "high"
						} else if level == "minimal" {
							want = "low"
						}
					}
					ev.Check("canonical effort is exact wire selection", jsonEqual(fields["reasoning"], mustMarshal(t, map[string]any{"effort": want})), "%s", w.provider.Requests()[0].Body)
				})
			}
		}
	}
}
