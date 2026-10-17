package e2e

import (
	"context"
	"github.com/tokenbeat-lab/barness/ai"
	"net/http"
	"strings"
	"testing"
)

func TestMiniMaxSharedAuthorization(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, mode := range []string{"disabled", "operation", "model-denied", "credential-owner", "header-key", "header-target", "payload-model", "payload-native", "payload-hosted"} {
			t.Run(r.name+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "MM05-authorization-"+r.name+mode)
				w := newMiniMaxSharedWorld(t, nil)
				b := w.host.Binding(miniMaxSharedA.tenant, r.target("MiniMax-M3").BindingID)
				target := r.target("MiniMax-M3")
				phase := ai.PhaseRequest
				reads := 1
				switch mode {
				case "disabled":
					b.Enabled = false
					phase = ai.PhaseBinding
					reads = 0
				case "operation":
					b.Operation = ai.OperationImage
					phase = ai.PhaseCapability
					reads = 0
				case "model-denied":
					target.ModelID = "unauthorized-model"
					phase = ai.PhaseCapability
					reads = 0
				case "credential-owner":
					w.host.PutCredentialAt(miniMaxSharedA.tenant, b.CredentialRef, ai.Credential{OwnerTenantID: miniMaxSharedB.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(miniMaxSharedB.secret)})
					phase = ai.PhaseConsistency
				}
				w.host.PutBinding(b)
				hooks := ai.Hooks{}
				if strings.HasPrefix(mode, "header-") {
					hooks.TransformHeaders = func(_ context.Context, _ ai.CallScope, h http.Header) error {
						if mode == "header-key" {
							h.Set("Authorization", "Bearer outside-account")
						} else {
							h.Set("Host", "outside.example")
						}
						return nil
					}
				}
				if strings.HasPrefix(mode, "payload-") {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						switch mode {
						case "payload-model":
							p.Body["model"] = "outside-model"
						case "payload-hosted":
							p.Body["tools"] = []any{map[string]any{"type": "web_search"}}
						case "payload-native":
							if r.name == "responses" {
								p.Body["input"] = []any{map[string]any{"type": "item_reference", "id": "foreign-state"}}
							} else {
								p.Body["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "file", "file_id": "foreign-state"}}}}
							}
						}
						return ai.KeepPayload(), nil
					}
				}
				scope := scopeFor(miniMaxSharedA, "mm05-auth-"+r.name+mode)
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, target, miniMaxMessagesRequest(), r.options()))
				miniMaxSharedRecord(ev, w, o)
				ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, ai.CodeTenantDenied, phase)
				ev.Check("authorization blocks before send at correct secret boundary", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0 && w.host.ReadsFor(scope.RequestID).Credentials == reads, "%+v", w.host.ReadsFor(scope.RequestID))
				w.released(ev)
			})
		}
	}
}
