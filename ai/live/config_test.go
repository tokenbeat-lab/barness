//go:build live

package live

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
)

// Environment variables of a live process (see the package documentation).
const (
	envLive    = "BARNESS_AI_LIVE"
	envCombo   = "BARNESS_AI_LIVE_COMBO"
	envAlias   = "BARNESS_AI_LIVE_ACCOUNT_ALIAS"
	envKeyBase = "BARNESS_AI_LIVE_KEY_"
)

// processConfig is what this process was given. Exactly one combination
// runs; every other is NOT_RUN.
type processConfig struct {
	// notRun, when set, is why nothing runs: the switch is off, no
	// combination was chosen or its key is missing. Not an error.
	notRun string
	// refused, when set, is a misconfiguration that fails the run: an
	// unknown combination, a missing alias or another combination's key.
	refused string
	combo   *combo
	alias   string
	proxy   *url.URL
	key     string
}

// keyVar is the one variable a combination's key is read from.
func keyVar(name string) string {
	return envKeyBase + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

// loadConfig reads the process environment. It reads only the variables
// named above: no vendor key variable, endpoint variable or credential file.
func loadConfig() processConfig {
	if os.Getenv(envCombo) == decisionsCandidate.name || os.Getenv(envCombo) == decisionsFormal.name {
		c := &decisionsCandidate
		if os.Getenv(envCombo) == decisionsFormal.name {
			c = &decisionsFormal
		}
		out := processConfig{combo: c, alias: os.Getenv(envAlias), key: strings.TrimSpace(os.Getenv(keyVar(c.name)))}
		if os.Getenv(envLive) != "1" {
			out.notRun = envLive + "=1 is not set"
		} else if out.key == "" {
			out.notRun = "candidate key is not set"
		}
		return out
	}
	if os.Getenv(keyVar(decisionsCandidate.name)) != "" || os.Getenv(keyVar(decisionsFormal.name)) != "" {
		return processConfig{refused: "candidate credential must be isolated in its own process"}
	}
	if os.Getenv(envLive) != "1" {
		return processConfig{notRun: envLive + "=1 is not set; the live smoke never runs implicitly"}
	}
	name := os.Getenv(envCombo)
	if name == "" {
		return processConfig{notRun: envCombo + " is not set; one process runs one combination"}
	}
	i := slices.IndexFunc(combos, func(c combo) bool { return c.name == name })
	if i < 0 {
		return processConfig{refused: fmt.Sprintf("%s=%q is not a first-phase combination", envCombo, name)}
	}
	c := &combos[i]
	cfg := processConfig{combo: c, alias: os.Getenv(envAlias)}
	for _, other := range combos {
		if other.name != c.name && os.Getenv(keyVar(other.name)) != "" {
			// The secret store injects each key into its own combination's
			// process only; seeing another one means it was not.
			cfg.refused = fmt.Sprintf("%s is set in the %s process; inject each key only into its own combination's process", keyVar(other.name), c.name)
			return cfg
		}
	}
	if raw := os.Getenv("BARNESS_AI_LIVE_PROXY"); raw != "" {
		var err error
		cfg.proxy, err = smokeProxy(raw)
		if err != nil {
			cfg.refused = err.Error()
			return cfg
		}
	}
	cfg.key = strings.TrimSpace(os.Getenv(keyVar(c.name)))
	if cfg.key == "" {
		cfg.notRun = keyVar(c.name) + " is not set; without a key the combination is NOT_RUN"
		return cfg
	}
	if strings.TrimSpace(cfg.alias) == "" {
		cfg.refused = envAlias + " is required with a key: the support matrix records which test account and region ran"
	}
	return cfg
}
