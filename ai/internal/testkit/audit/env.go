package audit

import (
	"os"
	"regexp"
	"sort"
	"strings"
)

// credentialVariable names environment variables whose values are treated
// as credentials.
var credentialVariable = regexp.MustCompile(`(?i)(KEY|TOKEN|SECRET|PASSWORD|PASSWD|CREDENTIAL|AUTH)`)

// Environment returns the auditing process's own values that must never
// reach a bundle: credential-looking environment variables, the home
// directory and the host name. Their presence means real, not synthetic,
// data was recorded. Labels name the variable, never the value.
func Environment() []Forbidden {
	var out []Forbidden
	for _, kv := range os.Environ() {
		name, value, ok := strings.Cut(kv, "=")
		if ok && credentialVariable.MatchString(name) && len(value) >= minForbidden {
			out = append(out, Forbidden{Value: value, Label: "environment variable " + name, Rule: RuleEnvironment})
		}
	}
	if home, err := os.UserHomeDir(); err == nil && len(home) >= minForbidden {
		out = append(out, Forbidden{Value: home, Label: "the home directory", Rule: RuleEnvironment})
	}
	if host, err := os.Hostname(); err == nil && len(host) >= minForbidden && host != "localhost" {
		out = append(out, Forbidden{Value: host, Label: "the host name", Rule: RuleEnvironment})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}
