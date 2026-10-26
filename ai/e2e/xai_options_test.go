package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIOptionsBeforeCredentials(t *testing.T) {
	for _, sc := range []struct {
		name string
		opts ai.ResponsesOptions
	}{
		{"zero-budget", ai.ResponsesOptions{MaxTokens: ai.Value(0)}},
		{"large-budget", ai.ResponsesOptions{MaxTokens: ai.Value(128001)}},
		{"negative-budget", ai.ResponsesOptions{MaxTokens: ai.Value(-1)}},
		{"minimal", ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMinimal}},
		{"max", ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMax}},
		{"summary", ai.ResponsesOptions{ReasoningSummary: "auto"}},
		{"tier-null", ai.ResponsesOptions{ServiceTier: ai.Null[string]()}},
		{"tier", ai.ResponsesOptions{ServiceTier: ai.Value("default")}},
		{"temperature-null", ai.ResponsesOptions{Temperature: ai.Null[float64]()}},
		{"temperature-too-high", ai.ResponsesOptions{Temperature: ai.Value(1.0001)}},
		{"temperature-negative", ai.ResponsesOptions{Temperature: ai.Value(-0.1)}},
		{"named-tool", ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Function: "probe"})}},
		{"required-without-tools", ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "required"})}},
		{"null-choice", ai.ResponsesOptions{ToolChoice: ai.Null[ai.ResponsesToolChoice]()}},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E04-preflight-"+sc.name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), sc.opts))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
			ev.Check("no secrets or HTTP", w.host.ReadsFor(t.Name()).Credentials == 0 && len(w.provider.Requests()) == 0, "reads=%+v", w.host.ReadsFor(t.Name()))
		})
	}
}

func TestXAISamplingSubset(t *testing.T) {
	for _, source := range []string{"call", "model", "profile", "hook"} {
		for _, sc := range []struct {
			field, raw string
			valid      bool
		}{
			{"temperature", "0", true}, {"temperature", "1", true}, {"temperature", "null", false}, {"temperature", "1.00000000000000001", false},
			{"top_p", "0.95", true}, {"top_p", "0", false}, {"top_p", "0.95000000000000001", false}, {"top_p", "null", false},
			{"max_output_tokens", "1", true}, {"max_output_tokens", "15", true}, {"max_output_tokens", "16", true}, {"max_output_tokens", "128000", true}, {"max_output_tokens", "1.00000000000000001", false}, {"max_output_tokens", "128001", false}, {"max_output_tokens", "1e99999", false},
			{"reasoning", `{"effort":"high","summary":null}`, false}, {"reasoning", `{"effort":"none"}`, false}, {"reasoning", `{"effort":"low"}`, true},
			{"service_tier", "null", false}, {"stop", `["done"]`, false}, {"frequency_penalty", "0", false}, {"text", `{"format":{"type":"json_object"}}`, false},
		} {
			name := source + "/" + sc.field + "/" + sc.raw
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "XAI-E04-sampling-"+name)
				sampling := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
				w := newXAIWorld(t, func(c *ai.Catalog) {
					if source == "model" {
						c.Models[0].SamplingParams = sampling
					}
					if source == "profile" {
						c.Models[0].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingHigh: sampling}
					}
				})
				enqueue(ev, w, xaiReply(t))
				opts := ai.ResponsesOptions{}
				hooks := ai.Hooks{}
				if source == "call" {
					opts.SamplingParams = sampling
				}
				if source == "hook" {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						var v any
						d := json.NewDecoder(strings.NewReader(sc.raw))
						d.UseNumber()
						if err := d.Decode(&v); err != nil {
							t.Fatal(err)
						}
						p.Body[sc.field] = v
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), opts))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if sc.valid {
					ev.Check("supported exact sampling succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
				} else {
					phase, code := ai.PhaseCapability, ai.CodeInvalidRequest
					if source == "hook" {
						phase = ai.PhaseRequest
						code = ai.CodeCallbackFailed
					}
					checkRejected(ev, o, code, phase)
					ev.Check("invalid sampling sends no HTTP", len(w.provider.Requests()) == 0, "requests=%+v", w.provider.Requests())
					if source != "hook" {
						ev.Check("invalid sampling reads no secret", w.host.ReadsFor(t.Name()).Credentials == 0, "reads=%+v", w.host.ReadsFor(t.Name()))
					}
				}
			})
		}
	}
}

func TestXAIReasoningAndBudget(t *testing.T) {
	for _, entry := range []string{"full", "simple"} {
		for _, level := range []ai.ThinkingLevel{"", ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMinimal, ai.ThinkingMax} {
			if entry == "full" && (level == ai.ThinkingMinimal || level == ai.ThinkingMax) {
				continue
			}
			for _, budget := range []int{0, 1, 15, 16, 128000} {
				name := fmt.Sprintf("%s/%s/%d", entry, level, budget)
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "XAI-E04-mapped-options-"+name)
					w := newXAIWorld(t, nil)
					enqueue(ev, w, xaiReply(t))
					var o outcome
					if entry == "full" {
						opts := ai.ResponsesOptions{ReasoningEffort: level}
						if budget > 0 {
							opts.MaxTokens = ai.Value(budget)
						} else {
							opts.MaxTokens = ai.Null[int]()
						}
						o.result, o.err = w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), opts)
					} else {
						opts := ai.SimpleOptions{Reasoning: level}
						if budget > 0 {
							opts.MaxTokens = ai.Value(budget)
						}
						o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), opts)
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("supported options succeed", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
						return
					}
					body := bodyFields(t, w.provider.Requests()[0].Body)
					want := budget
					if want == 0 {
						want = 128000
					}
					ev.Check("small budgets never raised", string(body["max_output_tokens"]) == fmt.Sprint(want), "%s", body["max_output_tokens"])
					if level == "" {
						ev.Check("default effort omitted", body["reasoning"] == nil, "%s", body["reasoning"])
					} else {
						want := string(level)
						if level == ai.ThinkingMinimal {
							want = "low"
						}
						if level == ai.ThinkingMax {
							want = "xhigh"
						}
						ev.Check("explicit supported effort kept", jsonEqual(body["reasoning"], []byte(`{"effort":"`+want+`"}`)), "%s", body["reasoning"])
					}
				})
			}
		}
	}
}
