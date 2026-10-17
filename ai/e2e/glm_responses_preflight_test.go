package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMResponsesPreflight(t *testing.T) {
	for _, c := range []struct {
		name   string
		change func(*ai.Binding)
		opts   ai.Options
		model  string
		code   ai.Code
		phase  ai.Phase
	}{
		{"wrong-options", nil, ai.ChatOptions{}, "glm-5.3", ai.CodeInvalidRequest, ai.PhaseCapability},
		{"disabled", func(b *ai.Binding) { b.Enabled = false }, ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMedium}, "glm-5.3", ai.CodeTenantDenied, ai.PhaseBinding},
		{"operation", func(b *ai.Binding) { b.Operation = ai.OperationImage }, ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMedium}, "glm-5.3", ai.CodeTenantDenied, ai.PhaseCapability},
		{"model-denied", nil, ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMedium}, "private-model", ai.CodeTenantDenied, ai.PhaseCapability},
		{"catalog-identity", func(b *ai.Binding) { b.AllowedModels = append(b.AllowedModels, "gpt-4.1-mini") }, nil, "gpt-4.1-mini", ai.CodeInvalidRequest, ai.PhaseCapability},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM04-preflight-"+c.name)
			w := newGLMResponsesWorld(t, nil)
			if c.change != nil {
				w.updateBindingOf(tenantA, "glm-responses", c.change)
			}
			scope := textScope("glm-preflight-" + c.name)
			o := drain(w.client.Stream(ctxFor(t), scope, ai.Target{BindingID: "glm-responses", ModelID: c.model}, glmResponsesRequest(), c.opts))
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
		{"named-not-repaired", ai.ResponsesOptions{ReasoningEffort: ai.ThinkingMedium}, ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["reasoning"] = map[string]any{"effort": "high"}
			return ai.KeepPayload(), nil
		}}, ai.CodeInvalidRequest, ai.PhaseCapability, 0},
		{"header-auth", nil, ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("Authorization", "Bearer outside-account")
			return nil
		}}, ai.CodeTenantDenied, ai.PhaseRequest, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM04-preflight-"+c.name)
			w := newGLMResponsesWorld(t, nil)
			scope := textScope("glm-preflight-" + c.name)
			o := drain(w.client.WithHooks(c.hooks).Stream(ctxFor(t), scope, ai.Target{BindingID: "glm-responses", ModelID: "glm-5.3"}, glmResponsesRequest(), c.opts))
			o.record(ev)
			ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, c.code, c.phase)
			ev.Check("proper credential boundary and zero attempts", w.host.ReadsFor(scope.RequestID).Credentials == c.credentials && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}
