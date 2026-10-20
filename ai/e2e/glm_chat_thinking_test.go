package e2e

import (
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Literals come from the China standard Chat schema, independently for the
// two models. Simple canonical clamping does not assert runtime intensity.
func TestGLMChatThinking(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, simple := range []bool{false, true} {
			for _, c := range []struct {
				level          ai.ThinkingLevel
				wire53, wire52 string
				full53         bool
			}{
				{ai.ThinkingMinimal, "low", "minimal", false},
				{ai.ThinkingLow, "low", "high", true},
				{ai.ThinkingMedium, "high", "high", false},
				{ai.ThinkingHigh, "high", "high", true},
				{ai.ThinkingXHigh, "max", "max", false},
				{ai.ThinkingMax, "max", "max", true},
			} {
				entry := "full"
				if simple {
					entry = "simple"
				}
				name := model + "/" + entry + "/" + string(c.level)
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "GLM02-thinking-"+name)
					w := newGLMChatWorld(t, nil)
					reply, raw := glmChatReply(t, model)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, reply)
					target, scope := ai.Target{BindingID: "glm-chat", ModelID: model}, textScope("glm-thinking-"+name)
					var o outcome
					if simple {
						o.result, o.err = w.client.CompleteSimple(ctxFor(t), scope, target, glmChatRequest(), ai.SimpleOptions{Reasoning: c.level})
					} else {
						o.result, o.err = w.client.Complete(ctxFor(t), scope, target, glmChatRequest(), ai.ChatOptions{ReasoningEffort: c.level})
					}
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
					if model == "glm-5.3" && !simple && !c.full53 {
						ev.Check("unsupported explicit level refused", errors.Is(o.err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseCapability}), "%v", o.err)
						ev.Check("refused before credential or attempt", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
						return
					}
					ev.Check("documented level succeeds", o.err == nil, "%v", o.err)
					reqs := w.provider.Requests()
					if !ev.Check("one send", len(reqs) == 1, "%d", len(reqs)) {
						return
					}
					fields := bodyFields(t, reqs[0].Body)
					wire := c.wire53
					if model == "glm-5.2" {
						wire = c.wire52
					}
					ev.Check("independent canonical to wire mapping", jsonEqual(fields["reasoning_effort"], mustMarshal(t, wire)) && jsonEqual(fields["thinking"], []byte(`{"type":"enabled"}`)), "%s", reqs[0].Body)
				})
			}
		}
	}
}
