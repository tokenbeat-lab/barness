package ai

import (
	"strings"
	"unicode/utf16"
)

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

// normalizeIDPart replaces every character outside [A-Za-z0-9_-] with "_",
// keeps at most 64 characters and trims trailing underscores. Like pi's
// regular expression it works on UTF-16 code units, so a character outside
// the Basic Multilingual Plane becomes two underscores.
func normalizeIDPart(part string) string {
	var b strings.Builder
	for _, r := range part {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		case utf16.RuneLen(r) == 2:
			b.WriteString("__")
		default:
			b.WriteByte('_')
		}
	}
	s := b.String()
	return strings.TrimRight(s[:min(len(s), 64)], "_")
}
