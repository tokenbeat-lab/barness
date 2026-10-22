package e2e

import (
	"context"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestKimiNarrowOutputLimit(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []string{"chat", "responses"} {
			for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
				for _, mode := range []string{"omitted", "hook-remove", "hook-overflow"} {
					t.Run(string(market)+"/"+api+"/"+entry+"/"+mode, func(t *testing.T) {
						ev := run.Case(t, "KIMI-"+string(market)+"-"+api+"-K04-narrow-budget-"+entry+"-"+mode)
						narrow := func(c *ai.Catalog) {
							for i := range c.Models {
								c.Models[i].MaxTokens = 4096
							}
						}
						w := newKimiChatWorld(t, market, narrow)
						target, field := kimiChatTarget(), "max_completion_tokens"
						if api == "responses" {
							w = newKimiResponsesWorld(t, market, narrow)
							target, field = kimiResponsesTarget(), "max_output_tokens"
							enqueue(ev, w, kimiResponsesReply(t))
						} else {
							enqueue(ev, w, kimiChatReply(t))
						}
						client := w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							if mode == "hook-remove" {
								delete(p.Body, field)
							}
							if mode == "hook-overflow" {
								p.Body[field] = 4097
							}
							return ai.KeepPayload(), nil
						}})
						var opts ai.Options = ai.ChatOptions{}
						if api == "responses" {
							opts = ai.ResponsesOptions{}
						}
						var o outcome
						switch entry {
						case "stream":
							o = drain(client.Stream(ctxFor(t), textScope(t.Name()), target, kimiChatRequest(), opts))
						case "complete":
							o.result, o.err = client.Complete(ctxFor(t), textScope(t.Name()), target, kimiChatRequest(), opts)
						case "stream-simple":
							o = drain(client.StreamSimple(ctxFor(t), textScope(t.Name()), target, kimiChatRequest(), ai.SimpleOptions{}))
						case "complete-simple":
							o.result, o.err = client.CompleteSimple(ctxFor(t), textScope(t.Name()), target, kimiChatRequest(), ai.SimpleOptions{})
						}
						o.record(ev)
						ev.Record("requests", w.provider.Requests())
						if mode != "omitted" {
							checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
							ev.Check("cannot remove or bypass hard limit", len(w.provider.Requests()) == 0, "%v", o.err)
							return
						}
						if ev.Check("narrowed default succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
							ev.Check("wire honors Host model cap", string(bodyFields(t, w.provider.Requests()[0].Body)[field]) == "4096", "%s", w.provider.Requests()[0].Body)
						}
					})
				}
			}
		}
	}
}
