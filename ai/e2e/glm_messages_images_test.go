package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMMessagesFinalImages(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		for _, container := range []string{"user", "system", "tool-result"} {
			t.Run(model+"/"+container, func(t *testing.T) {
				ev := run.Case(t, "GLM07-final-image-"+model+"-"+container)
				w := newGLMMessagesWorld(t, nil)
				hooks := glmMessagesPayload(func(p *ai.Payload) {
					image := map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": "synthetic-image-data"}}
					switch container {
					case "user":
						p.Body["messages"] = []any{map[string]any{"role": "user", "content": []any{image}}}
					case "system":
						p.Body["system"] = []any{image}
					case "tool-result":
						p.Body["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "call-synthetic", "content": []any{image}}}}}
					}
				})
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm07-final-image"), ai.Target{BindingID: "glm-messages", ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, nil))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("text-only model never receives callback-injected native image", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMMessagesImagePropertyData(t *testing.T) {
	ev := run.Case(t, "GLM07-image-property-data")
	w := newGLMMessagesWorld(t, nil)
	enqueue(ev, w, glmMessagesCaptureByName(t, "glm-5.3-tool-continuation").reply(t))
	req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}, Tools: []ai.Tool{{Name: "lookup_probe_code", Parameters: []byte(`{"type":"object","properties":{"type":{"const":"image"},"source":{"type":"string"}},"required":[]}`)}}}
	o := drain(w.client.Stream(ctxFor(t), textScope("glm07-image-property"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.3"}, req, nil))
	o.record(ev)
	ev.Check("schema data is not mistaken for native image blocks", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
}
