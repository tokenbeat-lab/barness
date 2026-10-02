package ai

import "slices"

// anthropicPayloadReferences are request fields that point at, or create,
// state or network targets on the vendor side: an execution container, MCP
// servers (with their own URLs and tokens) and server-side fallback models.
// The account may be shared by several tenants and the library cannot vouch
// for them, so a payload callback may not add them; barness-ai never sends
// them.
var anthropicPayloadReferences = []string{"container", "mcp_servers", "fallbacks"}

// authorizeAnthropicPayload checks a body a payload callback produced. It
// must still name the call's authorized model, declare only custom tools and
// the hosted tool types the binding allows, and carry no vendor-side
// reference: none of the fields above and no file reference (a file_id or a
// "file" source) anywhere in the system prompt or messages (ADR-0005).
func authorizeAnthropicPayload(m Model, hostedTools []string) func(map[string]any) string {
	return func(body map[string]any) string {
		if model, _ := body["model"].(string); model != m.ID {
			return "payload callback may not change the authorized model"
		}
		for _, field := range anthropicPayloadReferences {
			if _, ok := body[field]; ok {
				return "payload callback may not reference vendor-side state (" + field + ")"
			}
		}
		if tools, ok := body["tools"]; ok {
			list, _ := tools.([]any)
			if list == nil && tools != nil {
				return "payload callback produced malformed tools"
			}
			for _, t := range list {
				obj, isObject := t.(map[string]any)
				if !isObject {
					return "payload callback produced malformed tools"
				}
				// A tool without a type is a custom tool, as pi declares them.
				raw, typed := obj["type"]
				typ, _ := raw.(string)
				if typed && typ != "custom" && (typ == "" || !slices.Contains(hostedTools, typ)) {
					return "payload callback may only declare custom tools and the binding's hosted tools"
				}
			}
		}
		for _, field := range []string{"system", "messages"} {
			if ref := anthropicFileReference(body[field]); ref != "" {
				return "payload callback may not reference vendor-side state (" + ref + ")"
			}
		}
		return ""
	}
}

// anthropicFileReference finds a file_id or a file source anywhere in v.
func anthropicFileReference(v any) string {
	switch v := v.(type) {
	case []any:
		for _, e := range v {
			if ref := anthropicFileReference(e); ref != "" {
				return ref
			}
		}
	case map[string]any:
		if _, ok := v["file_id"]; ok {
			return "file_id"
		}
		if v["type"] == "file" {
			return "file source"
		}
		for _, e := range v {
			if ref := anthropicFileReference(e); ref != "" {
				return ref
			}
		}
	}
	return ""
}
