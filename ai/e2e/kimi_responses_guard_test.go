package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func TestKimiResponsesFinalGuard(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, source := range []string{"model", "profile", "call", "hook", "repair", "authorize-first"} {
			for _, sc := range []struct {
				field, raw string
				valid      bool
			}{
				{"max_output_tokens", "1", true}, {"max_output_tokens", "131072", true}, {"max_output_tokens", "131073", false}, {"max_output_tokens", "0", false}, {"max_output_tokens", "null", false}, {"max_output_tokens", "1.00000000000000001", false}, {"max_output_tokens", "1e0", true}, {"max_tokens", "1", false},
				{"reasoning", `{"effort":"low"}`, true}, {"reasoning", `{"effort":"max"}`, true}, {"reasoning", `{"effort":"medium"}`, false}, {"reasoning", `{"effort":"high","summary":"auto"}`, false}, {"reasoning", `null`, false},
				{"temperature", "1", false}, {"temperature", "null", false}, {"top_p", "0.95", false}, {"thinking", `{"type":"disabled"}`, false}, {"include", `["reasoning.encrypted_content"]`, false}, {"service_tier", `"standard"`, false}, {"text", `{"format":{"type":"json_object"}}`, false},
				{"tool_choice", `"auto"`, true}, {"tool_choice", `"none"`, false}, {"tool_choice", `"required"`, false}, {"tool_choice", `null`, false}, {"tool_choice", `{"type":"function","name":"lookup"}`, false},
				{"prompt_cache_options", `{"mode":"implicit","ttl":"1h"}`, true}, {"prompt_cache_options", `{"mode":"explicit"}`, false}, {"prompt_cache_options", `{"mode":"implicit","ttl":"30m"}`, false}, {"prompt_cache_retention", `"24h"`, false}, {"untrusted_custom_field", `true`, false},
			} {
				t.Run(string(market)+"/"+source+"/"+sc.field+"-"+sc.raw, func(t *testing.T) {
					ev := run.Case(t, kimiResponsesID(market, "K04-final-")+t.Name())
					params := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
					w := newKimiResponsesWorld(t, market, func(c *ai.Catalog) {
						for i := range c.Models {
							m := &c.Models[i]
							if m.Provider != market || m.API != ai.APIOpenAIResponses {
								continue
							}
							if source == "model" {
								m.SamplingParams = params
							}
							if source == "profile" {
								m.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingMax: params}
							}
						}
					})
					enqueue(ev, w, kimiResponsesReply(t))
					opts := ai.ResponsesOptions{}
					hooks := ai.Hooks{}
					if source == "call" || source == "repair" {
						opts.SamplingParams = params
					}
					if source == "hook" || source == "repair" || source == "authorize-first" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							var v any
							d := json.NewDecoder(strings.NewReader(sc.raw))
							d.UseNumber()
							if err := d.Decode(&v); err != nil {
								t.Fatal(err)
							}
							p.Body[sc.field] = v
							if source == "repair" {
								delete(p.Body, sc.field)
								p.Body["prompt_cache_options"] = map[string]any{"mode": "implicit", "ttl": "5m"}
							}
							if source == "authorize-first" {
								p.Body["model"] = "outside-allowlist"
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if source != "authorize-first" && (sc.valid || source == "repair") {
						ev.Check("valid final body sends", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
						return
					}
					code := ai.CodeInvalidRequest
					if source == "authorize-first" {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, ai.PhaseRequest)
					ev.Check("zero Attempts and sends", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%v", o.err)
				})
			}
		}
	}
}
func TestKimiResponsesNamedOptions(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, sc := range []struct {
			name string
			opts ai.ResponsesOptions
		}{
			{"medium", ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMedium}}, {"minimal", ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMinimal}}, {"xhigh", ai.ResponsesOptions{ReasoningEffort: ai.ThinkingXHigh}}, {"zero-output", ai.ResponsesOptions{MaxTokens: ai.Value(0)}}, {"large-output", ai.ResponsesOptions{MaxTokens: ai.Value(131073)}}, {"temperature", ai.ResponsesOptions{Temperature: ai.Value(1.0)}}, {"none-cache", ai.ResponsesOptions{CacheRetention: ai.CacheRetentionNone}}, {"summary", ai.ResponsesOptions{ReasoningSummary: "auto"}}, {"tier", ai.ResponsesOptions{ServiceTier: ai.Value("standard")}}, {"none-choice", ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "none"})}}, {"required-choice", ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "required"})}}, {"named-tool", ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Function: "lookup"})}},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K04-named-")+sc.name)
				w := newKimiResponsesWorld(t, market, nil)
				scope := textScope(t.Name())
				o := drain(w.client.Stream(ctxFor(t), scope, kimiResponsesTarget(), kimiResponsesRequest(), sc.opts))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("no secret read", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0, "%+v", w.host.ReadsFor(scope.RequestID))
			})
		}
		for _, level := range []ai.ThinkingLevel{ai.ThinkingMinimal, ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMax} {
			t.Run(string(market)+"/simple-"+string(level), func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K04-clamp-")+string(level))
				w := newKimiResponsesWorld(t, market, nil)
				enqueue(ev, w, kimiResponsesReply(t))
				o := drain(w.client.StreamSimple(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), ai.SimpleOptions{Reasoning: level}))
				o.record(ev)
				if !ev.Check("simple succeeds", o.err == nil, "%v", o.err) {
					return
				}
				want := map[ai.ThinkingLevel]string{ai.ThinkingMinimal: "low", ai.ThinkingLow: "low", ai.ThinkingMedium: "high", ai.ThinkingHigh: "high", ai.ThinkingXHigh: "max", ai.ThinkingMax: "max"}[level]
				ev.Check("wire effort independently clamped", jsonEqual(bodyFields(t, w.provider.Requests()[0].Body)["reasoning"], json.RawMessage(`{"effort":"`+want+`"}`)), "%s", w.provider.Requests()[0].Body)
			})
		}
	}
}
