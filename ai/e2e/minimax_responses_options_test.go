package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func TestMiniMaxResponsesEntries(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			t.Run(model+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "MM03-entries-"+t.Name())
				ev.Fixture("source.json", []byte(`{"source":"MiniMax official Responses schema 2026-10-18","mode":"synthetic"}`))
				w := newMiniMaxResponsesWorld(t, nil)
				enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
				var o outcome
				target, req := minimaxResponsesTarget(model), minimaxResponsesRequest()
				switch entry {
				case "complete":
					o.result, o.err = w.client.Complete(ctxFor(t), textScope(t.Name()), target, req, nil)
				case "complete-simple":
					o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope(t.Name()), target, req, ai.SimpleOptions{})
				case "stream-simple":
					o = drain(w.client.StreamSimple(ctxFor(t), textScope(t.Name()), target, req, ai.SimpleOptions{}))
				default:
					o = drain(w.client.Stream(ctxFor(t), textScope(t.Name()), target, req, nil))
				}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("one successful request", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
					return
				}
				r := w.provider.Requests()[0]
				fields := bodyFields(t, r.Body)
				ev.Check("explicit China Responses route", r.Path == "/v1/responses" && r.KeyAlias == tenantA.alias, "%+v", r)
				for _, field := range []string{"store", "reasoning", "include", "parallel_tool_calls", "service_tier", "prompt_cache_retention", "prompt_cache_options"} {
					_, present := fields[field]
					ev.Check("default omits "+field, !present, "%s", r.Body)
				}
			})
		}
	}
}

func TestMiniMaxResponsesFinalOptions(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, sc := range []struct {
			name, key, value string
			valid            bool
		}{
			{"budget-one", "max_output_tokens", "1", true}, {"budget-max", "max_output_tokens", "131072", true}, {"budget-fraction", "max_output_tokens", "1.00000000000000001", false}, {"budget-over", "max_output_tokens", "131073", false}, {"budget-null", "max_output_tokens", "null", false}, {"budget-zero", "max_output_tokens", "0", false},
			{"temperature-zero", "temperature", "0", false}, {"temperature-null", "temperature", "null", false}, {"temperature-one", "temperature", "1", true}, {"top-p-zero", "top_p", "0", false}, {"standard", "service_tier", `"standard"`, true}, {"priority", "service_tier", `"priority"`, false},
			{"parallel", "parallel_tool_calls", "false", false}, {"truncation", "truncation", `"disabled"`, false}, {"summary", "reasoning", `{"effort":"high","summary":"auto"}`, false}, {"format", "text", `{"format":{"type":"json_object"}}`, false}, {"text", "text", `{"format":{"type":"text"}}`, true}, {"strict", "tools", `[{"type":"function","name":"probe","parameters":{"type":"object"},"strict":false}]`, false},
		} {
			for _, source := range []string{"sampling", "hook"} {
				if source == "sampling" && sc.key == "tools" {
					continue
				}
				t.Run(model+"/"+sc.name+"/"+source, func(t *testing.T) {
					ev := run.Case(t, "MM03-final-options-"+t.Name())
					w := newMiniMaxResponsesWorld(t, nil)
					enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
					opts := ai.ResponsesOptions{}
					hooks := ai.Hooks{}
					if source == "sampling" {
						opts.SamplingParams = map[string]json.RawMessage{sc.key: json.RawMessage(sc.value)}
					} else {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							var v any
							d := json.NewDecoder(strings.NewReader(sc.value))
							d.UseNumber()
							if err := d.Decode(&v); err != nil {
								t.Fatal(err)
							}
							p.Body[sc.key] = v
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget(model), minimaxResponsesRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if sc.valid {
						ev.Check("supported final request succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
						if sc.key == "max_output_tokens" && len(w.provider.Requests()) == 1 {
							ev.Check("exact budget preserved", string(bodyFields(t, w.provider.Requests()[0].Body)[sc.key]) == sc.value, "%s", w.provider.Requests()[0].Body)
						}
					} else {
						checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
						ev.Check("rejected before attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
					}
				})
			}
		}
	}
}
