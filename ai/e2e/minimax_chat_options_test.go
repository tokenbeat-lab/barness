package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxChatNamedOptions(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		for _, level := range []ai.ThinkingLevel{ai.ThinkingMinimal, ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMax} {
			for _, simple := range []bool{false, true} {
				t.Run(model+"/"+string(level)+map[bool]string{false: "/full", true: "/simple"}[simple], func(t *testing.T) {
					ev := run.Case(t, "MM02-named-"+t.Name())
					w := newMiniMaxChatWorld(t, nil)
					enqueue(ev, w, miniMaxChatReply(t, model))
					var o outcome
					if simple {
						o = drain(w.client.StreamSimple(ctxFor(t), textScope(t.Name()), miniMaxChatTarget(model), miniMaxChatRequest(), ai.SimpleOptions{Reasoning: level}))
					} else {
						o = drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget(model), miniMaxChatRequest(), ai.ChatOptions{ReasoningEffort: level}))
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					supported := simple || level == ai.ThinkingHigh || model != "MiniMax-M3" && level != ai.ThinkingMinimal
					if !supported {
						checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
						ev.Check("before credentials", w.host.ReadsFor(textScope(t.Name()).RequestID).Credentials == 0 && len(w.provider.Requests()) == 0, "%v", w.provider.Requests())
						return
					}
					if !ev.Check("supported generation", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
						return
					}
					f := bodyFields(t, w.provider.Requests()[0].Body)
					ev.Check("adaptive switch", string(f["thinking"]) == `{"type":"adaptive"}`, "%s", w.provider.Requests()[0].Body)
					want := `"` + string(level) + `"`
					if simple && level == ai.ThinkingMinimal {
						want = `"low"`
					}
					if model == "MiniMax-M3" {
						want = ""
					}
					ev.Check("depth only when declared", string(f["reasoning_effort"]) == want, "%s", w.provider.Requests()[0].Body)
				})
			}
		}
	}
	for _, choice := range []ai.ChatToolChoice{{Mode: "auto"}, {Mode: "none"}, {Mode: "required"}, {Function: "lookup"}, {AllowedTools: &ai.ResponsesAllowedTools{Mode: "auto", Functions: []string{"lookup"}}}} {
		t.Run("choice-"+string(mustMarshal(t, choice)), func(t *testing.T) {
			ev := run.Case(t, "MM02-choice-"+t.Name())
			w := newMiniMaxChatWorld(t, nil)
			enqueue(ev, w, miniMaxChatReply(t, "MiniMax-M3"))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), ai.ChatOptions{ToolChoice: ai.Value(choice)}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if choice.Mode != "auto" {
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("before request", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%v", o.err)
				return
			}
			if ev.Check("public auto supported", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				ev.Check("default auto omitted", len(bodyFields(t, w.provider.Requests()[0].Body)["tool_choice"]) == 0, "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}
