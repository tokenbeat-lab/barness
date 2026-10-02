package ai

import "strings"

// responsesToolCallProviders are the target providers whose Responses tool
// call ids keep pi-ai's "<call_id>|<item id>" form (pi
// OPENAI_TOOL_CALL_PROVIDERS). pi also lists openai-codex and opencode, which
// barness-ai does not serve. Any other target, such as DeepSeek, gets the
// whole id sanitized into one part.
var responsesToolCallProviders = map[ProviderID]bool{ProviderOpenAI: true}

func (responsesAdapter) historyRules(m Model) historyRules {
	return historyRules{
		midConvoSystem: m.Compat.SupportsMidConvoSystemMessages,
		normalizeToolCallID: func(id string, source AssistantMessage) string {
			return responsesToolCallID(m, id, source)
		},
	}
}

// responsesToolCallID ports pi-ai's Responses normalizeToolCallId for a call
// replayed by the cross-model rules. Both parts are sanitized to OpenAI's id
// alphabet; an item id from another provider or API is replaced by a stable
// hash of it, since the target never issued it; and the item id always
// starts with "fc".
func responsesToolCallID(target Model, id string, source AssistantMessage) string {
	if !responsesToolCallProviders[target.Provider] || !strings.Contains(id, "|") {
		return normalizeIDPart(id)
	}
	// pi destructures split("|"): a third part is ignored.
	parts := strings.Split(id, "|")
	callID, itemID := normalizeIDPart(parts[0]), ""
	if source.Provider != target.Provider || source.API != target.API {
		itemID = "fc_" + shortHash(parts[1])
		itemID = itemID[:min(len(itemID), 64)]
	} else {
		itemID = normalizeIDPart(parts[1])
	}
	if !strings.HasPrefix(itemID, "fc_") {
		itemID = normalizeIDPart("fc_" + itemID)
	}
	return callID + "|" + itemID
}

// normalizeIDPart is sanitizeToolCallID with trailing underscores trimmed,
// as pi's Responses normalization does.
func normalizeIDPart(part string) string {
	return strings.TrimRight(sanitizeToolCallID(part), "_")
}
