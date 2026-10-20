package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestRegionalProviderPreflightAndFinalGuards(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		for _, api := range regionalAPIs {
			for _, mode := range []string{"full-option", "model", "capability", "second-auth", "payload-panic"} {
				t.Run(string(p)+"/"+string(api)+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "RP12-guard-"+t.Name())
					w, target := regionalWorld(t, p, api)
					code, phase, reads := ai.CodeInvalidRequest, ai.PhaseRequest, 1
					var opts ai.Options
					hooks := ai.Hooks{}
					switch mode {
					case "full-option":
						reads, phase = 0, ai.PhaseCapability
						switch api {
						case ai.APIOpenAICompletions:
							opts = ai.ChatOptions{ToolChoice: ai.Value(ai.ChatToolChoice{Mode: "required"})}
						case ai.APIOpenAIResponses:
							opts = ai.ResponsesOptions{ToolChoice: ai.Value(ai.ResponsesToolChoice{Mode: "required"})}
						default:
							opts = ai.AnthropicOptions{ThinkingDisplay: "summarized"}
						}
					case "second-auth":
						code = ai.CodeTenantDenied
						hooks.TransformHeaders = func(_ context.Context, _ ai.CallScope, h http.Header) error {
							if api == ai.APIAnthropicMessages {
								h.Set("Authorization", "Bearer regional-forbidden")
							} else {
								h.Set("X-Api-Key", "regional-forbidden")
							}
							return nil
						}
					default:
						if mode == "model" {
							code = ai.CodeTenantDenied
						}
						if mode == "payload-panic" {
							code = ai.CodeCallbackFailed
						}
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							if mode == "payload-panic" {
								panic("synthetic private callback")
							}
							if mode == "model" {
								p.Body["model"] = "unauthorized-model"
							} else {
								switch api {
								case ai.APIOpenAICompletions:
									p.Body["store"] = false
								case ai.APIOpenAIResponses:
									p.Body["include"] = []any{"reasoning.encrypted_content"}
								default:
									p.Body["output_config"] = map[string]any{"effort": "high"}
								}
							}
							return ai.KeepPayload(), nil
						}
					}
					scope := textScope("regional-guard-" + mode)
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, target, miniMaxToolRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
					checkRejected(ev, o, code, phase)
					ev.Check("rejected before Attempt at secret boundary", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0 && w.host.ReadsFor(scope.RequestID).Credentials == reads, "%+v", w.host.ReadsFor(scope.RequestID))
				})
			}
		}
	}
}
