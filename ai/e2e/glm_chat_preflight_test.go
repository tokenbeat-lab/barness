package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMChatPreflight(t *testing.T) {
	for _, c := range []struct {
		name   string
		change func(*ai.Binding)
		opts   ai.Options
		model  string
		code   ai.Code
		phase  ai.Phase
	}{
		{"wrong-options", nil, ai.ResponsesOptions{}, "glm-5.3", ai.CodeInvalidRequest, ai.PhaseCapability},
		{"disabled", func(b *ai.Binding) { b.Enabled = false }, ai.ChatOptions{ReasoningEffort: ai.ThinkingMedium}, "glm-5.3", ai.CodeTenantDenied, ai.PhaseBinding},
		{"operation", func(b *ai.Binding) { b.Operation = ai.OperationImage }, ai.ChatOptions{ReasoningEffort: ai.ThinkingMedium}, "glm-5.3", ai.CodeTenantDenied, ai.PhaseCapability},
		{"model-denied", nil, ai.ChatOptions{ReasoningEffort: ai.ThinkingMedium}, "private-model", ai.CodeTenantDenied, ai.PhaseCapability},
		{"catalog-identity", func(b *ai.Binding) { b.AllowedModels = append(b.AllowedModels, "gpt-4.1-mini") }, nil, "gpt-4.1-mini", ai.CodeInvalidRequest, ai.PhaseCapability},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM02-preflight-"+c.name)
			w := newGLMChatWorld(t, nil)
			if c.change != nil {
				w.updateBindingOf(tenantA, "glm-chat", c.change)
			}
			scope := textScope("glm-preflight-" + c.name)
			o := drain(w.client.Stream(ctxFor(t), scope, ai.Target{BindingID: "glm-chat", ModelID: c.model}, glmChatRequest(), c.opts))
			o.record(ev)
			ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, c.code, c.phase)
			ev.Check("no credential, provider request or attempt", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
	// A callback cannot repair an explicitly unsupported named canonical
	// input, nor override authentication. Both refuse before any Attempt.
	for _, c := range []struct {
		name        string
		opts        ai.Options
		hooks       ai.Hooks
		code        ai.Code
		phase       ai.Phase
		credentials int
	}{
		{"named-not-repaired", ai.ChatOptions{ReasoningEffort: ai.ThinkingMedium}, ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["reasoning_effort"] = "high"
			return ai.KeepPayload(), nil
		}}, ai.CodeInvalidRequest, ai.PhaseCapability, 0},
		{"header-auth", nil, ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("Authorization", "Bearer outside-account")
			return nil
		}}, ai.CodeTenantDenied, ai.PhaseRequest, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM02-preflight-"+c.name)
			w := newGLMChatWorld(t, nil)
			scope := textScope("glm-preflight-" + c.name)
			o := drain(w.client.WithHooks(c.hooks).Stream(ctxFor(t), scope, ai.Target{BindingID: "glm-chat", ModelID: "glm-5.3"}, glmChatRequest(), c.opts))
			o.record(ev)
			ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, c.code, c.phase)
			ev.Check("proper credential boundary and zero attempts", w.host.ReadsFor(scope.RequestID).Credentials == c.credentials && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

func TestGLMChatDefaultThinkingUsage(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM02-default-thinking-usage-"+model)
			w := newGLMChatWorld(t, nil)
			frames, raw := glmChatFrames(t, model)
			ev.Fixture("text.json", raw)
			// A deliberately incomplete usage frame must remain partial even
			// when the Host omits the level and retains default enabled thinking.
			frames[3] = []byte(`{"id":"glm-synthetic-text","choices":[],"usage":{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2}}}`)
			enqueue(ev, w, chatEvents(t, frames, ""))
			res, err := w.client.Complete(ctxFor(t), textScope("glm-default-usage-"+model), ai.Target{BindingID: "glm-chat", ModelID: model}, glmChatRequest(), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("missing reasoning reporting stays partial", err == nil && len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsagePartial, "err=%v attempts=%+v", err, res.Metadata.Attempts)
		})
	}
}
