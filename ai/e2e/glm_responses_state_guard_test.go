package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMResponsesFinalState(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, sc := range []struct {
			id, field, raw string
			code           ai.Code
		}{
			{"store", "store", `true`, ai.CodeTenantDenied}, {"previous", "previous_response_id", `"private-response"`, ai.CodeTenantDenied}, {"conversation", "conversation", `"private-conversation"`, ai.CodeTenantDenied}, {"prompt", "prompt", `{"id":"private-prompt"}`, ai.CodeTenantDenied}, {"background", "background", `false`, ai.CodeTenantDenied},
			{"item-reference", "input", `[{"type":"item_reference","id":"private-item"}]`, ai.CodeTenantDenied}, {"file-reference", "input", `[{"role":"user","content":[{"type":"input_file","file_id":"private-file"}]}]`, ai.CodeTenantDenied},
			{"cache-key", "prompt_cache_key", `"self-reported-session"`, ai.CodeTenantDenied},
			{"stream", "stream", `false`, ai.CodeInvalidRequest}, {"remove-store", "store", "", ai.CodeInvalidRequest}, {"remove-stream", "stream", "", ai.CodeInvalidRequest},
			{"schema-format", "text", `{"format":{"type":"json_schema","name":"strict","schema":{"type":"object"}}}`, ai.CodeInvalidRequest},
			{"native-extra", "input", `[{"type":"reasoning","id":"rs_untrusted","encrypted_content":"untrusted","content":{"type":"reasoning_text","text":"x"}}]`, ai.CodeInvalidRequest},
			{"output-message-id", "input", `[{"type":"message","role":"assistant","id":"msg_private","content":[{"type":"output_text","text":"x"}]}]`, ai.CodeInvalidRequest},
			{"function-namespace", "input", `[{"type":"function_call","call_id":"call_a","name":"lookup_probe_code","arguments":"{}","namespace":"private"}]`, ai.CodeInvalidRequest},
			{"native-array", "input", `[{"type":"reasoning","content":[{"type":"reasoning_text","text":"x"}]}]`, ai.CodeInvalidRequest},
		} {
			t.Run(model+"/"+sc.id, func(t *testing.T) {
				ev := run.Case(t, "GLM05-state-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				scenarios, raw := glmResponsesFixtures(t)
				ev.Fixture("text.json", raw)
				enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					if sc.raw == "" {
						delete(p.Body, sc.field)
					} else {
						var value any
						_ = json.Unmarshal([]byte(sc.raw), &value)
						p.Body[sc.field] = value
					}
					return ai.KeepPayload(), nil
				}}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-responses-state"), glmResponsesTarget(model), glmResponsesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, sc.code, ai.PhaseRequest)
				ev.Check("zero requests and Attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}
