package e2e

import (
	"context"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestMiniMaxSharedFrozenRetries(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		t.Run(r.name, func(t *testing.T) {
			ev := run.Case(t, "MM05-frozen-"+r.name)
			var cat *ai.Catalog
			var policy *ai.ResourcePolicy
			w := newMiniMaxSharedWorld(t, func(c *ai.Config) { cat, policy = c.Catalog, c.Policy })
			hash := catalogHash(t, w.client.Catalog())
			for i := range cat.Models {
				cat.Models[i].ID = "caller-mutated"
			}
			policy.MaxRequestBytes = 1
			ev.Check("constructor snapshot owns catalog", catalogHash(t, w.client.Catalog()) == hash, "catalog mutation escaped")
			miniMaxSharedRetries(w, r, miniMaxSharedA, 1)
			req, first, _, raw := r.script(t, "MiniMax-M3")
			ev.Fixture("synthetic-source.json", raw)
			bad := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic retry"}}`)}}
			w.provider.EnqueueAt(r.prefix(miniMaxSharedA), bad, first)
			entered, resume := make(chan struct{}), make(chan struct{})
			var order callLog
			var retained map[string]any
			params := map[string]json.RawMessage{"temperature": json.RawMessage(`0.25`)}
			opts := r.options()
			switch o := opts.(type) {
			case ai.ChatOptions:
				o.SamplingParams = params
				opts = o
			case ai.ResponsesOptions:
				o.SamplingParams = params
				opts = o
			}
			hooks := ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, h http.Header) error {
				order.add("headers")
				ev.Check("credential redacted at hook", h.Get("Authorization") == "[REDACTED]" || h.Get("X-Api-Key") == "[REDACTED]", "redaction absent")
				close(entered)
				select {
				case <-resume:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				order.add("payload")
				retained = p.Body
				field := map[string]string{"chat": "max_completion_tokens", "responses": "max_output_tokens", "messages": "max_tokens"}[r.name]
				p.Body[field] = 4096
				return ai.KeepPayload(), nil
			}, OnResponse: func(_ context.Context, _ ai.CallScope, info ai.ResponseInfo) error {
				order.add("response")
				retained["model"] = "after-serialization"
				retained["stream"] = false
				info.Header.Set("X-Request-Id", "changed-response-copy")
				return nil
			}}
			s := w.client.WithHooks(hooks).Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-frozen-"+r.name), r.target("MiniMax-M3"), req, opts)
			defer s.Close()
			waitFor(ev, "header barrier entered", entered)
			req.Tools[0].Parameters[0] = '['
			req.Messages[0] = ai.UserText("caller-mutated-input")
			params["temperature"][0] = '9'
			miniMaxSharedRetries(w, r, miniMaxSharedA, 0)
			binding := w.host.Binding(miniMaxSharedA.tenant, r.target("MiniMax-M3").BindingID)
			w.host.PutCredential(ai.Credential{OwnerTenantID: miniMaxSharedA.tenant, CredentialID: binding.CredentialRef, AccountScopeID: binding.AccountScopeID, Version: "v2", Active: true, APIKey: ai.NewSecret(miniMaxSharedB.secret)})
			close(resume)
			o := drain(s)
			miniMaxSharedRecord(ev, w, o)
			ev.Record("requests", w.provider.Requests())
			ev.Check("pinned retry and legal callbacks succeed", o.err == nil && o.result.Metadata.CredentialVersion == "v1" && len(o.result.Metadata.Attempts) == 2 && len(w.provider.Requests()) == 2, "%v %+v", o.err, o.result.Metadata)
			ev.Check("hooks once outside retries", reflect.DeepEqual(order.entries(), []string{"headers", "payload", "response"}), "%v", order.entries())
			if len(w.provider.Requests()) == 2 {
				a, b := w.provider.Requests()[0], w.provider.Requests()[1]
				ev.Check("identical authorized bytes and credential across attempts", string(a.Body) == string(b.Body) && a.KeyAlias == miniMaxSharedA.alias && b.KeyAlias == a.KeyAlias && !strings.Contains(string(a.Body), "caller-mutated") && !strings.Contains(string(a.Body), "after-serialization"), "frozen request changed")
			}
			o.result.Metadata.Attempts[0].ProviderRequestID = "caller-mutated"
			o.result.Message.Content[0] = ai.Text{Text: "caller-mutated"}
			again, _ := s.Result()
			ev.Check("returned Result owns slices", !strings.Contains(string(mustMarshal(t, again)), "caller-mutated"), "result mutation escaped")
			w.released(ev)
		})
	}
}
