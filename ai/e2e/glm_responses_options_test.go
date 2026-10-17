package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"slices"
	"testing"
)

func TestGLMResponsesEntries(t *testing.T) {
	scenarios, raw := glmResponsesFixtures(t)
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, hooked := range []bool{false, true} {
				name := model + "/" + entry
				if hooked {
					name += "/hooks"
				}
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "GLM04-entries-"+name)
					ev.Fixture("text.json", raw)
					w := newGLMResponsesWorld(t, nil)
					enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
					hooks := ai.Hooks{}
					called := false
					if hooked {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							called = true
							p.Body["temperature"] = 0.25
							return ai.KeepPayload(), nil
						}
					}
					client := w.client.WithHooks(hooks)
					var o outcome
					switch entry {
					case "complete":
						var opts ai.Options
						if hooked {
							opts = ai.ResponsesOptions{}
						}
						o.result, o.err = client.Complete(ctxFor(t), textScope(name), glmResponsesTarget(model), glmResponsesRequest(), opts)
					case "complete-simple":
						o.result, o.err = client.CompleteSimple(ctxFor(t), textScope(name), glmResponsesTarget(model), glmResponsesRequest(), ai.SimpleOptions{})
					case "stream-simple":
						o = drain(client.StreamSimple(ctxFor(t), textScope(name), glmResponsesTarget(model), glmResponsesRequest(), ai.SimpleOptions{}))
					default:
						o = drain(client.Stream(ctxFor(t), textScope(name), glmResponsesTarget(model), glmResponsesRequest(), nil))
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("entry succeeds", o.err == nil, "%v", o.err)
					ev.Check("hook invoked when configured", called == hooked, "%v", called)
					reqs := w.provider.Requests()
					if !ev.Check("one request", len(reqs) == 1, "%v", reqs) {
						return
					}
					r := reqs[0]
					fields := bodyFields(t, r.Body)
					ev.Check("independent Responses route and Bearer", r.Path == "/api/v1/responses" && r.KeyAlias == glmA.alias, "%+v", r)
					ev.Check("explicit stateless contract", string(fields["store"]) == "false", "%s", r.Body)
					for _, field := range []string{"reasoning", "include", "service_tier", "prompt_cache_retention", "prompt_cache_options", "previous_response_id", "conversation"} {
						_, present := fields[field]
						ev.Check("default omits "+field, !present, "%s", r.Body)
					}
					for _, header := range []string{"Session_id", "X-Client-Request-Id"} {
						_, present := r.Header[header]
						ev.Check("no affinity "+header, !present, "%v", r.Header)
					}
				})
			}
		}
	}
}

func TestGLMResponsesEffort(t *testing.T) {
	scenarios, raw := glmResponsesFixtures(t)
	for _, source := range []string{"full", "simple", "sampling", "hook"} {
		for _, level := range []string{"", "minimal", "low", "medium", "high", "xhigh", "max"} {
			name := source + "/" + level
			if level == "" {
				name += "omitted"
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "GLM04-effort-"+name)
				ev.Fixture("text.json", raw)
				w := newGLMResponsesWorld(t, nil)
				enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
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
				var o outcome
				if source == "simple" {
					o = drain(client.StreamSimple(ctxFor(t), textScope(name), glmResponsesTarget("glm-5.3"), glmResponsesRequest(), ai.SimpleOptions{Reasoning: ai.ThinkingLevel(level)}))
				} else {
					o = drain(client.Stream(ctxFor(t), textScope(name), glmResponsesTarget("glm-5.3"), glmResponsesRequest(), opts))
				}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				valid := source == "simple" || slices.Contains([]string{"", "low", "high", "max"}, level)
				if !valid {
					phase := ai.PhaseRequest
					if source == "full" {
						phase = ai.PhaseCapability
					}
					checkRejected(ev, o, ai.CodeInvalidRequest, phase)
					ev.Check("no attempted request", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
					return
				}
				ev.Check("supported or clamped effort succeeds", o.err == nil, "%v", o.err)
				if len(w.provider.Requests()) == 1 {
					fields := bodyFields(t, w.provider.Requests()[0].Body)
					if level == "" {
						_, has := fields["reasoning"]
						ev.Check("omitted effort leaves service default", !has, "%s", w.provider.Requests()[0].Body)
					} else {
						var reasoning map[string]any
						mustUnmarshal(t, fields["reasoning"], &reasoning)
						want := level
						if source == "simple" {
							switch level {
							case "minimal":
								want = "low"
							case "medium":
								want = "high"
							case "xhigh":
								want = "max"
							}
						}
						ev.Check("effort follows Responses profile only", reasoning["effort"] == want && len(reasoning) == 1, "%v", reasoning)
					}
				}
			})
		}
	}
}
