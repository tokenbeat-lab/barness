package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiMessagesToolsAndNative(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, signature := range []string{"signed", "unsigned"} {
				t.Run(string(market)+"/"+entry+"/"+signature, func(t *testing.T) {
					ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K06-K07-"+entry+"-"+signature)
					w := newKimiMessagesWorld(t, market, nil)
					script := string(miniMaxMessagesFixture(t, "tools.sse", "kimi-k3"))
					script = strings.ReplaceAll(script, "call-alpha", "opaque|/α-1")
					if signature == "unsigned" {
						script = strings.ReplaceAll(script, `"signature":"sig-"`, `"signature":""`)
						script = strings.ReplaceAll(script, `"signature":"part1-"`, `"signature":""`)
						script = strings.ReplaceAll(script, `"signature":"part2"`, `"signature":""`)
					}
					enqueue(ev, w, provider.SSE([][]byte{[]byte(script)}), kimiMessagesReply())
					req := miniMaxToolRequest()
					invoke := func() outcome {
						opts := ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "any"})}
						if len(req.Messages) > 1 {
							opts.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "auto"})
						}
						var o outcome
						switch entry {
						case "stream":
							o = drain(w.client.Stream(ctxFor(t), textScope("kimi-history"), kimiMessagesTarget, req, opts))
						case "complete":
							o.result, o.err = w.client.Complete(ctxFor(t), textScope("kimi-history"), kimiMessagesTarget, req, opts)
						case "stream-simple":
							o = drain(w.client.StreamSimple(ctxFor(t), textScope("kimi-history"), kimiMessagesTarget, req, ai.SimpleOptions{Reasoning: ai.ThinkingMedium, ToolChoice: ai.Value(ai.ToolChoiceAuto), CacheRetention: ai.CacheRetentionLong}))
						case "complete-simple":
							o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope("kimi-history"), kimiMessagesTarget, req, ai.SimpleOptions{Reasoning: ai.ThinkingXHigh, CacheRetention: ai.CacheRetentionLong})
						}
						return o
					}
					first := invoke()
					first.record(ev)
					if !ev.Check("mixed tool turn succeeds", first.err == nil && first.result.Message.StopReason == ai.StopReasonToolUse && len(first.result.Message.Content) == 4, "%v %+v", first.err, first.result.Message.Content) {
						return
					}
					thought := first.result.Message.Content[0].(ai.Thinking)
					expected := "sig-part1-part2"
					if signature == "unsigned" {
						expected = ""
					}
					ev.Check("real thinking and signatures retained", thought.Thinking == "Synthetic reasoning." && thought.Signature == expected, "%+v", thought)
					req.Messages = append(req.Messages, first.result.Message)
					for _, block := range first.result.Message.Content {
						if call, ok := block.(ai.ToolCall); ok {
							req.Messages = append(req.Messages, ai.ToolResultText(call, "Synthetic result.", false))
						}
					}
					second := invoke()
					ev.Record("continuation", second.result)
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("tool round trip completes", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
						return
					}
					body := string(w.provider.Requests()[1].Body)
					ev.Check("thinking kind preserved without fabricated signature", strings.Contains(body, `"type":"thinking"`) && (strings.Contains(body, `"signature"`) == (signature == "signed")), "%s", body)
					if strings.Contains(entry, "simple") {
						f := bodyFields(t, w.provider.Requests()[0].Body)
						expected := "high"
						if entry == "complete-simple" {
							expected = "max"
						}
						ev.Check("canonical effort clamps and long cache stays top level", jsonEqual(f["output_config"], []byte(`{"effort":"`+expected+`"}`)) && jsonEqual(f["cache_control"], []byte(`{"type":"ephemeral","ttl":"1h"}`)), "%s", w.provider.Requests()[0].Body)
					}
				})
			}
		}
	}
}

func TestKimiMessagesUnsignedProvenance(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, origin := range []string{"trusted", "restored", "untrusted", "account", "market", "protocol", "tenant"} {
			t.Run(string(market)+"/"+origin, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K08-unsigned-"+origin)
				w := newKimiMessagesWorld(t, market, nil)
				env := ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: "acct-kimi-" + string(market), ProviderID: market, API: ai.APIAnthropicMessages, ModelID: "kimi-k3"}
				source := ai.AssistantMessage{Provider: market, API: ai.APIAnthropicMessages, Model: "kimi-k3", StopReason: ai.StopReasonStop, Content: []ai.AssistantContent{ai.Thinking{Thinking: "Synthetic unsigned reasoning."}}}
				scope := textScope("kimi-origin")
				switch origin {
				case "account":
					env.AccountScopeID = "another-account"
				case "market":
					if market == ai.ProviderMoonshot {
						env.ProviderID = ai.ProviderMoonshotCN
					} else {
						env.ProviderID = ai.ProviderMoonshot
					}
				case "protocol":
					env.API = ai.APIOpenAIResponses
				case "tenant":
					env.TenantID = tenantB.tenant
					scope.TenantID = tenantB.tenant
				}
				if origin != "untrusted" {
					source.NativeState, _ = ai.TrustNativeState(scope, env)
				}
				if origin == "restored" {
					source, _, _ = restore(t, persist(t, source))
					source.NativeState, _ = ai.TrustNativeState(scope, env)
				}
				req := kimiMessagesRequest()
				req.Messages = append(req.Messages, source, ai.UserText("Continue."))
				enqueue(ev, w, kimiMessagesReply())
				o := drain(w.client.Stream(ctxFor(t), textScope("kimi-replay"), kimiMessagesTarget, req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("history continues safely", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
					return
				}
				body := string(w.provider.Requests()[0].Body)
				trusted := origin == "trusted" || origin == "restored"
				ev.Check("unsigned native wire kind requires exact trusted origin", strings.Contains(body, `"type":"thinking"`) == trusted && strings.Contains(body, "Synthetic unsigned reasoning.") && !strings.Contains(body, `"signature"`), "%s", body)
				d := o.result.Metadata.NativeStateDowngrades
				ev.Check("downgrade counted", d.NoEnvelope+d.AccountMismatch+d.CrossModel == map[bool]int{true: 0, false: 1}[trusted], "%+v", d)
			})
		}
	}
}
