package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Existing frozen exports run through their real APIs. Neither modelPatch nor
// field removal may replace an actual request/event/result comparison.
func TestRegionalProviderFrozenDifferential(t *testing.T) {
	for _, r := range regionalFrozenRoutes {
		for _, model := range r.models {
			for _, entry := range []string{"stream", "streamSimple"} {
				for _, scenario := range []string{"text", "tools"} {
					t.Run(r.name+"/"+model+"/"+scenario+"/"+entry, func(t *testing.T) {
						o := openOracle(t)
						w, target := regionalFrozenWorld(t, r)
						target.ModelID = model
						srv := provider.New(map[string]string{tenantA.secret: tenantA.alias})
						defer srv.Close()
						req := miniMaxMessagesRequest()
						if scenario == "tools" {
							if r.provider == ai.ProviderZAI {
								req = glmToolRequest(model)
								prior := req.Messages[0].(ai.AssistantMessage)
								prior.Provider = r.provider
								req.Messages[0] = prior
							} else {
								req = miniMaxToolRequest()
							}
						}
						stages := []string{scenario}
						if scenario == "tools" {
							stages = append(stages, "tool-replay")
						}
						for _, stage := range stages {
							t.Run(stage, func(t *testing.T) {
								ev := run.Case(t, "PIDIFF-RP06-"+r.name+"-"+model+"-"+stage+"-"+entry)
								ev.ReplayEnv(pioracle.EnableEnv + "=1")
								reply, raw := regionalFrozenScript(t, r, model, stage)
								ev.Fixture("regional-script-source", raw)
								ev.Record("response-script", reply)
								w.provider.Enqueue(reply)
								srv.Enqueue(reply)
								opts := map[string]any{"maxRetries": 0}
								var full ai.Options
								simple := ai.SimpleOptions{}
								if scenario == "tools" {
									simple.Reasoning = ai.ThinkingHigh
									if entry == "streamSimple" {
										opts["reasoning"] = "high"
									} else {
										if r.api == ai.APIOpenAICompletions {
											full = ai.ChatOptions{ReasoningEffort: ai.ThinkingHigh}
											opts["reasoningEffort"] = "high"
										} else {
											full = ai.AnthropicOptions{ThinkingEnabled: ai.Value(true)}
											opts["thinkingEnabled"] = true
										}
									}
								}
								var stream *ai.Stream
								if entry == "streamSimple" {
									stream = w.client.StreamSimple(ctxFor(t), textScope(ev.ID()), target, req, simple)
								} else {
									stream = w.client.Stream(ctxFor(t), textScope(ev.ID()), target, req, full)
								}
								b, _ := barnessObservation(t, ev, w, stream, 0)
								res, err := stream.Result()
								if !ev.Check("public regional turn succeeds", err == nil && res.Message.Provider == r.provider && res.Message.API == r.api && res.Message.Model == model, "%v %+v", err, res.Metadata) {
									return
								}
								base := "/v1"
								if r.api == ai.APIAnthropicMessages {
									base = "/anthropic"
								}
								piRun, e := o.Run(ctxFor(t), pioracle.Case{API: string(r.api), Provider: string(r.provider), Model: model, BaseURL: srv.URL() + base, APIKey: tenantA.secret, Entry: entry, Context: piContext(t, req), Options: mustMarshal(t, opts)})
								if !ev.Check("actual frozen regional turn executes", e == nil, "%v", e) {
									return
								}
								pi, _ := piObservation(ev, srv, piRun)
								regionalFrozenAssertions(t, ev, r, model, stage, entry, pi, b)
								regionalFrozenCompare(t, ev, o, w, r, piRun, pi, b, []byte(reply.Script))
								n := 1
								if stage == "tool-replay" {
									n = 2
								}
								ev.Check("both real clients send equal bounded turns", len(w.provider.Requests()) == n && len(srv.Requests()) == n, "b=%d pi=%d", len(w.provider.Requests()), len(srv.Requests()))
								if stage == "tools" {
									ev.Check("Host receives complete tool turn", res.Message.StopReason == ai.StopReasonToolUse, "%+v", res.Message)
									req.Messages = append(req.Messages, res.Message)
									for _, block := range res.Message.Content {
										if c, ok := block.(ai.ToolCall); ok {
											req.Messages = append(req.Messages, ai.ToolResultText(c, "Synthetic result.", false))
										}
									}
								} else {
									ev.Check("both final messages have visible content", textOf(res.Message) != "" && len(pi.Events) > 0, "%s", pi.Result)
								}
							})
						}
					})
				}
			}
		}
	}
}
