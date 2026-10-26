package e2e

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func xaiPublicPair(t *testing.T, call, item string) string {
	t.Helper()
	return "xar1|" + base64.RawURLEncoding.EncodeToString(mustMarshal(t, [2]string{call, item}))
}
func xaiTrustedMessage(t *testing.T, content ...ai.AssistantContent) ai.AssistantMessage {
	t.Helper()
	state, err := ai.TrustNativeState(textScope(t.Name()), ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: "acct-xai", ProviderID: ai.ProviderXAI, API: ai.APIOpenAIResponses, ModelID: "grok-4.7"})
	if err != nil {
		t.Fatal(err)
	}
	return ai.AssistantMessage{Provider: ai.ProviderXAI, API: ai.APIOpenAIResponses, Model: "grok-4.7", StopReason: ai.StopReasonToolUse, NativeState: state, Content: content}
}

func TestXAIHistoryProjection(t *testing.T) {
	for _, mode := range []string{"trusted", "restored", "untrusted", "cross-account", "cross-model", "cross-api", "rotation"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-E08-history-"+mode)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			callA, callB := "same|a", "same"
			itemA, itemB := "条目|"+strings.Repeat("α", 100), "a|条目"
			a := ai.ToolCall{ID: xaiPublicPair(t, callA, itemA), Name: "probe", Arguments: ai.UncheckedArguments(`{"file_id":"business","store":true,"model":"business"}`)}
			b := ai.ToolCall{ID: xaiPublicPair(t, callB, itemB), Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
			native := `{"type":"reasoning","id":"reason|α","summary":[{"type":"summary_text","text":"visible reason"}],"content":[],"encrypted_content":"opaque-cipher","extra":{"precise":9007199254740993,"store":"business"}}`
			msg := xaiTrustedMessage(t, ai.Thinking{Thinking: "visible reason", Signature: native}, ai.Text{Text: "visible answer", Signature: `{"v":1,"id":"message|long-α","phase":"final_answer"}`}, a, b)
			if mode != "trusted" && mode != "rotation" {
				var env *ai.NativeStateEnvelope
				msg, env, _ = restore(t, persist(t, msg))
				if mode != "untrusted" {
					if mode == "cross-account" {
						env.AccountScopeID = "other-account"
					}
					if mode == "cross-model" {
						env.ModelID = "old-model"
					}
					if mode == "cross-api" {
						env.API = ai.APIOpenAICompletions
						msg.API = ai.APIOpenAICompletions
					}
					msg.NativeState, _ = ai.TrustNativeState(textScope(t.Name()), *env)
				}
			}
			if mode == "rotation" {
				w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: "cred-xai", Version: "v2", AccountScopeID: "acct-xai", Active: true, APIKey: ai.NewSecret(xaiKey.secret)})
			}
			req := xaiRequest()
			req.Messages = append(req.Messages, msg, ai.ToolResultText(a, "first", false), ai.ToolResultText(b, "second", false))
			before := persist(t, msg)
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("complete history accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			ev.Check("source remains immutable", string(before) == string(persist(t, msg)), "history changed")
			fields := bodyFields(t, w.provider.Requests()[0].Body)
			var input []map[string]json.RawMessage
			mustUnmarshal(t, fields["input"], &input)
			trusted := mode == "trusted" || mode == "restored" || mode == "rotation"
			nativeCount, calls, outputs := 0, 0, 0
			for _, item := range input {
				var typ string
				_ = json.Unmarshal(item["type"], &typ)
				switch typ {
				case "reasoning":
					nativeCount++
					ev.Check("entire native JSON survives", jsonEqual(mustMarshal(t, item), []byte(native)), "%s", mustMarshal(t, item))
				case "function_call":
					want := []string{callA, callB}[calls]
					var actual string
					mustUnmarshal(t, item["call_id"], &actual)
					if mode == "cross-api" {
						ev.Check("foreign source gets an independent opaque association", actual != "" && actual != want, "%s", actual)
					} else {
						ev.Check("call component remains exact", actual == want, "%s", actual)
					}
					_, hasID := item["id"]
					ev.Check("only trusted state retains account item ID", hasID == trusted, "%s", mustMarshal(t, item))
					calls++
				case "function_call_output":
					outputs++
				case "message":
					ev.Check("assistant projection excludes phase/status annotations", item["phase"] == nil && item["status"] == nil, "%s", mustMarshal(t, item))
				}
			}
			ev.Check("reasoning trust and tool pairing retained", nativeCount == map[bool]int{true: 1, false: 0}[trusted] && calls == 2 && outputs == 2, "native=%d calls=%d outputs=%d", nativeCount, calls, outputs)
		})
	}
}

func TestXAIHistoryRejectsBrokenAssociations(t *testing.T) {
	for _, kind := range []string{"missing-result", "orphan-result", "duplicate-result", "duplicate-call", "bad-opaque", "bad-reasoning", "resource-reasoning"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "XAI-E07-invalid-history-"+kind)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			call := ai.ToolCall{ID: xaiPublicPair(t, "call|α", "item|β"), Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
			msg := xaiTrustedMessage(t, call)
			req := xaiRequest()
			switch kind {
			case "bad-opaque":
				call.ID = "xar1|bad"
				msg.Content = []ai.AssistantContent{call}
			case "bad-reasoning":
				msg.Content = append(msg.Content, ai.Thinking{Signature: `{"type":"reasoning","summary":null}`})
			case "resource-reasoning":
				msg.Content = append(msg.Content, ai.Thinking{Signature: `{"type":"reasoning","summary":[],"file_id":"foreign"}`})
			case "duplicate-call":
				msg.Content = append(msg.Content, call)
			}
			req.Messages = append(req.Messages, msg)
			if kind != "missing-result" {
				req.Messages = append(req.Messages, ai.ToolResultText(call, "one", false))
			}
			if kind == "duplicate-result" {
				req.Messages = append(req.Messages, ai.ToolResultText(call, "other", false))
			}
			if kind == "orphan-result" {
				req.Messages = append(req.Messages, ai.ToolResultMessage{ToolCallID: "orphan", Content: []ai.ToolResultContent{ai.Text{Text: "orphan"}}})
			}
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
			o.record(ev)
			ev.Check("invalid source refused before credential", o.err != nil && w.host.ReadsFor(t.Name()).Credentials == 0 && len(w.provider.Requests()) == 0, "err=%v reads=%+v", o.err, w.host.ReadsFor(t.Name()))
		})
	}
}

func TestXAIWholeOpaqueIDs(t *testing.T) {
	for _, sc := range []struct{ name, call, item string }{
		{"unicode", "调用|甲%7C", "条目|乙%25"},
		{"long", strings.Repeat("调用%|α", 100), strings.Repeat("item%|β", 100)},
		{"empty-item", "plain|call", ""},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E07-opaque-pair-"+sc.name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t))
			call := ai.ToolCall{ID: xaiPublicPair(t, sc.call, sc.item), Name: "probe", Arguments: ai.UncheckedArguments(`{}`)}
			req := xaiRequest()
			req.Messages = append(req.Messages, xaiTrustedMessage(t, call), ai.ToolResultText(call, "result", false))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("opaque pair accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			var input []map[string]any
			mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
			var actualCall, actualItem, actualOutput string
			for _, item := range input {
				if item["type"] == "function_call" {
					actualCall, _ = item["call_id"].(string)
					actualItem, _ = item["id"].(string)
				}
				if item["type"] == "function_call_output" {
					actualOutput, _ = item["call_id"].(string)
				}
			}
			ev.Check("vendor identifiers remain exact without truncation or separators", actualCall == sc.call && actualItem == sc.item && actualOutput == sc.call, "call=%q item=%q result=%q", actualCall, actualItem, actualOutput)
		})
	}
}

func TestXAIReasoningPreservesExactJSON(t *testing.T) {
	ev := run.Case(t, "XAI-E08-native-exact-json")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	native := `{"type":"reasoning","summary":[],"encrypted_content":"opaque","extra":{"integer":9007199254740993,"large":1e400,"nested":[null,true,"原样"]}}`
	req := xaiRequest()
	req.Messages = append(req.Messages, xaiTrustedMessage(t, ai.Thinking{Signature: native}))
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("all legal native JSON numbers accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
		return
	}
	var input []map[string]json.RawMessage
	mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
	for _, item := range input {
		if string(item["type"]) == `"reasoning"` {
			var extra map[string]json.RawMessage
			mustUnmarshal(t, item["extra"], &extra)
			ev.Check("native numeric types and exact values survive", string(extra["integer"]) == "9007199254740993" && string(extra["large"]) == "1e400", "%s", item["extra"])
			return
		}
	}
	ev.Check("native reasoning retained", false, "missing native item")
}

func TestXAIReasoningPreservesNullCiphertext(t *testing.T) {
	ev := run.Case(t, "XAI-E08-native-null-ciphertext")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, xaiReply(t))
	native := `{"type":"reasoning","id":"r-null","summary":[{"type":"summary_text","text":"visible"}],"encrypted_content":null}`
	req := xaiRequest()
	req.Messages = append(req.Messages, xaiTrustedMessage(t, ai.Thinking{Thinking: "visible", Signature: native}))
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), req, nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("native absence is preserved as null", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
		return
	}
	var input []map[string]json.RawMessage
	mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
	for _, item := range input {
		if string(item["type"]) == `"reasoning"` {
			ev.Check("null never becomes empty or omitted ciphertext", string(item["encrypted_content"]) == "null", "%s", item["encrypted_content"])
			return
		}
	}
	ev.Check("native item retained", false, "missing")
}
