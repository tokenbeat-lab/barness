package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMSharedCallbackFailures(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, mode := range []string{"headers-error", "headers-panic", "payload-error", "payload-panic", "response-error", "response-panic", "key", "target", "model", "native", "hosted"} {
				t.Run(r.name+"/"+model+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "GLM09-callback-"+t.Name())
					w := newGLMSharedWorld(t, nil)
					req, first, _, raw := r.script(t, model)
					ev.Fixture(r.name+".json", raw)
					ev.Record("response-script", first)
					w.provider.EnqueueAt(r.prefix(glmA), first)
					private := errors.New("synthetic-private-callback-detail")
					hooks := ai.Hooks{}
					if strings.HasPrefix(mode, "headers") || mode == "key" || mode == "target" {
						hooks.TransformHeaders = func(_ context.Context, _ ai.CallScope, h http.Header) error {
							switch mode {
							case "headers-error":
								return private
							case "headers-panic":
								panic(private)
							case "key":
								h.Set("Authorization", "Bearer "+glmB.secret)
							case "target":
								h.Set("Host", "other.example")
							}
							return nil
						}
					} else if strings.HasPrefix(mode, "response") {
						hooks.OnResponse = func(context.Context, ai.CallScope, ai.ResponseInfo) error {
							if mode == "response-panic" {
								panic(private)
							}
							return private
						}
					} else {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							switch mode {
							case "payload-error":
								return ai.KeepPayload(), private
							case "payload-panic":
								panic(private)
							case "model":
								p.Body["model"] = "unauthorized-model"
							case "native":
								switch r.name {
								case "responses":
									p.Body["input"] = append(p.Body["input"].([]any), map[string]any{"type": "item_reference", "id": "rs_other_tenant"})
								case "chat":
									p.Body["messages"] = append(p.Body["messages"].([]any), map[string]any{"role": "user", "content": []any{map[string]any{"type": "file", "file": map[string]any{"file_id": "file-other-tenant"}}}})
								case "messages":
									p.Body["messages"] = append(p.Body["messages"].([]any), map[string]any{"role": "user", "content": []any{map[string]any{"type": "document", "source": map[string]any{"type": "file", "file_id": "file-other-tenant"}}}})
								}
							case "hosted":
								p.Body["tools"] = []any{map[string]any{"type": "web_search", "server_url": "https://other.example"}}
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scopeFor(glmA, "glm09-cb-"+r.name+model+mode), r.target(model), req, r.options()))
					glmSharedRecord(ev, w, o)
					ev.Record("requests", w.provider.Requests())
					code := ai.CodeCallbackFailed
					if mode == "key" || mode == "target" || mode == "model" || mode == "native" || mode == "hosted" {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, ai.PhaseRequest)
					want := 0
					if strings.HasPrefix(mode, "response") {
						want = 1
					}
					ev.Check("preflight zero attempts; post-send retains actual Attempt", len(w.provider.Requests()) == want && len(o.result.Metadata.Attempts) == want, "%+v", o.result.Metadata)
					if want == 1 {
						ev.Check("response callback precedes body parsing, usage remains unreported", o.result.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported, "%+v", o.result.Metadata.Attempts)
					}
					ev.Check("callback detail remains private", !strings.Contains(o.result.Message.ErrorMessage, private.Error()), "%s", o.result.Message.ErrorMessage)
					w.released(ev)
				})
			}
		}
	}
}

func TestGLMSharedCopiesAndHooks(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			t.Run(r.name+"/"+model, func(t *testing.T) {
				ev := run.Case(t, "GLM09-copies-"+t.Name())
				var catalog *ai.Catalog
				var policy *ai.ResourcePolicy
				w := newGLMSharedWorld(t, func(c *ai.Config) {
					catalog = c.Catalog
					policy = c.Policy
					for i := range catalog.Models {
						catalog.Models[i].SamplingParams = map[string]json.RawMessage{"temperature": json.RawMessage(`0.25`)}
						catalog.Models[i].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingLow: {"top_p": json.RawMessage(`0.80`)}}
					}
				})
				original, _ := w.client.Catalog().Hash()
				for i := range catalog.Models {
					catalog.Models[i].ID = "caller-mutated"
					catalog.Models[i].SamplingParams["temperature"][0] = '9'
					catalog.Models[i].SamplingParamsByThinkingLevel[ai.ThinkingLow]["top_p"][0] = '9'
				}
				policy.MaxRequestBytes = 1
				exposed := w.client.Catalog()
				exposed.Models[0].ID = "return-mutated"
				exposed.Models[0].SamplingParams["temperature"][0] = '9'
				exposed.Models[0].SamplingParamsByThinkingLevel[ai.ThinkingLow]["top_p"][0] = '9'
				got, _ := w.client.Catalog().Hash()
				ev.Check("constructor and returned catalog are owned copies", got == original, "%s %s", got, original)
				req, first, _, raw := r.script(t, model)
				ev.Fixture(r.name+".json", raw)
				ev.Record("response-script", first)
				w.provider.EnqueueAt(r.prefix(glmA), first)
				opts := r.options()
				params := map[string]json.RawMessage{"temperature": json.RawMessage(`0.25`)}
				switch v := opts.(type) {
				case ai.ChatOptions:
					v.SamplingParams = params
					opts = v
				case ai.ResponsesOptions:
					v.SamplingParams = params
					opts = v

				}
				entered, resume := make(chan struct{}), make(chan struct{})
				var order callLog
				var retained map[string]any
				hooks := ai.Hooks{
					TransformHeaders: func(ctx context.Context, s ai.CallScope, h http.Header) error {
						order.add("headers")
						ev.Check("scope and redacted credentials reach only own hook", s.TenantID == glmA.tenant && (h.Get("Authorization") == "[REDACTED]" || h.Get("X-Api-Key") == "[REDACTED]"), "%+v", s)
						h.Set("X-Host-Tenant", s.TenantID)
						close(entered)
						select {
						case <-resume:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					},
					OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						order.add("payload")
						retained = p.Body
						field := "max_tokens"
						if r.name == "responses" {
							field = "max_output_tokens"
						}
						p.Body[field] = 4096
						return ai.KeepPayload(), nil
					},
					OnResponse: func(_ context.Context, _ ai.CallScope, info ai.ResponseInfo) error {
						order.add("response")
						retained["model"] = "after-serialization"
						retained["stream"] = false
						info.Header.Set("X-Request-Id", "caller-changed")
						return nil
					},
				}
				s := w.client.WithHooks(hooks).Stream(ctxFor(t), scopeFor(glmA, "glm09-copies-"+r.name+model), r.target(model), req, opts)
				defer s.Close()
				waitFor(ev, "snapshot reaches header barrier", entered)
				req.Tools[0].Parameters[0] = '['
				req.Messages[0] = ai.UserText("caller-mutated-input")
				params["temperature"][0] = '9'
				close(resume)
				var events []ai.Event
				for s.Next() {
					e := s.Event()
					events = append(events, e)
					if d, ok := e.(ai.ToolCallDeltaEvent); ok {
						m := d.Partial.Snapshot()
						if len(m.Content) > 0 {
							m.Content[0] = ai.Text{Text: "caller-mutated-partial"}
						}
					}
				}
				result, err := s.Result()
				o := outcome{result: result, err: err, events: events}
				glmSharedRecord(ev, w, o)
				ev.Check("snapshots and legal hooks preserve successful call", err == nil && result.Message.StopReason == ai.StopReasonToolUse, "%v", err)
				ev.Check("header payload response order", reflect.DeepEqual(order.entries(), []string{"headers", "payload", "response"}), "%v", order.entries())
				if len(result.Message.Content) > 0 {
					result.Message.Content[0] = ai.Text{Text: "caller-mutated-result"}
				}
				if len(result.Metadata.Attempts) > 0 {
					result.Metadata.Attempts[0].ProviderRequestID = "caller-mutated-attempt"
				}
				again, _ := s.Result()
				ev.Check("Result reread owns all returned slices", !strings.Contains(string(mustMarshal(t, again)), "caller-mutated"), "%+v", again)
				ev.Record("requests", w.provider.Requests())
				if len(w.provider.Requests()) == 1 {
					q := w.provider.Requests()[0]
					ev.Check("wire is frozen before caller mutations", strings.Contains(string(q.Body), model) && !strings.Contains(string(q.Body), "caller-mutated") && !strings.Contains(string(q.Body), "after-serialization") && q.KeyAlias == glmA.alias, "%s", q.Body)
					if r.name != "messages" {
						ev.Check("sampling raw JSON frozen", string(bodyFields(t, q.Body)["temperature"]) == "0.25", "%s", q.Body)
					}
				}
				w.released(ev)
			})
		}
	}
}
