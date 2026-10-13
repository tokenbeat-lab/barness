//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

// This is a host verification record, not a vendor permissions introspection
// response. The reference identifies the host's checked configuration; the
// report preserves that distinction. It contains aliases, never account IDs.
type decisionsPermissions struct {
	AccountAlias          string    `json:"accountAlias"`
	RegionAlias           string    `json:"regionAlias"`
	DedicatedTestKey      bool      `json:"dedicatedTestKey"`
	Model                 string    `json:"model"`
	Permissions           []string  `json:"permissions"`
	SpendLimitUSD         float64   `json:"spendLimitUSD"`
	VerifiedAt            time.Time `json:"verifiedAt"`
	VerificationReference string    `json:"verificationReference"`
}

var decisionsAlias = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._@-]{0,79}$`)

func decisionsConfiguration(rec *recorder) (*decisionsPermissions, []byte, error) {
	if err := decisionsEnvironment(); err != nil {
		return nil, nil, err
	}
	allowUnknown := os.Getenv("BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT")
	if allowUnknown != "" && allowUnknown != "0" && allowUnknown != "1" {
		return nil, nil, fmt.Errorf("invalid unknown-account opt-in")
	}
	// Maintainer authorization may allow missing account/region/permission
	// evidence. It never fabricates verification or ignores a supplied record.
	if allowUnknown == "1" && os.Getenv("BARNESS_AI_LIVE_PERMISSIONS_FILE") == "" {
		if err := configureDecisionsProxy(rec); err != nil {
			return nil, nil, err
		}
		return nil, nil, nil
	}
	alias, region := cfg.alias, os.Getenv("BARNESS_AI_LIVE_REGION_ALIAS")
	if !decisionsAlias.MatchString(alias) || !decisionsAlias.MatchString(region) {
		return nil, nil, fmt.Errorf("account and region aliases are required")
	}
	path := os.Getenv("BARNESS_AI_LIVE_PERMISSIONS_FILE")
	raw, err := readDecisionsFile(path, 16384)
	if err != nil {
		return nil, nil, fmt.Errorf("bounded host permission verification record is required")
	}
	// Exact canonical field names and duplicates are checked before Go's
	// case-insensitive typed decoder. The
	// remaining fields are scalars or a string list; nested objects are illegal.
	tokens := json.NewDecoder(strings.NewReader(string(raw)))
	first, err := tokens.Token()
	if err != nil || first != json.Delim('{') {
		return nil, nil, fmt.Errorf("host permission record must be a JSON object")
	}
	seen := map[string]bool{}
	for tokens.More() {
		key, err := tokens.Token()
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission record")
		}
		name, ok := key.(string)
		if !ok || !slices.Contains([]string{"accountAlias", "regionAlias", "dedicatedTestKey", "model", "permissions", "spendLimitUSD", "verifiedAt", "verificationReference"}, name) || seen[name] {
			return nil, nil, fmt.Errorf("unknown, noncanonical or duplicate permission field")
		}
		seen[name] = true
		var value json.RawMessage
		if tokens.Decode(&value) != nil {
			return nil, nil, fmt.Errorf("invalid permission field")
		}
	}
	if _, err = tokens.Token(); err != nil {
		return nil, nil, fmt.Errorf("invalid permission object")
	}
	var trailing any
	if tokens.Decode(&trailing) != io.EOF {
		return nil, nil, fmt.Errorf("trailing permission document")
	}
	var p decisionsPermissions
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if d.Decode(&p) != nil || p.AccountAlias != alias || p.RegionAlias != region || !p.DedicatedTestKey || p.Model != "gpt-6-luna" || !slices.Equal(p.Permissions, []string{"POST /v1/decisions"}) || p.SpendLimitUSD <= 0 || p.VerifiedAt.IsZero() || p.VerifiedAt.After(time.Now()) || strings.TrimSpace(p.VerificationReference) == "" {
		return nil, nil, fmt.Errorf("host permission record does not verify dedicated Decisions-only access, model and spend limit")
	}
	if err := configureDecisionsProxy(rec); err != nil {
		return nil, nil, err
	}
	return &p, raw, nil
}

func configureDecisionsProxy(rec *recorder) error {
	proxy := os.Getenv("BARNESS_AI_LIVE_PROXY")
	u, err := smokeProxy(proxy)
	if err != nil {
		return err
	}
	if transport, ok := rec.next.(*http.Transport); ok {
		transport.Proxy = http.ProxyURL(u)
	}
	// The proxy is trusted host transport configuration. The target remains
	// https://api.openai.com/v1/decisions, with ordinary TLS verification.
	return nil
}

func decisionsEnvironment() error {
	for _, kv := range os.Environ() {
		name, value, _ := strings.Cut(kv, "=")
		if value == "" {
			continue
		}
		if strings.HasPrefix(name, envKeyBase) && name != keyVar(cfg.combo.name) {
			return fmt.Errorf("foreign combination credential variable is set")
		}
		if slices.Contains([]string{"BARNESS_AI_LIVE_ENDPOINT", "OPENAI_BASE_URL", "OPENAI_API_BASE"}, name) {
			return fmt.Errorf("endpoint overrides are forbidden")
		}
		// A candidate process receives one injected key; it does not source .env.
		if slices.Contains([]string{"OPENAI_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GOOGLE_API_KEY", "DEEPSEEK_API_KEY", "TYPESAFE_KEY"}, name) {
			return fmt.Errorf("ambient vendor credentials are forbidden")
		}
	}
	return nil
}
