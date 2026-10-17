package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func TestMiniMaxChatFinalGuard(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, source := range []string{"model", "profile", "call", "hook", "repair", "authorization"} {
			for _, sc := range []struct {
				field, raw string
				valid      bool
			}{
				{"thinking", `{"type":"disabled"}`, model == "MiniMax-M3"}, {"thinking", `{"type":"enabled"}`, false}, {"thinking", `null`, false}, {"thinking", `{"type":"adaptive","budget_tokens":12}`, false},
				{"reasoning_effort", `"none"`, false}, {"reasoning_effort", `"minimal"`, false}, {"reasoning_effort", `"low"`, model != "MiniMax-M3"}, {"reasoning_effort", `true`, false},
				{"reasoning_split", `false`, false}, {"reasoning_split", `null`, false}, {"reasoning_split", `true`, true},
				{"max_tokens", `1`, false}, {"max_completion_tokens", `1`, true}, {"max_completion_tokens", `131072`, true}, {"max_completion_tokens", `131073`, false}, {"max_completion_tokens", `0`, false}, {"max_completion_tokens", `null`, false}, {"max_completion_tokens", `1.00000000000000001`, false}, {"max_completion_tokens", `1e0`, true}, {"max_completion_tokens", `1e1000000000`, false},
				{"temperature", `2`, true}, {"temperature", `2.1`, false}, {"temperature", `null`, true}, {"temperature", `true`, false}, {"temperature", `-0.1`, false},
				{"top_p", `0`, true}, {"top_p", `1.1`, false}, {"top_p", `null`, false},
				{"tool_choice", `"auto"`, true}, {"tool_choice", `"none"`, false}, {"parallel_tool_calls", `false`, false},
				{"stream_options", `{"include_usage":false}`, false}, {"tool_stream", `true`, false}, {"prompt_cache_retention", `"24h"`, false}, {"service_tier", `"priority"`, false}, {"service_tier", `"standard"`, true},
				{"stop", `["END"]`, false}, {"response_format", `{"type":"json_object"}`, false}, {"seed", `1`, false}, {"n", `2`, false}, {"n", `1`, true},
			} {
				t.Run(model+"/"+source+"/"+sc.field+"-"+sc.raw, func(t *testing.T) {
					ev := run.Case(t, "MM02-guard-"+t.Name())
					params := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
					w := newMiniMaxChatWorld(t, func(cat *ai.Catalog) {
						for i := range cat.Models {
							m := &cat.Models[i]
							if m.ID != model {
								continue
							}
							if source == "model" {
								m.SamplingParams = params
							}
							if source == "profile" {
								m.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{m.DefaultThinkingLevel: params}
							}
						}
					})
					enqueue(ev, w, miniMaxChatReply(t, model))
					opts := ai.ChatOptions{}
					hooks := ai.Hooks{}
					if source == "call" || source == "repair" {
						opts.SamplingParams = params
					}
					if source == "hook" || source == "repair" || source == "authorization" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							var v any
							dec := json.NewDecoder(strings.NewReader(sc.raw))
							dec.UseNumber()
							if err := dec.Decode(&v); err != nil {
								t.Fatal(err)
							}
							p.Body[sc.field] = v
							if source == "repair" {
								delete(p.Body, sc.field)
								p.Body["reasoning_split"] = true
								p.Body["stream_options"] = map[string]any{"include_usage": true}
							}
							if source == "authorization" {
								p.Body["model"] = "unauthorized"
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget(model), miniMaxChatRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if source != "authorization" && (sc.valid || source == "repair") {
						ev.Check("legal final body sends", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
						return
					}
					code := ai.CodeInvalidRequest
					if source == "authorization" {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, ai.PhaseRequest)
					ev.Check("zero attempts and provider calls", len(o.result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "%v", o.err)
				})
			}
		}
	}
}
