package ai

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"strings"
)

// MiniMax item/call IDs are opaque. A versioned JSON pair has no delimiter
// ambiguity, preserves Unicode and long values, and makes no fc_ assumption.
// The pipe also marks provider item state for the shared provenance boundary.
func miniMaxResponsesToolID(callID, itemID string) string {
	pair, _ := json.Marshal([2]string{callID, itemID})
	return "mmr1|" + base64.RawURLEncoding.EncodeToString(pair)
}
func miniMaxResponsesCallID(id string) (string, error) {
	encoded, ok := strings.CutPrefix(id, "mmr1|")
	if !ok {
		return "", errors.New("invalid MiniMax tool ID")
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	var pair []string
	if json.Unmarshal(raw, &pair) != nil || len(pair) != 2 || pair[0] == "" {
		return "", errors.New("invalid MiniMax tool ID")
	}
	return pair[0], nil
}
func miniMaxResponsesHistoryToolID(id string, source AssistantMessage) string {
	if source.Provider == ProviderMiniMaxCN && source.API == APIOpenAIResponses {
		return id
	}
	return miniMaxResponsesToolID(id, "")
}

type miniMaxReasoningPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type miniMaxInputReasoning struct {
	Type    string                 `json:"type"`
	Summary []miniMaxReasoningPart `json:"summary"`
}

// Official snapshot 2026-10-18: output.content uses reasoning_text whereas
// input.summary uses summary_text. This is a lossless field/part projection,
// never a generated summary. The original output stays in Thinking.Signature.
// ADR-0039 records that real acceptance must independently prove continuation.
func miniMaxResponsesReasoning(b Thinking) (miniMaxInputReasoning, error) {
	var raw map[string]json.RawMessage
	invalid := errors.New("MiniMax reasoning state is not a documented text item")
	if b.Redacted || json.Unmarshal([]byte(b.Signature), &raw) != nil {
		return miniMaxInputReasoning{}, invalid
	}
	for key := range raw {
		if !slices.Contains([]string{"type", "id", "status", "content", "summary"}, key) {
			return miniMaxInputReasoning{}, invalid
		}
	}
	var typ string
	if json.Unmarshal(raw["type"], &typ) != nil || typ != "reasoning" {
		return miniMaxInputReasoning{}, invalid
	}
	out := miniMaxInputReasoning{Type: "reasoning", Summary: []miniMaxReasoningPart{}}
	for _, field := range []string{"content", "summary"} {
		value, ok := raw[field]
		if !ok {
			continue
		}
		var parts []map[string]json.RawMessage
		if json.Unmarshal(value, &parts) != nil || string(value) == "null" {
			return miniMaxInputReasoning{}, invalid
		}
		for _, part := range parts {
			var kind, text string
			want := "reasoning_text"
			if field == "summary" {
				want = "summary_text"
			}
			if len(part) != 2 || json.Unmarshal(part["type"], &kind) != nil || kind != want || json.Unmarshal(part["text"], &text) != nil || string(part["text"]) == "null" {
				return miniMaxInputReasoning{}, invalid
			}
			out.Summary = append(out.Summary, miniMaxReasoningPart{Type: "summary_text", Text: text})
		}
	}
	if len(out.Summary) == 0 && b.Thinking != "" {
		out.Summary = append(out.Summary, miniMaxReasoningPart{Type: "summary_text", Text: b.Thinking})
	}
	return out, nil
}
