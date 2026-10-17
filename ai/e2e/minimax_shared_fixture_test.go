package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"strings"
	"testing"
)

var miniMaxSharedA = tenantKey{tenant: "tenant-a", secret: "synthetic-minimax-standard-a", alias: "key:minimax-standard-a@v1"}
var miniMaxSharedB = tenantKey{tenant: "tenant-b", secret: "synthetic-minimax-controlled-plan-b", alias: "key:minimax-controlled-plan-b@v1"}

func init() { knownKeys = append(knownKeys, miniMaxSharedA, miniMaxSharedB) }

type miniMaxSharedRoute struct {
	name string
	api  ai.API
}

var miniMaxSharedRoutes = []miniMaxSharedRoute{{"chat", ai.APIOpenAICompletions}, {"responses", ai.APIOpenAIResponses}, {"messages", ai.APIAnthropicMessages}}

func (r miniMaxSharedRoute) prefix(k tenantKey) string { return "/" + k.tenant + "/minimax-" + r.name }
func (r miniMaxSharedRoute) target(model string) ai.Target {
	return ai.Target{BindingID: "minimax-" + r.name, ModelID: model}
}
func (r miniMaxSharedRoute) options() ai.Options {
	switch r.name {
	case "chat":
		return ai.ChatOptions{SessionID: "same-session"}
	case "responses":
		return ai.ResponsesOptions{SessionID: "same-session"}
	default:
		return ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
	}
}
func newMiniMaxSharedWorld(t *testing.T, configure func(*ai.Config)) *mixedWorld {
	t.Helper()
	w := newMixedWorld(t, func(c *ai.Config) {
		c.Policy = validPolicy()
		c.Policy.MaxQueuedObservations = 512
		candidate := minimax.CandidateCatalog()
		c.Catalog.Version += "-minimax-candidate"
		c.Catalog.Models = append(c.Catalog.Models, candidate.Models...)
		if configure != nil {
			configure(c)
		}
	}, tenantA, tenantB)
	for _, k := range []tenantKey{miniMaxSharedA, miniMaxSharedB} {
		for _, r := range miniMaxSharedRoutes {
			var b ai.Binding
			switch r.name {
			case "chat":
				b = minimax.ChatBinding(k.tenant, "acct-"+k.alias, "same-credential", w.provider.URL()+r.prefix(k))
			case "responses":
				b = minimax.ResponsesBinding(k.tenant, "acct-"+k.alias, "same-credential", w.provider.URL()+r.prefix(k))
			default:
				b = minimax.MessagesBinding(k.tenant, "acct-"+k.alias, "same-credential", w.provider.URL()+r.prefix(k))
			}
			w.host.PutBinding(b)
			w.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "v1", Active: true, APIKey: ai.NewSecret(k.secret)})
		}
	}
	return w
}

// These protocol fixtures evolve independently from GLM: MiniMax uses split
// reasoning/adaptive controls and exact raw Responses projection/signed Messages.
func (r miniMaxSharedRoute) script(t *testing.T, model string) (ai.Request, provider.Reply, provider.Reply, []byte) {
	t.Helper()
	req := miniMaxToolRequest()
	var first, last provider.Reply
	switch r.name {
	case "chat":
		frames := []json.RawMessage{
			json.RawMessage(`{"choices":[{"delta":{"reasoning_content":"Check both.","content":"Looking.","tool_calls":[{"index":0,"id":"call-alpha","function":{"name":"inspect","arguments":"{\"label\":\"alpha\"}"}},{"index":1,"id":"call-beta","function":{"name":"inspect","arguments":"{\"label\":\"beta\"}"}}]},"finish_reason":"tool_calls"}]}`),
			json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":3}}}`),
		}
		first = chatEvents(t, frames, "")
		last = miniMaxChatReply(t, model)
	case "responses":
		first = minimaxResponsesReply(t, `{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"reasoning-mm05","summary":[],"content":[{"type":"reasoning_text","text":"Check both."}]}}`, `{"type":"response.output_item.done","output_index":1,"item":{"type":"function_call","id":"item-alpha","call_id":"call-alpha","name":"inspect","arguments":"{\"label\":\"alpha\"}"}}`, `{"type":"response.output_item.done","output_index":2,"item":{"type":"function_call","id":"item-beta","call_id":"call-beta","name":"inspect","arguments":"{\"label\":\"beta\"}"}}`, strings.ReplaceAll(minimaxResponsesCompleted, "MiniMax-M3", model))
		last = minimaxResponsesReply(t, strings.ReplaceAll(minimaxResponsesCompleted, "MiniMax-M3", model))
	default:
		first = provider.SSE([][]byte{miniMaxMessagesFixture(t, "tools.sse", model)})
		last = miniMaxMessagesReply(t, model)
	}
	raw := mustMarshal(t, map[string]any{"synthetic": true, "protocol": r.name, "model": model, "first": first, "last": last})
	return req, first, last, raw
}
func miniMaxSharedRecord(ev *evidence.Case, w *mixedWorld, o outcome) {
	o.record(ev)
	w.record(ev, mixedOutcome{metadata: o.result.Metadata, usage: o.result.Message.Usage, chat: o.result, err: o.err})
	wire := observationText(ev.T(), w.rec.all())
	for _, secret := range []string{"Check both.", "Synthetic reasoning.", "Synthetic input.", "MM05-result", miniMaxSharedA.secret, miniMaxSharedB.secret} {
		ev.Check("Observer excludes private payload", !strings.Contains(wire, secret), "private content entered observation")
	}
}
func miniMaxSharedContinue(req ai.Request, m ai.AssistantMessage) ai.Request {
	req.Messages = append(req.Messages, m)
	for _, block := range m.Content {
		if c, ok := block.(ai.ToolCall); ok {
			req.Messages = append(req.Messages, ai.ToolResultText(c, "MM05-result", false))
		}
	}
	return req
}
func miniMaxSharedOneTerminal(ev *evidence.Case, o outcome) {
	n := 0
	for _, e := range o.events {
		if e.Type() == "done" || e.Type() == "error" {
			n++
		}
	}
	ev.Check("exactly one terminal", n == 1, "%v", eventSummary(o.events))
}
