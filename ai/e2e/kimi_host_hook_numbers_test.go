package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestKimiHostHookExactNumbers(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, route := range []int{0, 1, 2} {
			for _, encoding := range []string{"number", "raw"} {
				for _, kind := range []string{"fractional-budget", "fixed-sampling", "authority-precedence"} {
					if route == 2 && kind == "fixed-sampling" {
						continue
					}
					t.Run(fmt.Sprintf("%s/%d/%s/%s", market, route, encoding, kind), func(t *testing.T) {
						ev := run.Case(t, fmt.Sprintf("KIMI-%s-%d-K14-exact-hook-%s-%s", market, route, encoding, kind))
						w := newKimiHostWorld(t)
						b := w.bindings[market][route]
						reply, raw := kimiHostScript(t, b.API)
						ev.Fixture("synthetic-script.json", raw)
						w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
						numeric := func(s string) any {
							if encoding == "raw" {
								return json.RawMessage(s)
							}
							return json.Number(s)
						}
						hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							field := "max_completion_tokens"
							if route == 1 {
								field = "max_output_tokens"
							}
							if route == 2 {
								field = "max_tokens"
							}
							p.Body[field] = numeric("131072.00000000000000001")
							if kind == "fixed-sampling" {
								p.Body[field] = numeric("131072")
								p.Body["top_p"] = numeric("0.950000000000000000001")
							}
							if kind == "authority-precedence" {
								p.Body["model"] = "unauthorized-model"
							}
							return ai.KeepPayload(), nil
						}}
						result, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("exact-hook"), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, nil)
						ev.Record("result", result)
						ev.Record("requests", w.provider.Requests())
						code := ai.CodeInvalidRequest
						if kind == "authority-precedence" {
							code = ai.CodeTenantDenied
						}
						e, ok := err.(*ai.Error)
						ev.Check("hook numeric syntax reaches exact guard after authority validation", ok && e.Code == code && e.Phase == ai.PhaseRequest && len(result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "err=%v attempts=%+v", err, result.Metadata.Attempts)
					})
				}
			}
		}
	}
}

func TestKimiHostHookExactNumbersAccepted(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, route := range []int{0, 1, 2} {
			t.Run(fmt.Sprintf("%s/%d", market, route), func(t *testing.T) {
				ev := run.Case(t, fmt.Sprintf("KIMI-%s-%d-K14-legal-hook-exact-number", market, route))
				w := newKimiHostWorld(t)
				b := w.bindings[market][route]
				reply, raw := kimiHostScript(t, b.API)
				ev.Fixture("synthetic-script.json", raw)
				w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					field := "max_completion_tokens"
					if route == 1 {
						field = "max_output_tokens"
					}
					if route == 2 {
						field = "max_tokens"
					}
					p.Body[field] = json.Number("131072")
					if route == 0 {
						p.Body["top_p"] = json.RawMessage("0.95")
					}
					return ai.KeepPayload(), nil
				}}
				result, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("legal-exact-hook"), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, nil)
				ev.Record("result", result)
				ev.Record("requests", w.provider.Requests())
				ev.Check("legal exact hook numbers remain usable", err == nil && len(w.provider.Requests()) == 1, "err=%v", err)
			})
		}
	}
}

func TestKimiHostHookLegacyNumberPreservation(t *testing.T) {
	ev := run.Case(t, "KIMI-K15-legacy-openai-hook-number-preservation")
	w := newWorld(t, tenantA)
	reply, raw := kimiHostScript(t, ai.APIOpenAICompletions)
	ev.Fixture("synthetic-script.json", raw)
	enqueue(ev, w, reply)
	result, err := w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		p.Body["top_p"] = json.Number("0.950000000000000000001")
		return ai.KeepPayload(), nil
	}}).Complete(ctxFor(t), textScope("legacy-exact-hook"), ai.Target{BindingID: "chat", ModelID: "gpt-4.1-mini"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, nil)
	ev.Record("result", result)
	ev.Record("requests", w.provider.Requests())
	if ev.Check("legacy OpenAI hook remains supported", err == nil && len(w.provider.Requests()) == 1, "err=%v", err) {
		fields := bodyFields(t, w.provider.Requests()[0].Body)
		ev.Check("legacy exact numeric JSON is retained", string(fields["top_p"]) == "0.950000000000000000001", "top_p=%s", fields["top_p"])
	}
}
