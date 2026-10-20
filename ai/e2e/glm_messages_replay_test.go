package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func glmMessagesCapturedRequest(t *testing.T, c glmMessagesCapture) ai.Request {
	t.Helper()
	var body struct {
		Messages []struct{ Content []struct{ Text string } }
		Tools    []struct {
			Name, Description string
			InputSchema       json.RawMessage `json:"input_schema"`
		}
	}
	mustUnmarshal(t, c.request(t), &body)
	req := ai.Request{Messages: []ai.Message{ai.UserText(body.Messages[0].Content[0].Text)}}
	for _, tool := range body.Tools {
		req.Tools = append(req.Tools, ai.Tool{Name: tool.Name, Description: tool.Description, Parameters: tool.InputSchema})
	}
	return req
}

// These two independent real tool captures exercise the Host's validation,
// execution and continuation through all four public entries and omission/auto.
func TestGLMMessagesTools(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		first := glmMessagesCaptureByName(t, model+"-tool")
		last := glmMessagesCaptureByName(t, model+"-tool-continuation")
		for _, entry := range []string{"stream", "complete", "stream-simple", "complete-simple"} {
			for _, choice := range []string{"default", "auto"} {
				t.Run(model+"/"+entry+"/"+choice, func(t *testing.T) {
					ev := run.Case(t, "GLM07-tools-"+model+"-"+entry+"-"+choice)
					w := newGLMMessagesWorld(t, nil)
					enqueue(ev, w, first.reply(t), last.reply(t))
					req := glmMessagesCapturedRequest(t, first)
					target := ai.Target{BindingID: "glm-messages", ModelID: model}
					full := ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
					simple := ai.SimpleOptions{MaxTokens: ai.Value(4096)}
					if choice == "auto" {
						full.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "auto"})
						simple.ToolChoice = ai.Value(ai.ToolChoiceAuto)
					}
					invoke := func(id string) outcome {
						var o outcome
						switch entry {
						case "stream":
							o = drain(w.client.Stream(ctxFor(t), textScope(id), target, req, full))
						case "complete":
							o.result, o.err = w.client.Complete(ctxFor(t), textScope(id), target, req, full)
						case "stream-simple":
							o = drain(w.client.StreamSimple(ctxFor(t), textScope(id), target, req, simple))
						case "complete-simple":
							o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope(id), target, req, simple)
						}
						return o
					}
					o := invoke("glm07-tool-first")
					o.record(ev)
					if !ev.Check("native tool/signature fragments and order preserved", o.err == nil && jsonEqual(mustMarshal(t, o.result.Message.Content), first.Content) && o.result.Message.StopReason == ai.StopReasonToolUse, "%v %+v", o.err, o.result.Message.Content) {
						return
					}
					callIndex := len(o.result.Message.Content) - 1
					call, err := ai.ValidateToolCall(o.result.Message, callIndex, req.Tools)
					if !ev.Check("Host validates tool schema before execution", err == nil, "%v", err) {
						return
					}
					var args struct{ Label string }
					ev.Check("Host decodes fragmented arguments", call.Decode(&args) == nil && args.Label == "alpha", "%+v", args)
					tool := o.result.Message.Content[callIndex].(ai.ToolCall)
					req.Messages = append(req.Messages, o.result.Message, ai.ToolResultText(tool, "BRN42", false))
					second := invoke("glm07-tool-final")
					ev.Record("final-result", second.result)
					ev.Record("requests", w.provider.Requests())
					ev.Check("tool continuation completes with actual final content", second.err == nil && second.result.Message.StopReason == ai.StopReasonStop && jsonEqual(mustMarshal(t, second.result.Message.Content), last.Content), "%v %+v", second.err, second.result.Message.Content)
					ev.Check("same-model native provenance retained", second.result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "%+v", second.result.Metadata)
					if ev.Check("two actual attempts", len(w.provider.Requests()) == 2, "%+v", w.provider.Requests()) {
						for i, c := range []glmMessagesCapture{first, last} {
							want := bodyFields(t, c.request(t))
							if choice == "default" {
								delete(want, "tool_choice")
							}
							ev.Check("wire preserves verified schema/native history/result pairing", jsonEqual(w.provider.Requests()[i].Body, mustMarshal(t, want)), "%s", w.provider.Requests()[i].Body)
						}
					}
				})
			}
		}
	}
}

// Unsigned here is a deliberate Host transformation of actual signed output.
// Own probe 04 proved field absence is accepted; it did not observe unsigned
// provider output, empty-string signatures or redacted blocks.
func TestGLMMessagesNativeReplay(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		for _, signature := range []string{"signed", "unsigned"} {
			for _, origin := range []string{"in-process", "restored", "untrusted", "tenant", "account", "api", "model"} {
				t.Run(model+"/"+signature+"/"+origin, func(t *testing.T) {
					ev := run.Case(t, "GLM07-native-"+model+"-"+signature+"-"+origin)
					w := newGLMMessagesWorld(t, nil)
					name, opts := model+"-tool", ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
					lastName := model + "-tool-continuation"
					if model == "glm-5.2" {
						name, lastName, opts = model+"-thinking-enabled", model+"-thinking-continuation", glmMessagesEnabled()
					} else {
						opts.ToolChoice = ai.Value(ai.AnthropicToolChoice{Mode: "auto"})
					}
					first, last := glmMessagesCaptureByName(t, name), glmMessagesCaptureByName(t, lastName)
					enqueue(ev, w, first.reply(t), last.reply(t))
					req := glmMessagesCapturedRequest(t, first)
					target := ai.Target{BindingID: "glm-messages", ModelID: model}
					res, err := w.client.Complete(ctxFor(t), textScope("glm07-native-first"), target, req, opts)
					if !ev.Check("capture produces native thinking", err == nil, "%v", err) {
						return
					}
					m := res.Message
					if signature == "unsigned" {
						b := m.Content[0].(ai.Thinking)
						b.Signature = ""
						m.Content[0] = b
					}
					if origin != "in-process" {
						stored := persist(t, m)
						ev.Record("host-stored-record", json.RawMessage(stored))
						var env *ai.NativeStateEnvelope
						m, env, _ = restore(t, stored)
						if origin != "untrusted" {
							scope := textScope("glm07-native-restore")
							switch origin {
							case "tenant":
								env.TenantID, scope.TenantID = "other-synthetic-tenant", "other-synthetic-tenant"
							case "account":
								env.AccountScopeID = "other-synthetic-account"
							case "api":
								env.API = ai.APIOpenAIResponses
							case "model":
								env.ModelID = "other-model"
							}
							m.NativeState, err = ai.TrustNativeState(scope, *env)
							ev.Check("Host attests stored provenance", err == nil, "%v", err)
						}
					}
					before := persist(t, m)
					req.Messages = append(req.Messages, m)
					if model == "glm-5.3" {
						req.Messages = append(req.Messages, ai.ToolResultText(m.Content[len(m.Content)-1].(ai.ToolCall), "BRN42", false))
					} else {
						req.Messages = append(req.Messages, ai.UserText("What is that integer plus 1? Reply only with the integer."))
					}
					res2, err := w.client.Complete(ctxFor(t), textScope("glm07-native-final"), target, req, opts)
					ev.Record("result", res2)
					ev.Record("requests", w.provider.Requests())
					ev.Check("continuation succeeds and caller history unchanged", err == nil && string(before) == string(persist(t, m)), "%v", err)
					wantDowngrade := ai.NativeStateDowngrades{}
					switch origin {
					case "untrusted", "tenant":
						wantDowngrade.NoEnvelope = 1
					case "account":
						wantDowngrade.AccountMismatch = 1
					case "api", "model":
						wantDowngrade.CrossModel = 1
					}
					ev.Check("native envelope boundary applies equally to unsigned thinking", res2.Metadata.NativeStateDowngrades == wantDowngrade, "%+v", res2.Metadata.NativeStateDowngrades)
					if !ev.Check("two calls reached Provider", len(w.provider.Requests()) == 2, "%+v", w.provider.Requests()) {
						return
					}
					var wire struct {
						Messages []struct{ Content []map[string]any }
					}
					mustUnmarshal(t, w.provider.Requests()[1].Body, &wire)
					b := wire.Messages[1].Content[0]
					if wantDowngrade == (ai.NativeStateDowngrades{}) {
						thinking := m.Content[0].(ai.Thinking)
						ev.Check("native thinking stays thinking with exact text", b["type"] == "thinking" && b["thinking"] == thinking.Thinking, "%+v", b)
						_, present := b["signature"]
						ev.Check("signature preserved or omitted, never invented or empty", signature == "signed" && present && b["signature"] == thinking.Signature || signature == "unsigned" && !present, "%+v", b)
						var baseline map[string]any
						mustUnmarshal(t, last.request(t), &baseline)
						if signature == "unsigned" {
							delete(baseline["messages"].([]any)[1].(map[string]any)["content"].([]any)[0].(map[string]any), "signature")
						}
						ev.Check("native replay matches own successful baseline with only signature omission", jsonEqual(w.provider.Requests()[1].Body, mustMarshal(t, baseline)), "%s", w.provider.Requests()[1].Body)
					} else {
						ev.Check("downgrade emits visible text without native data", b["type"] == "text" && b["signature"] == nil && b["thinking"] == nil && !strings.Contains(string(w.provider.Requests()[1].Body), "other-synthetic-account"), "%+v", b)
					}
				})
			}
		}
	}
}
