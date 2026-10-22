package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func TestKimiChatFinalGuard(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, source := range []string{"model", "profile", "call", "hook", "repair", "authorize-first"} {
			for _, sc := range []struct {
				field, raw string
				valid      bool
			}{
				{"temperature", "1", true}, {"temperature", "1.00000000000000001", false}, {"temperature", "null", false}, {"top_p", "0.95", true}, {"top_p", "0.95000000000000001", false}, {"n", "1", true}, {"n", "2", false}, {"presence_penalty", "0", true}, {"frequency_penalty", "0", true}, {"frequency_penalty", "0.00000000000000000001", false},
				{"max_completion_tokens", "131072", true}, {"max_completion_tokens", "131073", false}, {"max_completion_tokens", "0", false}, {"max_completion_tokens", "1e0", true}, {"max_completion_tokens", "1.00000000000000001", false}, {"max_completion_tokens", "null", false}, {"max_tokens", "1", false},
				{"reasoning_effort", `"low"`, true}, {"reasoning_effort", `"high"`, true}, {"reasoning_effort", `"max"`, true}, {"reasoning_effort", `"medium"`, false}, {"reasoning_effort", `null`, false}, {"thinking", `{"type":"disabled"}`, false},
				{"prompt_cache_options", `{"type":"implicit","ttl":"1h"}`, true}, {"prompt_cache_options", `{"type":"implicit","ttl":"30m"}`, false}, {"prompt_cache_options", `{"type":"implicit","ttl":"5m","extra":true}`, false}, {"prompt_cache_retention", `"24h"`, false},
				{"tool_choice", `"none"`, true}, {"tool_choice", `"auto"`, true}, {"tool_choice", `null`, false}, {"tool_choice", `"required"`, false}, {"tool_choice", `{"type":"function","function":{"name":"lookup"}}`, false},
				{"stream_options", `{"include_usage":false}`, false}, {"response_format", `{"type":"json_object"}`, false}, {"untrusted_custom_name", `true`, false},
			} {
				t.Run(string(market)+"/"+source+"/"+sc.field+"-"+sc.raw, func(t *testing.T) {
					ev := run.Case(t, kimiChatID(market, "K04-final-")+t.Name())
					params := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
					w := newKimiChatWorld(t, market, func(c *ai.Catalog) {
						for i := range c.Models {
							if c.Models[i].Provider != market {
								continue
							}
							if source == "model" {
								c.Models[i].SamplingParams = params
							}
							if source == "profile" {
								c.Models[i].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingMax: params}
							}
						}
					})
					enqueue(ev, w, kimiChatReply(t))
					opts := ai.ChatOptions{}
					hooks := ai.Hooks{}
					if source == "call" || source == "repair" {
						opts.SamplingParams = params
					}
					if source == "hook" || source == "repair" || source == "authorize-first" {
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
								p.Body["prompt_cache_options"] = map[string]any{"type": "implicit", "ttl": "5m"}
								p.Body["stream_options"] = map[string]any{"include_usage": true}
							}
							if source == "authorize-first" {
								p.Body["model"] = "unauthorized"
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if source != "authorize-first" && (sc.valid || source == "repair") {
						ev.Check("legal final body sends", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
						return
					}
					code := ai.CodeInvalidRequest
					if source == "authorize-first" {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, ai.PhaseRequest)
					ev.Check("no HTTP or Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%v", o.err)
					if sc.field == "untrusted_custom_name" && o.err != nil {
						ev.Check("untrusted field name absent from error", !strings.Contains(o.err.Error(), sc.field), "%v", o.err)
					}
				})
			}
		}
	}
}

func TestKimiChatNamedOptions(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, sc := range []struct {
			name string
			o    ai.ChatOptions
		}{
			{"medium", ai.ChatOptions{ReasoningEffort: ai.ThinkingMedium}}, {"minimal", ai.ChatOptions{ReasoningEffort: ai.ThinkingMinimal}}, {"xhigh", ai.ChatOptions{ReasoningEffort: ai.ThinkingXHigh}}, {"zero-output", ai.ChatOptions{MaxTokens: ai.Value(0)}}, {"large-output", ai.ChatOptions{MaxTokens: ai.Value(131073)}}, {"bad-temperature", ai.ChatOptions{Temperature: ai.Value(0.5)}}, {"none-cache", ai.ChatOptions{CacheRetention: ai.CacheRetentionNone}}, {"named-tool", ai.ChatOptions{ToolChoice: ai.Value(ai.ChatToolChoice{Function: "lookup"})}}, {"allowed-tool", ai.ChatOptions{ToolChoice: ai.Value(ai.ChatToolChoice{AllowedTools: &ai.ResponsesAllowedTools{Mode: "auto", Functions: []string{"lookup"}}})}},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K04-named-")+sc.name)
				w := newKimiChatWorld(t, market, nil)
				scope := textScope(t.Name())
				o := drain(w.client.Stream(ctxFor(t), scope, kimiChatTarget(), kimiChatRequest(), sc.o))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("no secret read or network", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0, "%+v", w.host.ReadsFor(scope.RequestID))
			})
		}
		for _, level := range []ai.ThinkingLevel{ai.ThinkingMinimal, ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMax} {
			t.Run(string(market)+"/simple-"+string(level), func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K04-clamp-")+string(level))
				w := newKimiChatWorld(t, market, nil)
				enqueue(ev, w, kimiChatReply(t))
				o := drain(w.client.StreamSimple(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), ai.SimpleOptions{Reasoning: level}))
				o.record(ev)
				if !ev.Check("simple succeeds", o.err == nil, "%v", o.err) {
					return
				}
				want := map[ai.ThinkingLevel]string{ai.ThinkingMinimal: `"low"`, ai.ThinkingLow: `"low"`, ai.ThinkingMedium: `"high"`, ai.ThinkingHigh: `"high"`, ai.ThinkingXHigh: `"max"`, ai.ThinkingMax: `"max"`}[level]
				ev.Check("canonical clamped wire effort", string(bodyFields(t, w.provider.Requests()[0].Body)["reasoning_effort"]) == want, "%s", w.provider.Requests()[0].Body)
			})
		}
	}
}
