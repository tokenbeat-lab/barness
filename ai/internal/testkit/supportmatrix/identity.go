package supportmatrix

// ValidIdentity checks the published route, independent of a report's own
// claims. Two equally incorrect records cannot validate each other.
func (r Row) ValidIdentity() bool {
	var operation, provider, api string
	switch r.Combo {
	case "openai-responses":
		operation, provider, api = "chat", "openai", "openai-responses"
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
	case "typesafe-classifier":
		operation, provider, api = "classifier", "typesafe", "typesafe-system-one"
	default:
		return false
	}
	return r.Operation == operation && r.Provider == provider && r.API == api
}
