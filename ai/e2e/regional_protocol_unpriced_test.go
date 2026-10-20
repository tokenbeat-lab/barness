package e2e

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestRegionalProviderUnpricedWrites(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		for _, api := range regionalAPIs {
			t.Run(string(p)+"/"+string(api), func(t *testing.T) {
				ev := run.Case(t, "RP15-unknown-write-"+t.Name())
				w, target := regionalWorld(t, p, api, func(c *ai.Config) {
					for i := range c.Catalog.Models {
						c.Catalog.Models[i].Cost.CacheWrite = 0
					}
				})
				var reply provider.Reply
				switch api {
				case ai.APIOpenAICompletions:
					reply = chatEvents(t, []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"known"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":3}}}`)}, "")
				case ai.APIOpenAIResponses:
					reply = minimaxResponsesReply(t, `{"type":"response.completed","response":{"status":"completed","output":[],"usage":{"input_tokens":10,"output_tokens":2,"total_tokens":12,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":3}}}}`)
				default:
					reply = provider.SSE([][]byte{[]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"regional-priced\",\"usage\":{\"input_tokens\":10,\"cache_read_input_tokens\":0,\"cache_creation_input_tokens\":3}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")})
				}
				enqueue(ev, w, reply)
				o := drain(w.client.Stream(ctxFor(t), textScope("regional-unknown-write"), target, miniMaxMessagesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("unknown write price cannot claim successful free usage", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				ev.Check("reported usage remains auditable", len(o.result.Metadata.Attempts) == 1 && o.result.Message.Usage.CacheWrite == 3 && o.result.Metadata.Attempts[0].Usage.CacheWrite == 3, "%+v", o.result)
			})
		}
	}
}

func TestRegionalProviderUnpricedWriteLifetime(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		t.Run(string(p), func(t *testing.T) {
			ev := run.Case(t, "RP15-unknown-write-lifetime-"+string(p))
			w, target := regionalWorld(t, p, ai.APIAnthropicMessages)
			// The synthetic catalog prices ordinary writes, but cannot grant an
			// unverified native Anthropic TTL premium to a compatible service.
			reply := provider.SSE([][]byte{[]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"regional-ttl\",\"usage\":{\"input_tokens\":10,\"cache_creation_input_tokens\":3,\"cache_creation\":{\"ephemeral_1h_input_tokens\":3}}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")})
			enqueue(ev, w, reply)
			o := drain(w.client.Stream(ctxFor(t), textScope("regional-write-ttl"), target, miniMaxMessagesRequest(), nil))
			o.record(ev)
			long, known := o.result.Message.Usage.CacheWrite1h.Get()
			ev.Check("unknown lifetime price cannot succeed", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
			ev.Check("lifetime count retained without invented premium", known && long == 3 && o.result.Message.Usage.CacheWrite == 3 && o.result.Message.Usage.Cost.CacheWrite == 0, "%+v", o.result.Message.Usage)
		})
	}
}
