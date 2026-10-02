package ai

import "slices"

// authorizeChatPayload checks a body a payload callback produced. It must
// still name the call's authorized model, keep the tenant-scoped cache key
// and stateless storage, declare only function tools (as tools or the
// legacy functions) and the hosted tool types the binding allows, and
// reference nothing stored on the vendor account (ADR-0005, ADR-0013):
// hosted web search, named by its field web_search_options, needs the
// binding's allowance like any hosted tool; a file_id or a stored audio
// response anywhere in the messages is refused, since the account may be
// shared by several tenants and the library cannot tell whose they are.
func authorizeChatPayload(m Model, cacheKey string, hostedTools []string) func(map[string]any) string {
	return func(body map[string]any) string {
		if model, _ := body["model"].(string); model != m.ID {
			return "payload callback may not change the authorized model"
		}
		key, hasKey := body["prompt_cache_key"]
		if (cacheKey == "" && hasKey) || (cacheKey != "" && key != cacheKey) {
			return "payload callback may not change the prompt cache key"
		}
		if store, ok := body["store"]; ok && store != false {
			return "payload callback may not store the response on the vendor"
		}
		if _, ok := body["web_search_options"]; ok && !slices.Contains(hostedTools, "web_search_options") {
			return "payload callback may only declare function tools and the binding's hosted tools"
		}
		if tools, ok := body["tools"]; ok {
			list, _ := tools.([]any)
			if list == nil && tools != nil {
				return "payload callback produced malformed tools"
			}
			for _, t := range list {
				obj, _ := t.(map[string]any)
				typ, _ := obj["type"].(string)
				if typ != "function" && (typ == "" || !slices.Contains(hostedTools, typ)) {
					return "payload callback may only declare function tools and the binding's hosted tools"
				}
			}
		}
		if chatStoredReference(body["messages"]) {
			return "payload callback may not reference vendor-side state (file_id or stored audio)"
		}
		return ""
	}
}

// chatStoredReference finds a file_id, or an assistant message's audio
// reference (an earlier audio response stored by id), anywhere in v.
func chatStoredReference(v any) bool {
	switch v := v.(type) {
	case []any:
		return slices.ContainsFunc(v, chatStoredReference)
	case map[string]any:
		if _, ok := v["file_id"]; ok {
			return true
		}
		if audio, ok := v["audio"].(map[string]any); ok && v["role"] == "assistant" {
			if _, ok := audio["id"]; ok {
				return true
			}
		}
		for _, e := range v {
			if chatStoredReference(e) {
				return true
			}
		}
	}
	return false
}
