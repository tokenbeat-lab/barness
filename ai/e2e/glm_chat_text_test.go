package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
)

// The two Host candidates exercise all public entries, nil/full options and
// WithHooks. Only the local Provider is mocked; the public Client is real.
func TestGLMChatText(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, hooked := range []bool{false, true} {
				name := model + "/" + entry
				if hooked {
					name += "/hooks"
				} else {
					name += "/plain"
				}
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "GLM02-text-"+name)
					cat := glm.ChatCatalog()
					hash, err := cat.Hash()
					ev.Check("candidate hash computes", err == nil, "%v", err)
					ev.Record("candidate-catalog", cat)
					ev.Record("candidate-hash", hash)
					w := newGLMChatWorld(t, nil)
					reply, raw := glmChatReply(t, model)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, reply)
					var order []string
					hooks := ai.Hooks{}
					if hooked {
						hooks = ai.Hooks{
							TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
								order = append(order, "header")
								h.Set("X-Host-Trace", "glm")
								return nil
							},
							OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
								order = append(order, "payload")
								ev.Check("callback identity", p.ProviderID == ai.ProviderID("zhipu") && p.ModelID == model && p.API == ai.APIOpenAICompletions, "%+v", p)
								p.Body["temperature"] = json.Number("0.25")
								return ai.KeepPayload(), nil
							},
							OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
								order = append(order, "response")
								return nil
							},
						}
					}
					client := w.client.WithHooks(hooks)
					target, scope := ai.Target{BindingID: "glm-chat", ModelID: model}, textScope("glm-text-"+name)
					req := glmChatRequest()
					var o outcome
					var envs []ai.EventEnvelope
					switch entry {
					case "complete":
						var opts ai.Options
						if hooked {
							opts = ai.ChatOptions{}
						}
						o.result, o.err = client.Complete(ctxFor(t), scope, target, req, opts)
					case "complete-simple":
						o.result, o.err = client.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{})
					default:
						var s *ai.Stream
						if entry == "stream-simple" {
							s = client.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{})
						} else {
							s = client.Stream(ctxFor(t), scope, target, req, nil)
						}
						defer s.Close()
						o.streamed = true
						for s.Next() {
							o.events = append(o.events, s.Event())
							envs = append(envs, s.Envelope())
						}
						o.result, o.err = s.Result()
					}
					o.record(ev)
					ev.Record("envelopes", envelopesJSON(envs))
					ev.Record("requests", w.provider.Requests())
					ev.Check("generation succeeds", o.err == nil && o.result.Message.StopReason == ai.StopReasonStop, "err=%v", o.err)
					m, meta := o.result.Message, o.result.Metadata
					var thinking ai.Thinking
					var text ai.Text
					for _, block := range m.Content {
						switch block := block.(type) {
						case ai.Thinking:
							thinking = block
						case ai.Text:
							text = block
						}
					}
					ev.Check("text and visible thinking survive", len(m.Content) == 2 && thinking.Thinking == "Let me think." && thinking.Signature == "reasoning_content" && text.Text == "Hello GLM.", "%+v", m.Content)
					if o.streamed {
						ev.Check("public stream events preserve block and terminal order", slices.Equal(eventSummary(o.events), []string{"start", "thinking_start#0", "thinking_delta#0:Let me think.", "text_start#1", "text_delta#1:Hello GLM.", "thinking_end#0:Let me think.", "text_end#1:Hello GLM.", "done:stop"}), "%v", eventSummary(o.events))
					}
					ev.Check("trailing usage survives finish", m.Usage.Input == 10 && m.Usage.Output == 8 && m.Usage.CacheRead == 2 && m.Usage.TotalTokens == 20, "%+v", m.Usage)
					ev.Check("call and message attribution", meta.ProviderID == ai.ProviderID("zhipu") && meta.API == ai.APIOpenAICompletions && meta.ModelID == model && meta.AccountScopeID == "acct-tenant-a-glm" && meta.BindingID == "glm-chat" && meta.CatalogHash == hash && m.Provider == ai.ProviderID("zhipu") && m.Model == model, "%+v", meta)
					ev.Check("one attributed attempt", len(meta.Attempts) == 1 && meta.Attempts[0].Usage == m.Usage, "%+v", meta.Attempts)
					ev.Check("events attributed", !slices.ContainsFunc(envs, func(e ai.EventEnvelope) bool {
						return e.Call.ProviderID != ai.ProviderID("zhipu") || e.Call.API != ai.APIOpenAICompletions || e.Call.ModelID != model || e.Call.AccountScopeID != "acct-tenant-a-glm"
					}), "%+v", envs)
					if hooked {
						ev.Check("callback order", slices.Equal(order, []string{"header", "payload", "response"}), "%v", order)
					}
					reqs := w.provider.Requests()
					if !ev.Check("one provider request", len(reqs) == 1, "%d", len(reqs)) {
						return
					}
					r := reqs[0]
					ev.Check("official resource suffix and explicit credential", r.Path == "/api/paas/v4/chat/completions" && r.KeyAlias == glmA.alias, "%+v", r)
					fields := bodyFields(t, r.Body)
					for _, forbidden := range []string{"store", "max_completion_tokens", "prompt_cache_key", "prompt_cache_retention", "stream_options", "thinking", "reasoning_effort"} {
						_, present := fields[forbidden]
						ev.Check("default omits "+forbidden, !present, "%s", r.Body)
					}
					var messages []map[string]any
					mustUnmarshal(t, fields["messages"], &messages)
					ev.Check("system role and independent assistant replay", len(messages) == 3 && messages[0]["role"] == "system" && messages[1]["reasoning_content"] == nil, "%v", messages)
					if entry == "stream-simple" || entry == "complete-simple" {
						ev.Check("simple budget uses max_tokens", string(fields["max_tokens"]) == "131072", "%s", r.Body)
					}
				})
			}
		}
	}
}
