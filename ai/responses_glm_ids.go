package ai

import "strings"

var glmResponsesIDEscape = strings.NewReplacer("%", "%25", "|", "%7C")
var glmResponsesIDUnescape = strings.NewReplacer("%7C", "|", "%25", "%")

// GLM IDs are opaque strings: neither call_id nor item id excludes '|'.
// Escape both components before joining so distinct native pairs never share
// a public ToolCall.ID. Percent must escape too; replacement decoding is one
// pass, preserving literal native %7C/%25 sequences and Unicode unchanged.
// Ordinary captured IDs keep their established call_id|item_id spelling.
func glmResponsesToolID(callID, itemID string) string {
	return glmResponsesIDEscape.Replace(callID) + "|" + glmResponsesIDEscape.Replace(itemID)
}

func glmResponsesCallID(publicID string) string {
	callID, _, _ := strings.Cut(publicID, "|")
	return glmResponsesIDUnescape.Replace(callID)
}

// A GLM target needs correlation, not OpenAI's restricted ID alphabet. Its
// own public IDs already encode the pair and remain valid across native-state
// downgrades (the native item id is never sent). Another protocol's whole
// opaque ID becomes a new call component, keeping calls/results bijective
// without interpreting that source's delimiter or account-bound item state.
func glmResponsesHistoryToolID(id string, source AssistantMessage) string {
	if source.Provider == ProviderZhipu && source.API == APIOpenAIResponses {
		return id
	}
	return glmResponsesToolID(id, "")
}
