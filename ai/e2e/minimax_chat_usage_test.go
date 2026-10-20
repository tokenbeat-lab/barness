package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxChatUsage(t *testing.T) {
	for _, sc := range []struct {
		name, usage string
		disabled    bool
		report      ai.UsageReporting
	}{
		{"missing", "", false, ai.UsageUnreported},
		{"partial", `{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2}}`, false, ai.UsagePartial},
		{"off", `{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2}}`, true, ai.UsageComplete},
		{"zero", `{"prompt_tokens":0,"completion_tokens":0,"prompt_tokens_details":{"cached_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}`, false, ai.UsageComplete},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM02-usage-"+sc.name)
			w := newMiniMaxChatWorld(t, nil)
			frames := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"Hello."},"finish_reason":"stop"}]}`)}
			if sc.usage != "" {
				frames = append(frames, json.RawMessage(`{"choices":[],"usage":`+sc.usage+`}`))
			}
			ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
			enqueue(ev, w, chatEvents(t, frames, ""))
			opts := ai.ChatOptions{}
			if sc.disabled {
				opts.SamplingParams = map[string]json.RawMessage{"thinking": json.RawMessage(`{"type":"disabled"}`)}
			}
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), opts))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("reporting reflects final thinking selection", o.err == nil && len(o.result.Metadata.Attempts) == 1 && o.result.Metadata.Attempts[0].UsageReporting == sc.report, "%v %+v", o.err, o.result.Metadata.Attempts)
		})
	}
}
