package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Unsigned and redacted are synthetic cases. Until ticket06 observes a valid
// service continuation, only signed native thinking is replayable here.
func TestMiniMaxMessagesNativeReplay(t *testing.T) {
	for _, shape := range []string{"signed", "absent", "empty", "redacted"} {
		for _, origin := range []string{"in-process", "restored", "untrusted", "account", "model", "tenant"} {
			t.Run(shape+"/"+origin, func(t *testing.T) {
				ev := run.Case(t, "MM04-native-"+shape+"-"+origin)
				w := newMiniMaxMessagesWorld(t, nil)
				script := miniMaxMessagesFixture(t, "tools.sse", "MiniMax-M3")
				enqueue(ev, w, provider.SSE([][]byte{script}), miniMaxMessagesReply(t, "MiniMax-M3"))
				req := miniMaxToolRequest()
				first := drain(w.client.Stream(ctxFor(t), textScope("mm04-native-first"), miniMaxMessagesTarget("MiniMax-M3"), req, nil))
				first.record(ev)
				if !ev.Check("source turn succeeds", first.err == nil && len(first.result.Message.Content) == 4, "%v", first.err) {
					return
				}
				source := first.result.Message
				thought := source.Content[0].(ai.Thinking)
				switch shape {
				case "absent", "empty":
					thought.Signature = ""
				case "redacted":
					thought.Redacted = true
					thought.Thinking = "[Reasoning redacted]"
					thought.Signature = "opaque-synthetic"
				}
				source.Content[0] = thought
				env, _ := source.NativeState.Envelope()
				switch origin {
				case "restored", "untrusted":
					source, _, _ = restore(t, persist(t, source))
					if origin == "restored" {
						var err error
						source.NativeState, err = ai.TrustNativeState(textScope("mm04-restored"), env)
						if err != nil {
							t.Fatal(err)
						}
					}
				case "account":
					env.AccountScopeID = "another-account"
					source.NativeState, _ = ai.TrustNativeState(textScope("mm04-account"), env)
				case "model":
					source.Model = "MiniMax-M3.1-Flash-Preview"
				case "tenant":
					env.TenantID = tenantB.tenant
					scope := textScope("mm04-tenant")
					scope.TenantID = tenantB.tenant
					source.NativeState, _ = ai.TrustNativeState(scope, env)
				}
				req.Messages = append(req.Messages, source)
				for _, b := range source.Content {
					if call, ok := b.(ai.ToolCall); ok {
						req.Messages = append(req.Messages, ai.ToolResultText(call, "Synthetic result.", false))
					}
				}
				second := drain(w.client.Stream(ctxFor(t), textScope("mm04-native-final"), miniMaxMessagesTarget("MiniMax-M3"), req, nil))
				ev.Record("continuation-result", second.result)
				ev.Record("continuation-error", errString(second.err))
				ev.Record("requests", w.provider.Requests())
				trusted := origin == "in-process" || origin == "restored"
				if shape == "redacted" && trusted {
					checkRejected(ev, second, ai.CodeInvalidRequest, ai.PhaseRequest)
					ev.Check("unverified opaque native state never sent", len(w.provider.Requests()) == 1 && len(second.result.Metadata.Attempts) == 0, "%+v", second.result.Metadata)
					return
				}
				if !ev.Check("safe continuation succeeds", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
					return
				}
				body := string(w.provider.Requests()[1].Body)
				ev.Check("only trusted signed thinking replays natively", strings.Contains(body, `"type":"thinking"`) == (shape == "signed" && trusted), "%s", body)
				ev.Check("opaque payload never becomes ordinary text", shape != "redacted" || !strings.Contains(body, "opaque-synthetic"), "%s", body)
				if shape == "signed" && !trusted {
					d := second.result.Metadata.NativeStateDowngrades
					ev.Check("native provenance downgrade remains explicit", d.NoEnvelope+d.AccountMismatch+d.CrossModel == 1, "%+v", d)
				}
			})
		}
	}
}

// Redacted results remain lossless, and their opaque shape cannot be sent as
// native state until its service contract is verified.
func TestMiniMaxMessagesRedactedOutput(t *testing.T) {
	ev := run.Case(t, "MM04-redacted-output")
	w := newMiniMaxMessagesWorld(t, nil)
	frames := []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"opaque-source","model":"MiniMax-M3"}}`), json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"redacted_thinking","data":"opaque-synthetic"}}`), json.RawMessage(`{"type":"content_block_stop","index":0}`), json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"max_tokens"}}`), json.RawMessage(`{"type":"message_stop"}`)}
	enqueue(ev, w, sseEvents(t, frames, provider.FramingLF))
	o := drain(w.client.Stream(ctxFor(t), textScope("mm04-redacted-output"), miniMaxMessagesTarget("MiniMax-M3"), miniMaxMessagesRequest(), nil))
	o.record(ev)
	if ev.Check("opaque output retained", o.err == nil && len(o.result.Message.Content) == 1, "%v", o.err) {
		b, ok := o.result.Message.Content[0].(ai.Thinking)
		ev.Check("opaque state retains kind and data", ok && b.Redacted && b.Signature == "opaque-synthetic", "%+v", b)
	}
}
