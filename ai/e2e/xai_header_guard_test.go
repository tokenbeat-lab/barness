package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAICacheHeaderAuthority(t *testing.T) {
	for _, sc := range []struct {
		name, key string
		values    []string
		denied    bool
	}{
		{"absent", "", nil, false},
		{"canonical", "X-Grok-Conv-Id", []string{"raw-session"}, true},
		{"lowercase", "x-grok-conv-id", []string{"raw-session"}, true},
		{"mixed-case", "x-GrOk-CoNv-Id", []string{"raw-session"}, true},
		{"empty-value", "X-Grok-Conv-Id", []string{""}, true},
		{"multiple-values", "X-Grok-Conv-Id", []string{"", "foreign-session"}, true},
		{"no-values", "X-Grok-Conv-Id", nil, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-review-cache-header-"+sc.name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			h := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, header http.Header) error {
				if sc.key != "" {
					header[sc.key] = sc.values
				}
				return nil
			}}
			o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), ai.ResponsesOptions{SessionID: "derived-session"}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if sc.denied {
				checkRejected(ev, o, ai.CodeTenantDenied, ai.PhaseRequest)
				ev.Check("alternate cache routing never reaches HTTP", len(w.provider.Requests()) == 0, "requests=%d", len(w.provider.Requests()))
			} else {
				ev.Check("absent routing header preserves call", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
			}
		})
	}
	t.Run("openai-custom-header", func(t *testing.T) {
		ev := run.Case(t, "XAI-review-cache-header-other-provider")
		w := newWorld(t, tenantA)
		enqueue(ev, w, xaiReply(t))
		h := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, header http.Header) error {
			header.Set("X-Grok-Conv-Id", "host-diagnostic")
			return nil
		}}
		o := drain(w.client.WithHooks(h).Stream(ctxFor(t), textScope(t.Name()), ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}, xaiRequest(), nil))
		o.record(ev)
		ev.Record("requests", w.provider.Requests())
		ev.Check("other Provider custom header semantics preserved", o.err == nil && len(w.provider.Requests()) == 1 && w.provider.Requests()[0].Header["X-Grok-Conv-Id"] == "host-diagnostic", "%v", o.err)
	})
}
