package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// These are explicitly synthetic Host histories and loopback replies. They
// exercise the public Client for every affected destination, not live support.
func TestGrokHistoryAllOldTargets(t *testing.T) {
	for _, route := range []string{"openai-responses", "deepseek-responses", "openai-chat", "deepseek-chat", "anthropic-messages", "google-gemini", "zhipu-chat", "zhipu-messages", "zai-chat", "zai-messages", "minimax-chat", "minimax-responses", "minimax-messages", "minimax-cn-chat", "minimax-cn-responses", "minimax-cn-messages", "kimi-cn-chat", "kimi-cn-responses", "kimi-cn-messages", "kimi-intl-chat", "kimi-intl-responses", "kimi-intl-messages"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "GROK-HISTORY-WIRE-"+route)
			w, b, model := grokHistoryWorld(t, route)
			reply, _ := regionalHostScript(t, b, model)
			if b.API == ai.APIGoogleGenerativeAI {
				reply = dataEvents(t, []json.RawMessage{json.RawMessage(`{"candidates":[{"content":{"role":"model","parts":[{"text":"Ready."}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":8,"candidatesTokenCount":4,"totalTokenCount":12},"modelVersion":"gemini-3.8-flash"}`)}, provider.FramingLF)
			}
			if b.ProviderID == ai.ProviderMoonshot || b.ProviderID == ai.ProviderMoonshotCN {
				reply, _ = kimiHostScript(t, b.API)
			}
			enqueue(ev, w, reply)
			scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: t.Name()}
			req, err := grokhistory.Request(scope)
			if err != nil {
				t.Fatal(err)
			}
			before := mustMarshal(t, req)
			o := drain(w.client.StreamSimple(ctxFor(t), scope, ai.Target{BindingID: b.BindingID, ModelID: model}, req, ai.SimpleOptions{MaxTokens: ai.Value(64)}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("one successful destination request", o.err == nil && o.result.Message.StopReason == ai.StopReasonStop && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			raw := w.provider.Requests()[0].Body
			ev.Check("actual target wire preserves collision and reuse associations", grokhistory.Verify(b.API, raw) == nil, "%v", grokhistory.Verify(b.API, raw))
			ev.Check("Host source remains immutable", string(before) == string(mustMarshal(t, req)), "source changed")
			m := o.result.Metadata
			ev.Check("resolved destination metadata and foreign downgrade", m.ProviderID == b.ProviderID && m.API == b.API && m.ModelID == model && m.AccountScopeID == b.AccountScopeID && m.NativeStateDowngrades == (ai.NativeStateDowngrades{CrossModel: 2}), "%+v", m)
			// Mutating a real captured request must fail the independently used
			// proof; a PASS label or an unrelated native-history case is insufficient.
			var body map[string]any
			mustUnmarshal(t, raw, &body)
			delete(body, "input")
			delete(body, "messages")
			delete(body, "contents")
			ev.Check("missing transcript cannot satisfy native proof", grokhistory.Verify(b.API, mustMarshal(t, body)) != nil, "accepted missing history")
		})
	}
}

func grokHistoryWorld(t *testing.T, route string) (*world, ai.Binding, string) {
	t.Helper()
	if len(route) > 5 && route[:5] == "kimi-" {
		w := newKimiHostWorld(t)
		p := ai.ProviderMoonshot
		if route[5:7] == "cn" {
			p = ai.ProviderMoonshotCN
		}
		for _, b := range w.bindings[p] {
			if grokRouteSuffix(route, b.API) {
				return w.world, b, "kimi-k3"
			}
		}
	}
	for _, p := range []ai.ProviderID{"zhipu", "zai", "minimax", "minimax-cn"} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			if route != string(p)+grokSuffix(api) {
				continue
			}
			w := newRegionalHostWorld(t)
			model := "MiniMax-M3"
			if p == "zhipu" || p == "zai" {
				model = "glm-5.3"
			}
			for _, b := range w.bindings[p] {
				if b.API == api {
					return w.world, b, model
				}
			}
		}
	}
	w := newWorld(t, tenantA)
	switch route {
	case "openai-responses":
		return w, primaryBinding(tenantA, w.provider.URL()), "gpt-5-mini"
	case "openai-chat":
		return w, chatBinding(tenantA, w.provider.URL()), "gpt-5-mini"
	case "deepseek-responses":
		return w, deepseekBinding(tenantA, w.provider.URL()), "deepseek-flash"
	case "deepseek-chat":
		return w, deepseekChatBinding(tenantA, w.provider.URL()), "deepseek-flash"
	case "anthropic-messages":
		return w, anthropicBinding(tenantA, w.provider.URL()), "claude-haiku-4-5"
	case "google-gemini":
		return w, geminiBinding(tenantA, w.provider.URL()), "gemini-3.8-flash"
	}
	t.Fatalf("unknown controlled route %s", route)
	return nil, ai.Binding{}, ""
}

func grokSuffix(api ai.API) string {
	switch api {
	case ai.APIOpenAICompletions:
		return "-chat"
	case ai.APIOpenAIResponses:
		return "-responses"
	default:
		return "-messages"
	}
}
func grokRouteSuffix(route string, api ai.API) bool {
	return route[len(route)-len(grokSuffix(api)):] == grokSuffix(api)
}
