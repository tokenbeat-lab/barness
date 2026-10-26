// Package grokhistory defines the fixed synthetic foreign-history qualification
// probe shared by the 22 existing destinations affected by xAI ID allocation.
// It uses only public Host messages. No vendor capture or private allocator is
// used, and successful controlled execution is never live qualification.
package grokhistory

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

const Scenario = "xai-origin-history"

const (
	visibleReasoning = "SYNTHETIC_XAI_VISIBLE_REASONING"
	visibleAnswer    = "SYNTHETIC_XAI_VISIBLE_ANSWER"
	cipher           = "SYNTHETIC_XAI_FOREIGN_CIPHER"
	reasoningID      = "SYNTHETIC_XAI_FOREIGN_REASONING_RESOURCE"
	messageID        = "SYNTHETIC_XAI_FOREIGN_MESSAGE_RESOURCE"
)

var resultMarkers = []string{"SYNTHETIC_XAI_RESULT_FIRST", "SYNTHETIC_XAI_RESULT_SECOND", "SYNTHETIC_XAI_RESULT_REUSED"}

// Applies is the reviewed finite destination scope, not a provider registry.
func Applies(combo string) bool {
	switch combo {
	case "openai-responses", "deepseek-responses", "openai-chat", "deepseek-chat", "anthropic-messages", "google-gemini",
		"zhipu-chat", "zhipu-messages", "zai-chat", "zai-messages", "minimax-chat", "minimax-responses", "minimax-messages",
		"minimax-cn-chat", "minimax-cn-responses", "minimax-cn-messages", "kimi-cn-chat", "kimi-cn-responses", "kimi-cn-messages", "kimi-intl-chat", "kimi-intl-responses", "kimi-intl-messages":
		return true
	}
	return false
}

// VerifyHost checks the recorded immutable public input against this fixed
// synthetic fixture. Trust itself is deliberately not JSON serializable.
func VerifyHost(raw []byte) bool {
	req, _ := Request(ai.CallScope{TenantID: "synthetic-verification"})
	want, _ := json.Marshal(req)
	var gotJSON, wantJSON any
	if json.Unmarshal(raw, &gotJSON) != nil || json.Unmarshal(want, &wantJSON) != nil {
		return false
	}
	a, _ := json.Marshal(gotJSON)
	b, _ := json.Marshal(wantJSON)
	return bytes.Equal(a, b)
}

// Request combines two same-turn opaque pairs with colliding legacy prefixes
// and reuses the identical complete first pair in another assistant turn.
// Long common prefixes cover sanitizer/truncation rules; repeated full identity
// also exercises destinations with injective opaque association codecs.
func Request(scope ai.CallScope) (ai.Request, error) {
	trusted, err := ai.TrustNativeState(scope, ai.NativeStateEnvelope{TenantID: scope.TenantID, AccountScopeID: "synthetic-xai-source-account", ProviderID: ai.ProviderXAI, API: ai.APIOpenAIResponses, ModelID: "grok-4.7"})
	if err != nil {
		return ai.Request{}, err
	}
	pair := func(suffix string) string {
		raw, _ := json.Marshal([]string{strings.Repeat("same-prefix_", 8) + suffix, "SYNTHETIC_XAI_FOREIGN_ITEM_" + suffix})
		return "xar1|" + base64.RawURLEncoding.EncodeToString(raw)
	}
	a := ai.ToolCall{ID: pair("a+b"), Name: "inspect_xai_history", Arguments: ai.UncheckedArguments(`{"label":"first"}`)}
	b := ai.ToolCall{ID: pair("a/b"), Name: a.Name, Arguments: ai.UncheckedArguments(`{"label":"second"}`)}
	assistant := func(blocks ...ai.AssistantContent) ai.AssistantMessage {
		return ai.AssistantMessage{Provider: ai.ProviderXAI, API: ai.APIOpenAIResponses, Model: "grok-4.7", NativeState: trusted, StopReason: ai.StopReasonToolUse, Content: blocks}
	}
	return ai.Request{Messages: []ai.Message{
		ai.UserText("Inspect the synthetic stored history."),
		assistant(ai.Thinking{Thinking: visibleReasoning, Signature: `{"type":"reasoning","id":"` + reasoningID + `","summary":[],"encrypted_content":"` + cipher + `"}`}, ai.Text{Text: visibleAnswer, Signature: `{"v":1,"id":"` + messageID + `"}`}, a, b),
		ai.ToolResultText(a, resultMarkers[0], false), ai.ToolResultText(b, resultMarkers[1], false),
		ai.UserText("Repeat the first inspection in another turn."), assistant(a), ai.ToolResultText(a, resultMarkers[2], false),
		ai.UserText("The tools have finished. Reply with exactly history ready, without calling tools."),
	}, Tools: []ai.Tool{{Name: a.Name, Description: "Inspect a synthetic stored history label.", Parameters: json.RawMessage(`{"type":"object","properties":{"label":{"type":"string"}},"required":["label"],"additionalProperties":false}`)}}}, nil
}
