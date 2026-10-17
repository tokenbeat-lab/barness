package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxUnpricedAutomaticWriteCannotSucceed(t *testing.T) {
	for _, route := range miniMaxSharedRoutes {
		for _, sourced := range []bool{false, true} {
			name := route.name + "-synthetic"
			if sourced {
				name = route.name + "-sourced"
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "MM10-unpriced-write-"+name)
				w := newMiniMaxSharedWorld(t, func(c *ai.Config) {
					if sourced {
						cat := minimax.PricedM3Catalog()
						c.Catalog = &cat
					}
				})
				var reply provider.Reply
				switch route.name {
				case "chat":
					reply = chatEvents(t, []json.RawMessage{json.RawMessage(`{"model":"MiniMax-M3","choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":30,"completion_tokens":2,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":10}}}`)}, "")
				case "responses":
					reply = minimaxResponsesReply(t, `{"type":"response.completed","response":{"model":"MiniMax-M3","status":"completed","output":[],"usage":{"input_tokens":30,"output_tokens":2,"total_tokens":32,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":10}}}}`)
				default:
					reply = provider.SSE([][]byte{[]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"priced\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"MiniMax-M3\",\"content\":[],\"usage\":{\"input_tokens\":20,\"output_tokens\":0,\"cache_read_input_tokens\":0,\"cache_creation_input_tokens\":10}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")})
				}
				w.provider.Enqueue(reply)
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), route.target("MiniMax-M3"), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic pricing observation.")}}, route.options()))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				miniMaxSharedOneTerminal(ev, o)
				if sourced {
					ev.Check("unverified write has scoped failure", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				} else {
					ev.Check("explicit positive synthetic rate preserves normalization", o.err == nil && o.result.Message.Usage.Cost.CacheWrite > 0, "%v %+v", o.err, o.result)
				}
				ev.Check("known write usage and actual Attempt retained", len(o.result.Metadata.Attempts) == 1 && o.result.Message.Usage.CacheWrite == 10 && o.result.Metadata.Attempts[0].Usage.CacheWrite == 10, "%+v", o.result)
			})
		}
	}
}

func TestMiniMaxWritePriceUsesSelectedTier(t *testing.T) {
	for _, shape := range []struct {
		name       string
		base, tier float64
		input      int64
		valid      bool
	}{
		{"unpriced-base", 0, 1, 30, false}, {"priced-tier", 0, 1, 600000, true}, {"priced-base", 1, 0, 30, true}, {"unpriced-tier", 1, 0, 600000, false},
	} {
		t.Run(shape.name, func(t *testing.T) {
			ev := run.Case(t, "MM10-write-tier-"+shape.name)
			w := newMiniMaxSharedWorld(t, func(c *ai.Config) {
				cat := minimax.CandidateCatalog()
				for i := range cat.Models {
					cat.Models[i].Cost.CacheWrite = shape.base
					cat.Models[i].Cost.Tiers = []ai.CostTier{{InputTokensAbove: 512000, CostRates: ai.CostRates{Input: 2, Output: 3, CacheRead: 0.2, CacheWrite: shape.tier}}}
				}
				c.Catalog = &cat
			})
			reply := chatEvents(t, []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"model":"MiniMax-M3","choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":%d,"completion_tokens":2,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":10}}}`, shape.input))}, "")
			w.provider.Enqueue(reply)
			route := miniMaxSharedRoutes[0]
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), route.target("MiniMax-M3"), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic write tier.")}}, route.options()))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("actually selected write rate determines success", (o.err == nil) == shape.valid, "%v %+v", o.err, o.result.Message.Usage)
			ev.Check("known usage retained", len(o.result.Metadata.Attempts) == 1 && o.result.Message.Usage.CacheWrite == 10, "%+v", o.result)
			miniMaxSharedOneTerminal(ev, o)
		})
	}
}

func TestMiniMaxOneHourWriteHasNoNativeAnthropicPrice(t *testing.T) {
	ev := run.Case(t, "MM10-write-one-hour-unpriced")
	w := newMiniMaxSharedWorld(t, nil)
	w.provider.Enqueue(provider.SSE([][]byte{[]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"one-hour\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"MiniMax-M3\",\"content\":[],\"usage\":{\"input_tokens\":20,\"output_tokens\":0,\"cache_creation_input_tokens\":10,\"cache_creation\":{\"ephemeral_1h_input_tokens\":10}}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")}))
	route := miniMaxSharedRoutes[2]
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), route.target("MiniMax-M3"), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic cache observation.")}}, route.options()))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	long, known := o.result.Message.Usage.CacheWrite1h.Get()
	ev.Check("unsupported one-hour write never succeeds", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
	ev.Check("known lifetime usage retained", known && long == 10 && o.result.Message.Usage.CacheWrite == 10 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result)
	miniMaxSharedOneTerminal(ev, o)
}
