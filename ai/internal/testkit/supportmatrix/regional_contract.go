package supportmatrix

import "strings"

// RegionalRoute is an independent standard-API qualification target. Returning
// values keeps the route registry immutable and outside Host model metadata.
type RegionalRoute struct {
	Combo, Provider, API, Endpoint string
}

func RegionalRouteOf(combo string) (RegionalRoute, bool) {
	r := RegionalRoute{Combo: combo}
	switch combo {
	case "zai-chat":
		r.Provider, r.API, r.Endpoint = "zai", "openai-completions", "https://api.z.ai/api/paas/v4"
	case "zai-responses":
		r.Provider, r.API, r.Endpoint = "zai", "openai-responses", "https://api.z.ai/api/v1"
	case "zai-messages":
		r.Provider, r.API, r.Endpoint = "zai", "anthropic-messages", "https://api.z.ai/api/anthropic"
	case "minimax-cn-chat":
		r.Provider, r.API, r.Endpoint = "minimax-cn", "openai-completions", "https://api.minimax.cn/v1"
	case "minimax-cn-responses":
		r.Provider, r.API, r.Endpoint = "minimax-cn", "openai-responses", "https://api.minimax.cn/v1"
	case "minimax-cn-messages":
		r.Provider, r.API, r.Endpoint = "minimax-cn", "anthropic-messages", "https://api.minimax.cn/anthropic"
	default:
		return RegionalRoute{}, false
	}
	return r, true
}

func RegionalModels(combo string) []string {
	r, ok := RegionalRouteOf(combo)
	if !ok {
		return nil
	}
	if r.Provider == "zai" {
		return []string{"glm-5.3", "glm-5.2"}
	}
	return []string{"MiniMax-M3"}
}

// RegionalExpected is model-qualified. Shared adapters and another model's
// successful execution never satisfy a missing target capability.
func RegionalExpected(combo string) []string {
	var out []string
	for _, model := range RegionalModels(combo) {
		for _, kind := range strings.Fields("full-stream full-complete simple-stream simple-complete tool-call tool-result native-history cancel recovery") {
			out = append(out, model+"/"+kind)
		}
	}
	return out
}

// RegionalNativeSourceCase is fixed before any request, never selected by
// searching successful turns. ZAI Chat/Responses tool calls may omit reasoning
// even with thinking enabled; their fixed full-stream source must contain
// actual native reasoning. Tool-call/result remain independently required in the history.
func RegionalNativeSourceCase(combo, model string) string {
	kind := "tool-call"
	if combo == "zai-chat" || combo == "zai-responses" {
		kind = "full-stream"
	}
	return RegionalCaseID(combo, model+"/"+kind)
}

func regionalNativeSources(combo string) map[string]string {
	out := map[string]string{}
	for _, model := range RegionalModels(combo) {
		out[model] = RegionalNativeSourceCase(combo, model)
	}
	return out
}

// New namespace prevents an old brand/market authorization or remaining
// budget from silently authorizing a new regional qualification process.
func RegionalAuthorization(combo string) string { return "provider-sites/" + combo + "/standard/v1" }
