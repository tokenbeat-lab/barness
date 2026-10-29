package ai

import (
	"net/url"
	"strings"
)

// xaiBindingProblem fixes the global standard service whose prices and model
// bounds this integration implements. A custom transport does not expand it.
func xaiBindingProblem(b Binding) string {
	if b.ProviderID != ProviderXAI {
		return ""
	}
	if !((b.Operation == OperationChat && b.API == APIOpenAIResponses) || (b.Operation == OperationImage && b.API == API("xai-images"))) {
		return "xAI operation or protocol is outside the implemented subset"
	}
	u, err := url.Parse(b.Endpoint)
	if err != nil || u.Scheme != "https" || (!strings.EqualFold(u.Host, "api.x.ai") && !strings.EqualFold(u.Host, "api.x.ai:443")) || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(b.Endpoint, "#") || u.RawPath != "" || (u.Path != "/v1" && u.Path != "/v1/") {
		return "xAI requires the global standard API endpoint"
	}
	return ""
}
