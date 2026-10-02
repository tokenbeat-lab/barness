package ai

import "slices"

// geminiPayloadReferences are request fields and tool kinds that point at
// state or network targets on the vendor side: content cached on the
// account, MCP servers (with their own URLs and credentials) and file search
// stores. The account may be shared by several tenants and the library
// cannot vouch for them, so a payload callback may not add them, whatever
// the binding's hosted tools; barness-ai never sends them.
var (
	geminiPayloadReferences = []string{"cachedContent"}
	geminiReferenceTools    = []string{"mcpServers", "fileSearch"}
)

// authorizeGeminiPayload checks a body a payload callback produced. The
// model is in the URL the adapter builds, so a body may only name the
// authorized one. Tools may declare functions and the hosted tool kinds the
// binding allows (a Gemini tool names its kind by its field, such as
// googleSearch); nothing may reference vendor-side state: none of the fields
// above, and no fileData (a Files API reference) anywhere in the system
// instruction or contents (ADR-0005, ADR-0012).
func authorizeGeminiPayload(m Model, hostedTools []string) func(map[string]any) string {
	return func(body map[string]any) string {
		if model, ok := body["model"]; ok && model != m.ID && model != "models/"+m.ID {
			return "payload callback may not change the authorized model"
		}
		for _, field := range geminiPayloadReferences {
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
				for kind := range obj {
					switch {
					case kind == "functionDeclarations":
					case slices.Contains(geminiReferenceTools, kind):
						return "payload callback may not reference vendor-side state (" + kind + ")"
					case !slices.Contains(hostedTools, kind):
						return "payload callback may only declare functions and the binding's hosted tools"
					}
				}
			}
		}
		for _, field := range []string{"systemInstruction", "contents"} {
			if geminiFileReference(body[field]) {
				return "payload callback may not reference vendor-side state (fileData)"
			}
		}
		return ""
	}
}

// geminiFileReference finds a fileData part anywhere in v.
func geminiFileReference(v any) bool {
	switch v := v.(type) {
	case []any:
		return slices.ContainsFunc(v, geminiFileReference)
	case map[string]any:
		if _, ok := v["fileData"]; ok {
			return true
		}
		for _, e := range v {
			if geminiFileReference(e) {
				return true
			}
		}
	}
	return false
}
