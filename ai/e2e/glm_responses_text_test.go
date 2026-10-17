package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"reflect"
	"testing"
)

func TestGLMResponsesContent(t *testing.T) {
	scenarios, raw := glmResponsesFixtures(t)
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, sc := range scenarios {
			if sc.Model != "" && sc.Model != model {
				continue
			}
			t.Run(model+"/"+sc.Name, func(t *testing.T) {
				ev := run.Case(t, "GLM04-content-"+model+"-"+sc.Name)
				ev.Fixture("text.json", raw)
				w := newGLMResponsesWorld(t, nil)
				enqueue(ev, w, glmResponsesReply(t, sc.Events))
				o := drain(w.client.Stream(ctxFor(t), textScope("glm04-"+sc.Name), glmResponsesTarget(model), glmResponsesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if sc.Failure {
					checkFailed(ev, o.result, o.err)
					return
				}
				ev.Check("legal completed without DONE succeeds", o.err == nil && (o.result.Message.StopReason == ai.StopReasonStop || (sc.StopReason == "length" && o.result.Message.StopReason == ai.StopReasonLength)), "%v", o.err)
				var thinking, text string
				var native string
				for _, b := range o.result.Message.Content {
					switch b := b.(type) {
					case ai.Thinking:
						thinking += b.Thinking
						native = b.Signature
					case ai.Text:
						text += b.Text
					}
				}
				ev.Check("authoritative visible content survives", thinking == sc.Thinking && text == sc.Text, "thinking=%q text=%q", thinking, text)
				if len(sc.Native) > 0 {
					ev.Check("native reasoning retains original wire shape", jsonEqual([]byte(native), sc.Native), "%s", native)
				}
				if len(o.events) > 0 {
					end, ok := o.events[len(o.events)-1].(ai.DoneEvent)
					ev.Check("done agrees with Result", ok && reflect.DeepEqual(end.Message, o.result.Message), "%v", o.events)
				}
			})
		}
	}
}
