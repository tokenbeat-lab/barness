package supportmatrix

// ValidIdentity checks the published route, independent of a report's own
// claims. Two equally incorrect records cannot validate each other.
func (r Row) ValidIdentity() bool {
	if route, ok := RegionalRouteOf(r.Combo); ok {
		return r.Operation == "chat" && r.Provider == route.Provider && r.API == route.API
	}
	var operation, provider, api string
	switch r.Combo {
	case "xai-responses":
		operation, provider, api = "chat", "xai", "openai-responses"
	case "openai-responses":
		operation, provider, api = "chat", "openai", "openai-responses"
	case "minimax-chat":
		operation, provider, api = "chat", "minimax", "openai-completions"
	case "minimax-responses":
		operation, provider, api = "chat", "minimax", "openai-responses"
	case "minimax-messages":
		operation, provider, api = "chat", "minimax", "anthropic-messages"
	case "zhipu-chat":
		operation, provider, api = "chat", "zhipu", "openai-completions"
	case "zhipu-responses":
		operation, provider, api = "chat", "zhipu", "openai-responses"
	case "zhipu-messages":
		operation, provider, api = "chat", "zhipu", "anthropic-messages"
	case "openai-chat":
		operation, provider, api = "chat", "openai", "openai-completions"
	case "anthropic-messages":
		operation, provider, api = "chat", "anthropic", "anthropic-messages"
	case "google-gemini":
		operation, provider, api = "chat", "google", "google-generative-ai"
	case "deepseek-responses":
		operation, provider, api = "chat", "deepseek", "openai-responses"
	case "deepseek-chat":
		operation, provider, api = "chat", "deepseek", "openai-completions"
	case "openai-images":
		operation, provider, api = "image", "openai", "openai-images"
	case "google-interactions-image":
		operation, provider, api = "image", "google", "google-interactions"
	case "typesafe-classifier", JevFormalCombo:
		operation, provider, api = "classifier", "typesafe", "typesafe-system-one"
	case "openai-decisions":
		operation, provider, api = "classifier", "openai", "openai-decisions"
	default:
		return false
	}
	return r.Operation == operation && r.Provider == provider && r.API == api
}
