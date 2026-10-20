package ai

import (
	"encoding/json"
	"strings"
)

// Actual international M3 frames mirror the same text in both fields. Only
// exact text-entry equality is integrated; different representations must not
// silently discard reasoning. See ADR-0043 and the immutable research capture.
func miniMaxChatMirror(content, details json.RawMessage) bool {
	text, ok := rawString(content)
	if !ok {
		return false
	}
	var entries []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(details, &entries) != nil || len(entries) == 0 {
		return false
	}
	var mirrored strings.Builder
	for _, entry := range entries {
		if entry.Type != "reasoning.text" {
			return false
		}
		mirrored.WriteString(entry.Text)
	}
	return text == mirrored.String()
}

const miniMaxMirrorSignatureType = "minimax.chat.reasoning-mirror"

// The existing opaque signature records mirror provenance across persistence.
// Native detail entries stay intact inside it; no transport sees this wrapper.
func (p *chatParser) reasoningSignature() string {
	details := p.details.signature()
	if p.model.Provider == ProviderMiniMax && p.miniMaxReasoningFormat == "mirror" {
		return `{"type":"` + miniMaxMirrorSignatureType + `","reasoning_details":` + details + `}`
	}
	return details
}

func miniMaxMirroredDetails(signature string) (json.RawMessage, bool) {
	var envelope map[string]json.RawMessage
	if json.Unmarshal([]byte(signature), &envelope) != nil || len(envelope) != 2 {
		return nil, false
	}
	typ, _ := rawString(envelope["type"])
	if typ != miniMaxMirrorSignatureType {
		return nil, false
	}
	return chatReasoningDetails(string(envelope["reasoning_details"]))
}
