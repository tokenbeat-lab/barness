package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// No GLM source returned redacted thinking. These synthetic Host records test
// the conservative boundary: no native replay claim until own evidence exists.
func TestGLMMessagesRedactedBoundary(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		for _, trusted := range []bool{false, true} {
			name := model + "/untrusted"
			if trusted {
				name = model + "/trusted"
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "GLM07-redacted-"+name)
				w := newGLMMessagesWorld(t, nil)
				m := ai.AssistantMessage{API: ai.APIAnthropicMessages, Provider: ai.ProviderZhipu, Model: model, StopReason: ai.StopReasonStop,
					Content: []ai.AssistantContent{ai.Thinking{Thinking: "[Reasoning redacted]", Signature: "synthetic-unproven-redacted", Redacted: true}, ai.Text{Text: "Visible answer"}}}
				if trusted {
					var err error
					m.NativeState, err = ai.TrustNativeState(textScope("glm07-redacted"), ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: "acct-tenant-a-glm-messages", ProviderID: ai.ProviderZhipu, API: ai.APIAnthropicMessages, ModelID: model})
					ev.Check("Host record can attest provenance", err == nil, "%v", err)
				} else {
					enqueue(ev, w, glmMessagesCaptureByName(t, model+"-tool-continuation").reply(t))
				}
				req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic question"), m, ai.UserText("Continue")}}
				o := drain(w.client.Stream(ctxFor(t), textScope("glm07-redacted"), ai.Target{BindingID: "glm-messages", ModelID: model}, req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if trusted {
					checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
					ev.Check("unproven native redacted form is never sent", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
				} else {
					ev.Check("untrusted redacted state downgrades and continuation succeeds", o.err == nil && o.result.Metadata.NativeStateDowngrades.NoEnvelope == 1, "%v %+v", o.err, o.result.Metadata)
					if len(w.provider.Requests()) == 1 {
						ev.Check("no opaque redacted data reaches provider", !jsonHasKey(t, w.provider.Requests()[0].Body, "data"), "%s", w.provider.Requests()[0].Body)
					}
				}
			})
		}
	}
}

func TestGLMMessagesCapabilityAuthorization(t *testing.T) {
	ev := run.Case(t, "GLM07-capability-authorization-order")
	w := newGLMMessagesWorld(t, nil)
	hooks := glmMessagesPayload(func(p *ai.Payload) {
		p.Body["model"] = "unauthorized-model"
		p.Body["thinking"] = map[string]any{"type": "adaptive"}
	})
	o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm07-authorization"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.3"}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, nil))
	o.record(ev)
	checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
	ev.Check("authorization rejection wins before capability and send", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
}

// Protocol fields are checked at their containers. A tool schema property or
// argument named display/cache_control is user data and is not such a field.
func TestGLMMessagesToolPropertyNames(t *testing.T) {
	ev := run.Case(t, "GLM07-tool-property-names")
	w := newGLMMessagesWorld(t, nil)
	c := glmMessagesCaptureByName(t, "glm-5.3-tool-continuation")
	enqueue(ev, w, c.reply(t))
	req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}, Tools: []ai.Tool{{Name: "lookup_probe_code", Parameters: []byte(`{"type":"object","properties":{"display":{"type":"string"},"cache_control":{"type":"string"}},"required":[]}`)}}}
	o := drain(w.client.Stream(ctxFor(t), textScope("glm07-tool-properties"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.3"}, req, nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	ev.Check("ordinary schema property names accepted by local contract guard", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
}

func TestGLMMessagesSimpleToolNone(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM07-simple-tool-none-"+model)
			w := newGLMMessagesWorld(t, nil)
			o := drain(w.client.StreamSimple(ctxFor(t), textScope("glm07-tool-none"), ai.Target{BindingID: "glm-messages", ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, ai.SimpleOptions{ToolChoice: ai.Value(ai.ToolChoiceNone)}))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("unproven simple none choice never attempted", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}
