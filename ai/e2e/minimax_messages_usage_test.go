package e2e

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxMessagesUsage(t *testing.T) {
	for _, c := range []struct {
		name, start, delta   string
		report               ai.UsageReporting
		in, out, read, write int64
		interrupted          bool
	}{
		{"complete", `{"input_tokens":5,"output_tokens":1,"cache_read_input_tokens":2,"cache_creation_input_tokens":3}`, `{"output_tokens":7}`, ai.UsageComplete, 5, 7, 2, 3, false},
		{"authoritative", `{"input_tokens":5,"output_tokens":1,"cache_read_input_tokens":2,"cache_creation_input_tokens":3}`, `{"input_tokens":6,"output_tokens":7,"cache_read_input_tokens":4,"cache_creation_input_tokens":1,"cache_creation":{"ephemeral_1h_input_tokens":1}}`, ai.UsageComplete, 6, 7, 4, 1, false},
		{"zero", `{"input_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}`, `{"output_tokens":0}`, ai.UsageComplete, 0, 0, 0, 0, false},
		{"partial", `{"input_tokens":5}`, `{"output_tokens":7}`, ai.UsagePartial, 5, 7, 0, 0, false},
		{"unreported", `null`, `null`, ai.UsageUnreported, 0, 0, 0, 0, false},
		{"interrupted", `{"input_tokens":5,"cache_read_input_tokens":2}`, `null`, ai.UsagePartial, 5, 0, 2, 0, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-usage-"+c.name)
			rec := newRecorder()
			w := newMiniMaxMessagesConfigured(t, nil, func(config *ai.Config) { config.Observer = rec; config.Policy.MaxQueuedObservations = 64 })
			frames := []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"synthetic-usage","model":"MiniMax-M3","usage":` + c.start + `}}`)}
			if !c.interrupted {
				frames = append(frames, json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"max_tokens"},"usage":`+c.delta+`}`), json.RawMessage(`{"type":"message_stop"}`))
			}
			enqueue(ev, w, sseEvents(t, frames, provider.FramingLF))
			scope := textScope("mm04-usage-" + c.name)
			o := drain(w.client.Stream(ctxFor(t), scope, miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("terminal or interruption respected", (o.err != nil) == c.interrupted, "%v", o.err)
			u := o.result.Message.Usage
			ev.Check("usage uses authoritative parts without double counting", u.Input == c.in && u.Output == c.out && u.CacheRead == c.read && u.CacheWrite == c.write && u.TotalTokens == c.in+c.out+c.read+c.write, "%+v", u)
			ev.Check("no Anthropic one-hour cache assumption", u.CacheWrite1h.IsZero(), "%+v", u)
			if len(o.result.Metadata.Attempts) == 1 {
				a := o.result.Metadata.Attempts[0]
				ev.Check("Attempt preserves own reported usage", a.UsageReporting == c.report && reflect.DeepEqual(a.Usage, u), "%+v", a)
			}
			obs := rec.awaitCall(ev, scope.RequestID)
			ev.Check("single call and attempt lifecycle observed", reflect.DeepEqual(kinds(obs), lifecycle(scope.RequestID, 1)), "%+v", kinds(obs))
		})
	}
}
