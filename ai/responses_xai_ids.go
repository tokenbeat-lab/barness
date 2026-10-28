package ai

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

// This spelling is an independent xAI persistence contract. It deliberately
// keeps a small codec beside the route instead of coupling saved IDs to
// another provider's independently evolving codec.
func xaiResponsesToolID(callID, itemID string) string {
	pair, _ := json.Marshal([2]string{callID, itemID})
	return "xar1|" + base64.RawURLEncoding.EncodeToString(pair)
}
func xaiResponsesIDPair(id string) (string, string, error) {
	invalid := errors.New("invalid xAI Responses tool ID")
	encoded, ok := strings.CutPrefix(id, "xar1|")
	if !ok {
		return "", "", invalid
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil {
		return "", "", invalid
	}
	var pair []json.RawMessage
	if json.Unmarshal(raw, &pair) != nil || len(pair) != 2 {
		return "", "", invalid
	}
	call, ok := rawString(pair[0])
	item, itemOK := rawString(pair[1])
	if !ok || call == "" || !itemOK {
		return "", "", invalid
	}
	return call, item, nil
}
func xaiResponsesCallID(id string) (string, error) {
	call, _, err := xaiResponsesIDPair(id)
	return call, err
}
func xaiResponsesHistoryToolID(id string, source AssistantMessage) string {
	if source.Provider == ProviderXAI && source.API == APIOpenAIResponses {
		call, _, err := xaiResponsesIDPair(id)
		if err != nil {
			return id
		}
		return xaiResponsesToolID(call, "")
	}
	// The complete source ID is correlation data. Never interpret a foreign
	// provider's separators or turn it into an account-scoped xAI item reference.
	sum := sha256.Sum256([]byte(id))
	return xaiResponsesToolID("xai_"+hex.EncodeToString(sum[:]), "")
}
