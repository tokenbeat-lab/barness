package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxSharedThinkingSemantics(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "MM05-thinking-semantics-"+model)
			w := newMiniMaxChatWorld(t, func(cat *ai.Catalog) {
				for i := range cat.Models {
					m := &cat.Models[i]
					if m.ID != model {
						continue
					}
					m.Compat.RequiresThinking = true
					m.ThinkingLevelMap.Off = ai.Null[string]()
					m.DefaultThinkingLevel = ai.ThinkingHigh
					if model != "MiniMax-M3" {
						m.ThinkingLevelMap.Low = ai.Null[string]()
						m.ThinkingLevelMap.Medium = ai.Null[string]()
						m.ThinkingLevelMap.XHigh = ai.Null[string]()
						m.ThinkingLevelMap.Max = ai.Null[string]()
					}
				}
			})
			enqueue(ev, w, miniMaxChatReply(t, model))
			o := drain(w.client.Stream(ctxFor(t), textScope("mm05-thinking-"+model), miniMaxChatTarget(model), miniMaxChatRequest(), ai.ChatOptions{ReasoningEffort: ai.ThinkingHigh}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Record("catalog", w.client.Catalog())
			if !ev.Check("generation succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			f := bodyFields(t, w.provider.Requests()[0].Body)
			want := `"high"`
			if model == "MiniMax-M3" {
				want = ""
			}
			ev.Check("explicit semantic capability fixes depth shape", string(f["reasoning_effort"]) == want && string(f["thinking"]) == `{"type":"adaptive"}`, "%s", w.provider.Requests()[0].Body)
		})
	}
}
