package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type miniMaxHostRoute struct {
	name   string
	api    ai.API
	suffix string
}

var miniMaxHostRoutes = []miniMaxHostRoute{{"chat", ai.APIOpenAICompletions, "/chat/completions"}, {"responses", ai.APIOpenAIResponses, "/responses"}, {"messages", ai.APIAnthropicMessages, "/v1/messages"}}
var miniMaxHostKeys = []tenantKey{{tenantA.tenant, "synthetic-mm08-standard-a", "key:mm08-standard-a"}, {tenantA.tenant, "synthetic-mm08-plan-a", "key:mm08-plan-a"}, {tenantB.tenant, "synthetic-mm08-standard-b", "key:mm08-standard-b"}, {tenantB.tenant, "synthetic-mm08-plan-b", "key:mm08-plan-b"}}

func init() { knownKeys = append(knownKeys, miniMaxHostKeys...) }
func miniMaxHostKey(k tenantKey, p minimaxhost.Purpose) tenantKey {
	offset := 0
	if k.tenant == tenantB.tenant {
		offset = 2
	}
	if p == minimaxhost.MPlan {
		offset++
	}
	return miniMaxHostKeys[offset]
}
func miniMaxHostModel(p minimaxhost.Purpose) string {
	if p == minimaxhost.MPlan {
		return "MiniMax-M3.1-Flash-Preview"
	}
	return "MiniMax-M3"
}
func miniMaxHostPrefix(k tenantKey, p minimaxhost.Purpose, r miniMaxHostRoute) string {
	return "/mm08/" + k.tenant + "/" + string(p) + "/" + r.name
}
func miniMaxHostTarget(p minimaxhost.Purpose, r miniMaxHostRoute) ai.Target {
	return ai.Target{BindingID: "minimax-" + string(p) + "-" + r.name, ModelID: miniMaxHostModel(p)}
}
func miniMaxHostScope(k tenantKey, id string) ai.CallScope {
	return ai.CallScope{TenantID: k.tenant, ActorID: "authenticated-example-actor", RequestID: id + "-" + requestid.New()}
}

type miniMaxHostWorld struct {
	limitWorld
	rec *recorder
}

func newMiniMaxHostWorld(t *testing.T) *miniMaxHostWorld {
	t.Helper()
	cat := ai.BuiltinCatalog()
	candidate := minimax.CandidateCatalog()
	cat.Version = "mm08-controlled-candidate-synthetic-prices"
	cat.Models = append(cat.Models, candidate.Models...)
	rec := newRecorder()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	gauges := probe.New()
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		c.Observer = rec
		c.Policy.MaxQueuedObservations = 128
		c.Transport = bodies
		c.Probe = gauges
	}, tenantA, tenantB)
	for _, k := range []tenantKey{tenantA, tenantB} {
		for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
			key := miniMaxHostKey(k, purpose)
			account := "acct-mm08-" + k.tenant + "-" + string(purpose)
			ref := "cred-mm08-" + k.tenant + "-" + string(purpose)
			bindings, err := minimaxhost.Bindings(minimaxhost.Authorization{TenantID: k.tenant, AccountScopeID: account, CredentialRef: ref, Version: "mm08-v1", Purpose: purpose})
			if err != nil {
				t.Fatal(err)
			}
			for _, b := range bindings {
				for _, r := range miniMaxHostRoutes {
					if b.API == r.api {
						b.Endpoint = w.provider.URL() + miniMaxHostPrefix(k, purpose, r)
					}
				}
				w.host.PutBinding(b)
			}
			w.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: ref, Version: "mm08-key-v1", AccountScopeID: account, Active: true, APIKey: ai.NewSecret(key.secret)})
		}
	}
	return &miniMaxHostWorld{limitWorld: limitWorld{world: w, bodies: bodies, probe: gauges}, rec: rec}
}

func miniMaxHostReply(t *testing.T, r miniMaxHostRoute, model string) provider.Reply {
	t.Helper()
	switch r.api {
	case ai.APIOpenAICompletions:
		return miniMaxChatReply(t, model)
	case ai.APIAnthropicMessages:
		return miniMaxMessagesReply(t, model)
	default:
		return minimaxResponsesReply(t, `{"type":"response.completed","response":{"id":"mm08-response","model":"`+model+`","status":"completed","output":[{"type":"message","id":"mm08-message","role":"assistant","content":[{"type":"output_text","text":"Host response."}]}],"usage":{"input_tokens":8,"output_tokens":4,"total_tokens":12}}}`)
	}
}
func miniMaxHostInvoke(ctx context.Context, w *miniMaxHostWorld, k tenantKey, p minimaxhost.Purpose, r miniMaxHostRoute, entry string, req ai.Request, hooks ai.Hooks, scope ai.CallScope) outcome {
	c := w.client.WithHooks(hooks)
	target := miniMaxHostTarget(p, r)
	var full ai.Options
	switch r.api {
	case ai.APIOpenAICompletions:
		full = ai.ChatOptions{MaxTokens: ai.Value(512)}
	case ai.APIOpenAIResponses:
		full = ai.ResponsesOptions{MaxTokens: ai.Value(512)}
	case ai.APIAnthropicMessages:
		full = ai.AnthropicOptions{MaxTokens: ai.Value(512)}
	}
	switch entry {
	case "stream":
		return drain(c.Stream(ctx, scope, target, req, full))
	case "simple-stream":
		return drain(c.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(512)}))
	case "simple-complete":
		res, err := c.CompleteSimple(ctx, scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(512)})
		return outcome{result: res, err: err}
	default:
		res, err := c.Complete(ctx, scope, target, req, full)
		return outcome{result: res, err: err}
	}
}
func miniMaxHostToolReply(t *testing.T, r miniMaxHostRoute, model string) provider.Reply {
	switch r.api {
	case ai.APIAnthropicMessages:
		return provider.SSE([][]byte{miniMaxMessagesFixture(t, "tools.sse", model)})
	case ai.APIOpenAICompletions:
		return chatEvents(t, []json.RawMessage{json.RawMessage(`{"id":"mm08-tools","model":"` + model + `","choices":[{"delta":{"reasoning_content":"Inspect the label.","tool_calls":[{"index":0,"id":"host-call","function":{"name":"inspect","arguments":"{\"label\":\"alpha\"}"}}]},"finish_reason":"tool_calls"}]}`)}, "")
	default:
		return minimaxResponsesReply(t, `{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"host-reason","summary":[],"content":[{"type":"reasoning_text","text":"Inspect the label."}]}}`, `{"type":"response.output_item.done","output_index":1,"item":{"type":"function_call","id":"host-item","call_id":"host-call","name":"inspect","arguments":"{\"label\":\"alpha\"}"}}`, `{"type":"response.completed","response":{"id":"mm08-tools","model":"`+model+`","status":"completed","output":[],"usage":{"input_tokens":8,"output_tokens":4,"total_tokens":12}}}`)
	}
}
