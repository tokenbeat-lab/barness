package supportmatrix

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
)

// Rebuild public Host inputs instead of replacing the encoded payload. Native
// trust is restored only from the fixed same-bundle predecessor, never by
// searching for a convenient successful result or trusting serialized handles.
func kimiRegionalReplayRequest(root *os.Root, r Report, s ScenarioResult, meta ai.CallMetadata) (ai.Request, error) {
	raw, e := regionalRead(root, s.CaseID+"/request.json")
	if e != nil {
		return ai.Request{}, e
	}
	_, kind, _ := strings.Cut(s.ID, "/")
	if kind == grokhistory.Scenario {
		if !grokhistory.VerifyHost(raw) {
			return ai.Request{}, refuse("Kimi fixed foreign Host input differs")
		}
		return grokhistory.Request(ai.CallScope{TenantID: meta.TenantID})
	}
	var input struct {
		SystemPrompt string
		Messages     []json.RawMessage
		Tools        []ai.Tool
	}
	if regionalDecode(raw, &input) != nil {
		return ai.Request{}, refuse("Kimi public replay input invalid")
	}
	out := ai.Request{SystemPrompt: input.SystemPrompt, Tools: input.Tools}
	for _, raw := range input.Messages {
		var fields map[string]json.RawMessage
		if regionalDecode(raw, &fields) != nil {
			return ai.Request{}, refuse("Kimi public message invalid")
		}
		switch {
		case fields["api"] != nil:
			if kind != "native-history" && kind != "tool-result" {
				return ai.Request{}, refuse("unowned Kimi assistant input")
			}
			var stored regionalMessage
			if regionalDecode(raw, &stored) != nil {
				return ai.Request{}, refuse("Kimi stored assistant invalid")
			}
			a, e := stored.assistant()
			if e != nil {
				return ai.Request{}, e
			}
			sourceCase := RegionalCaseID(r.Combo, s.Model+"/tool-call")
			if kind == "native-history" {
				sourceCase = RegionalNativeSourceCase(r.Combo, s.Model)
			}
			priorRaw, e := regionalRead(root, sourceCase+"/result.json")
			var prior regionalResult
			if e != nil || regionalDecode(priorRaw, &prior) != nil {
				return ai.Request{}, refuse("Kimi fixed replay predecessor absent")
			}
			previous, e := prior.Message.assistant()
			m := prior.Metadata
			if e != nil || !reflect.DeepEqual(a, previous) || m.TenantID != meta.TenantID || m.AccountScopeID != meta.AccountScopeID || m.RequestID == meta.RequestID || m.BindingID != r.Combo || m.ProviderID != ai.ProviderID(r.Provider) || m.API != ai.API(r.API) || m.ModelID != s.Model {
				return ai.Request{}, refuse("Kimi native replay predecessor provenance differs")
			}
			a.NativeState, e = ai.TrustNativeState(ai.CallScope{TenantID: meta.TenantID}, ai.NativeStateEnvelope{TenantID: m.TenantID, AccountScopeID: m.AccountScopeID, ProviderID: m.ProviderID, API: m.API, ModelID: m.ModelID})
			if e != nil {
				return ai.Request{}, e
			}
			out.Messages = append(out.Messages, a)
		case fields["toolCallId"] != nil:
			var m struct {
				ToolCallID, ToolName string
				IsError              bool
				Content              []json.RawMessage
			}
			if regionalDecode(raw, &m) != nil || kind != "tool-result" {
				return ai.Request{}, refuse("Kimi tool-result input invalid")
			}
			message := ai.ToolResultMessage{ToolCallID: m.ToolCallID, ToolName: m.ToolName, IsError: m.IsError}
			for _, raw := range m.Content {
				block, e := kimiReplayText(raw)
				if e != nil {
					return ai.Request{}, e
				}
				message.Content = append(message.Content, block)
			}
			out.Messages = append(out.Messages, message)
		default:
			var m struct{ Content []json.RawMessage }
			if regionalDecode(raw, &m) != nil {
				return ai.Request{}, refuse("Kimi user input invalid")
			}
			user := ai.UserMessage{}
			for _, raw := range m.Content {
				block, e := kimiReplayText(raw)
				if e != nil {
					return ai.Request{}, e
				}
				user.Content = append(user.Content, block)
			}
			out.Messages = append(out.Messages, user)
		}
	}
	return out, nil
}
func kimiReplayText(raw []byte) (ai.Text, error) {
	var fields map[string]json.RawMessage
	var text ai.Text
	if regionalDecode(raw, &fields) != nil || regionalDecode(raw, &text) != nil || fields["text"] == nil {
		return text, refuse("Kimi captured text input invalid")
	}
	for key := range fields {
		if key != "text" && key != "textSignature" {
			return text, refuse("Kimi captured text has unsupported fields")
		}
	}
	return text, nil
}
