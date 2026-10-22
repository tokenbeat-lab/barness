// Package hostconfig validates URL configuration shared by the explicit Host
// examples. It never infers a Provider from a hostname or chooses a fallback.
package hostconfig

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

var ErrEndpoint = errors.New("hostconfig: endpoint does not match explicitly authorized service/protocol")

func ValidateAPIs(endpoints map[ai.API]string) error {
	for api := range endpoints {
		if api != ai.APIOpenAICompletions && api != ai.APIOpenAIResponses && api != ai.APIAnthropicMessages {
			return ErrEndpoint
		}
	}
	return nil
}

// Endpoint validates official service and protocol, or an explicit trusted
// gateway. Both paths reject URL credentials, query parameters and fragments;
// HTTP is limited to controlled loopback and still needs Client opt-in.
func Endpoint(official, configured string, trustedGateway bool) (string, error) {
	if configured == "" {
		return official, nil
	}
	u, err := url.Parse(configured)
	if err != nil || u.Opaque != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawFragment != "" || strings.Contains(configured, "#") || u.Hostname() == "" || u.RawPath != "" {
		return "", ErrEndpoint
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		if u.Scheme != "http" || !trustedGateway || (u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback())) {
			return "", ErrEndpoint
		}
	}
	u.Path = strings.TrimRight(u.Path, "/")
	officialHost := false
	switch strings.TrimSuffix(strings.ToLower(u.Hostname()), ".") {
	case "open.bigmodel.cn", "api.z.ai", "api.minimax.io", "api.minimax.cn", "api.moonshot.ai", "api.moonshot.cn":
		officialHost = true
	}
	if (!trustedGateway || officialHost) && u.String() != official {
		return "", ErrEndpoint
	}
	return u.String(), nil
}
