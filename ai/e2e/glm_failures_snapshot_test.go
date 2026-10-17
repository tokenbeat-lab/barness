package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGLMRetrySnapshot(t *testing.T) {
	for _, protocol := range []string{"responses", "chat", "messages"} {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			t.Run(protocol+"/"+model, func(t *testing.T) {
				ev := run.Case(t, "GLM08-snapshot-"+protocol+"-"+model)
				w, target, bodies := glmFailureWorld(t, protocol, 1)
				target.ModelID = model
				scope := textScope("glm08-snapshot-" + protocol + model)
				first := provider.Reply{Status: 429, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic throttle"}}`)}, Script: `{"error":{"message":"synthetic throttle"}}`}
				first.OnReceive = func() {
					b, err := w.host.ResolveBinding(context.Background(), textScope("glm08-configuration-change"), target.BindingID)
					if err != nil {
						return
					}
					b.Version = "rotated-v2"
					b.Retry.MaxRetries = 0
					w.host.PutBinding(b)
					w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "rotated-v2", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret("synthetic-rotated-glm-key")})
				}
				var events []json.RawMessage
				switch protocol {
				case "responses":
					events = []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}
				case "chat":
					events = []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}]}`)}
				case "messages":
					events = []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"synthetic"}}`), json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"}}`), json.RawMessage(`{"type":"message_stop"}`)}
				}
				enqueue(ev, w, first, glmOwnReply(t, protocol, events))
				calls := 0
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					calls++
					field := "max_tokens"
					if protocol == "responses" {
						field = "max_output_tokens"
					}
					p.Body[field] = 4096
					return ai.KeepPayload(), nil
				}}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, target, glmResponsesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("bounded retry succeeds under pinned policy", o.err == nil && len(o.result.Metadata.Attempts) == 2, "%v %+v", o.err, o.result.Metadata.Attempts)
				reqs := w.provider.Requests()
				if ev.Check("two external attempts", len(reqs) == 2, "%d", len(reqs)) {
					ev.Check("final payload bytes and credential frozen", calls == 1 && bytes.Equal(reqs[0].Body, reqs[1].Body) && reqs[0].KeyAlias == glmA.alias && reqs[1].KeyAlias == glmA.alias, "calls=%d", calls)
					ev.Check("no endpoint/model fallback", reqs[0].Path == reqs[1].Path && strings.Contains(string(reqs[1].Body), model), "%+v", reqs)
				}
				ev.Check("configuration resolved once", w.host.ReadsFor(scope.RequestID) == (host.Reads{Bindings: 1, Credentials: 1}), "%+v", w.host.ReadsFor(scope.RequestID))
				ev.Check("all bodies closed", bodies.Open() == 0, "%d", bodies.Open())
			})
		}
	}
}

func TestGLMFailureRequestIDAbsent(t *testing.T) {
	for _, protocol := range []string{"responses", "chat", "messages"} {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			t.Run(protocol+"/"+model, func(t *testing.T) {
				ev := run.Case(t, "GLM08-absent-id-"+protocol+"-"+model)
				w, target, _ := glmFailureWorld(t, protocol, 0)
				target.ModelID = model
				r := provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json", "X-Ds-Trace-Id": "not-glm", "request-id": "not-the-confirmed-header"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic"}}`)}, Script: `{"error":{"message":"synthetic"}}`}
				enqueue(ev, w, r)
				o := drain(w.client.Stream(ctxFor(t), textScope("local-id-never-vendor"+protocol+model), target, glmResponsesRequest(), nil))
				o.record(ev)
				ev.Check("missing provider ID is not invented", len(o.result.Metadata.Attempts) == 1 && o.result.Metadata.Attempts[0].ProviderRequestID == "", "%+v", o.result.Metadata.Attempts)
			})
		}
	}
}
