package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxSourcedPriceTiers(t *testing.T) {
	for _, route := range miniMaxSharedRoutes {
		for _, input := range []int{511999, 512000, 512001} {
			t.Run(fmt.Sprintf("%s/%d", route.name, input), func(t *testing.T) {
				ev := run.Case(t, fmt.Sprintf("MM10-price-%s-%d", route.name, input))
				w := newMiniMaxSharedWorld(t, func(c *ai.Config) { catalog := minimax.PricedM3Catalog(); c.Catalog = &catalog })
				cached := 100
				uncached := input - cached
				output := 20
				var reply provider.Reply
				switch route.name {
				case "chat":
					reply = chatEvents(t, []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"model":"MiniMax-M3","choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":%d,"completion_tokens":20,"prompt_tokens_details":{"cached_tokens":100}}}`, input))}, "")
				case "responses":
					reply = minimaxResponsesReply(t, fmt.Sprintf(`{"type":"response.completed","response":{"model":"MiniMax-M3","status":"completed","output":[{"id":"m","type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":%d,"output_tokens":20,"total_tokens":%d,"input_tokens_details":{"cached_tokens":100}}}}`, input, input+20))
				default:
					reply = provider.SSE([][]byte{[]byte(fmt.Sprintf("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"priced\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"MiniMax-M3\",\"content\":[],\"usage\":{\"input_tokens\":%d,\"output_tokens\":0,\"cache_read_input_tokens\":100,\"cache_creation_input_tokens\":0}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":20}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n", uncached))})
				}
				w.provider.Enqueue(reply)
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), route.target("MiniMax-M3"), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic pricing only.")}}, route.options()))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				// Independent decimal source values: fixed CNY/6.7220 rounded10 places.
				inRate, outRate, cacheRate := .3124070217, 1.2496280869, .0624814043
				if input > 512000 {
					inRate, outRate, cacheRate = .6248140434, 2.4992561738, .1249628087
				}
				want := float64(uncached)*inRate/1e6 + float64(output)*outRate/1e6 + float64(cached)*cacheRate/1e6
				ev.Check("official sourced tier inclusive boundary", o.err == nil && math.Abs(o.result.Message.Usage.Cost.Total-want) < 1e-12, "err=%v usage=%+v want=%g", o.err, o.result.Message.Usage, want)
			})
		}
	}
}
