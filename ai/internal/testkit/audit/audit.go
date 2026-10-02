// Package audit is barness-ai's redaction audit of evidence bundles (spec
// Testing Decisions §6, issue 24): it proves that no key, Authorization value
// or content from outside the synthetic fixtures reached a bundle, offline or
// live, including its observation records and error captures.
//
// Bundles are redacted at the source (evidence.Run.RedactSecret, the live
// recorder); the audit reads what was actually written and reports every
// violation. Its rules:
//
//   - RuleSecret: a registered value (a test key, the live key) appears.
//   - RuleKeyShaped: text shaped like a first-phase vendor key appears.
//   - RuleCredentialHeader: a request credential header in a JSON capture
//     holds anything but a redaction marker or a test alias.
//   - RuleBearer: a bearer token appears in free text.
//   - RuleCredentialQuery: a URL carries a key query parameter that is not
//     redacted.
//   - RuleObservationField: an observation record (observations*.json) has
//     a field outside the Observer's metadata, i.e. it may carry content.
//   - RuleResponseHeader: a live capture kept a response header value
//     outside the recorder's allowlist.
//   - RuleEnvironment: a value from the auditing environment (a
//     credential-looking variable, the home directory, the host name)
//     appears, i.e. real data rather than synthetic fixtures.
//   - RuleUnparseable: a .json file does not parse, so its JSON rules could
//     not run.
//
// A finding names the file, the rule and where; it never repeats the value
// it found.
package audit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Rule names one audit rule.
type Rule string

// The audit rules; see the package documentation.
const (
	RuleSecret           Rule = "secret"
	RuleKeyShaped        Rule = "key-shaped"
	RuleCredentialHeader Rule = "credential-header"
	RuleBearer           Rule = "bearer-token"
	RuleCredentialQuery  Rule = "credential-query"
	RuleObservationField Rule = "observation-field"
	RuleResponseHeader   Rule = "response-header"
	RuleEnvironment      Rule = "environment"
	RuleUnparseable      Rule = "unparseable"
)

// Finding is one violation. Detail never contains the offending value.
type Finding struct {
	File   string `json:"file"`
	Rule   Rule   `json:"rule"`
	Detail string `json:"detail"`
}

// Forbidden is a value that must not appear anywhere in a bundle; Label
// names it in findings instead of the value.
type Forbidden struct {
	Value string
	Label string
	// Rule is RuleSecret when unset.
	Rule Rule
}

// minForbidden is the shortest value worth forbidding: shorter values would
// match ordinary text.
const minForbidden = 6

// Bundle audits every file below dir, which must be an evidence bundle (it
// holds manifest.json), against the built-in rules and the forbidden values.
// Findings are sorted by file.
func Bundle(dir string, forbidden []Forbidden) ([]Finding, error) {
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
		return nil, fmt.Errorf("audit: %s is not an evidence bundle: %w", dir, err)
	}
	var findings []Finding
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		findings = append(findings, File(rel, data, forbidden)...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("audit: %w", err)
	}
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].File < findings[j].File })
	return findings, nil
}

// File audits one file's content; name is its path within the bundle.
func File(name string, data []byte, forbidden []Forbidden) []Finding {
	a := &fileAudit{name: name}
	for _, f := range forbidden {
		if len(f.Value) < minForbidden {
			continue
		}
		if n := bytes.Count(data, []byte(f.Value)); n > 0 {
			rule := f.Rule
			if rule == "" {
				rule = RuleSecret
			}
			a.add(rule, "%s appears %d time(s)", f.Label, n)
		}
	}
	text := string(data)
	for _, m := range keyShaped.FindAllStringSubmatchIndex(text, -1) {
		a.add(RuleKeyShaped, "key-shaped text at byte %d: %s", m[4], mask(text[m[4]:m[5]]))
	}
	for _, m := range bearer.FindAllStringSubmatchIndex(text, -1) {
		if token := text[m[2]:m[3]]; !redacted(token) {
			a.add(RuleBearer, "bearer token at byte %d: %s", m[0], mask(token))
		}
	}
	for _, m := range queryKey.FindAllStringSubmatchIndex(text, -1) {
		if v := text[m[2]:m[3]]; !redactedQuery(v) {
			a.add(RuleCredentialQuery, "key query parameter at byte %d: %s", m[0], mask(v))
		}
	}
	if strings.HasSuffix(name, ".json") {
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			a.add(RuleUnparseable, "not JSON (%s): JSON rules not applied", errorKind(err))
		} else {
			a.walk(v, "$", strings.HasPrefix(filepath.Base(name), "observations"))
		}
	}
	return a.findings
}

type fileAudit struct {
	name     string
	findings []Finding
}

func (a *fileAudit) add(rule Rule, format string, args ...any) {
	a.findings = append(a.findings, Finding{File: a.name, Rule: rule, Detail: fmt.Sprintf(format, args...)})
}

// walk applies the JSON rules. observation is set for an Observer record
// file, whose every field must be metadata.
func (a *fileAudit) walk(v any, path string, observation bool) {
	switch v := v.(type) {
	case map[string]any:
		for _, k := range sortedKeys(v) {
			child, at := v[k], path+"."+k
			if observation && !observationFields[k] {
				a.add(RuleObservationField, "%s is not an Observer metadata field", at)
			}
			if credentialHeader.MatchString(k) {
				for _, s := range headerValues(child) {
					if !redacted(s) {
						a.add(RuleCredentialHeader, "%s holds a value that is neither redacted nor a test alias: %s", at, mask(s))
					}
				}
			}
			if k == "responseHeaders" {
				if headers, ok := child.(map[string]any); ok {
					for _, name := range sortedKeys(headers) {
						if !keptResponseHeader.MatchString(name) {
							a.add(RuleResponseHeader, "%s.%s keeps a value outside the allowlist", at, name)
						}
					}
				}
			}
			a.walk(child, at, observation)
		}
	case []any:
		for i, child := range v {
			a.walk(child, fmt.Sprintf("%s[%d]", path, i), observation)
		}
	}
}

// headerValues returns a header value's strings: a string or a list of them.
func headerValues(v any) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, e := range v {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// keyShaped matches first-phase vendor key forms: OpenAI and DeepSeek
// (sk-, sk-proj-), Anthropic (sk-ant-), Google (AIza). The leading boundary
// keeps words such as "task-runner" out.
var keyShaped = regexp.MustCompile(`(^|[^A-Za-z0-9_\-])(sk-[A-Za-z0-9_\-]{8,}|AIza[0-9A-Za-z_\-]{20,})`)

// bearer matches a bearer token in free text; tokens shorter than 16
// characters are words ("bearer Authorization"), not credentials.
var bearer = regexp.MustCompile(`Bearer\s+([A-Za-z0-9._~+/=:\-]{16,}|\[[^\]]*\])`)

// queryKey matches a key query parameter (Gemini's key form).
var queryKey = regexp.MustCompile(`[?&]key=([^&"\s\\]+)`)

// credentialHeader names the request headers that carry a credential in any
// first-phase protocol. A response's Set-Cookie is the server's, not a
// credential barness holds: offline fixtures set synthetic ones (the cookie
// jar hardening case), and a live capture never keeps its value
// (RuleResponseHeader).
var credentialHeader = regexp.MustCompile(`(?i)^(authorization|proxy-authorization|x-api-key|x-goog-api-key|api-key|cookie)$`)

// keptResponseHeader is the live recorder's allowlist of response header
// values: vendor request and trace ids, content type, retry and rate-limit
// state (ai/live recorder_test.go keptHeaderValue).
var keptResponseHeader = regexp.MustCompile(`(?i)(request[-_]?id|trace[-_]?id|^content-type$|^retry-after|ratelimit)`)

// alias matches the redaction markers and test aliases evidence uses in
// place of credentials.
var alias = regexp.MustCompile(`^(Bearer |Basic )?(\[[A-Z][A-Z0-9_\-]*(:[^\]]*)?\]|key:[A-Za-z0-9._@\-]+)$`)

func redacted(s string) bool { return s == "" || alias.MatchString(s) }

func redactedQuery(v string) bool {
	return strings.HasPrefix(v, "%5BREDACTED") || strings.HasPrefix(v, "[REDACTED") || strings.HasPrefix(v, "key%3A") || strings.HasPrefix(v, "key:")
}

// observationFields are the JSON fields of ai.Observation and the metadata
// it nests (CallMetadata, Attempt, ObservedError, Usage, cost, downgrade
// counts). A new Observer field must be reviewed for content and added here.
var observationFields = map[string]bool{
	"kind": true, "time": true, "call": true, "attempt": true, "duration": true, "stopReason": true, "error": true, "usage": true,
	"tenantId": true, "requestId": true, "actorId": true, "jobId": true, "bindingId": true, "resolved": true,
	"providerId": true, "api": true, "modelId": true, "accountScopeId": true, "bindingVersion": true,
	"credentialVersion": true, "catalogVersion": true, "catalogHash": true, "attempts": true, "nativeStateDowngrades": true,
	"noEnvelope": true, "accountMismatch": true, "crossModel": true,
	"attemptId": true, "httpStatus": true, "providerRequestId": true, "code": true, "phase": true, "retryAfter": true,
	"retryDelay": true, "usageReporting": true,
	"input": true, "output": true, "cacheRead": true, "cacheWrite": true, "cacheWrite1h": true, "reasoning": true,
	"totalTokens": true, "cost": true, "total": true,
}

// mask shows only a value's shape: its first three characters and length.
func mask(s string) string {
	s = strings.TrimLeft(s, " \"'=:")
	if len(s) <= 3 {
		return fmt.Sprintf("(%d chars)", len(s))
	}
	return fmt.Sprintf("%s… (%d chars)", s[:3], len(s))
}

func errorKind(err error) string {
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		return fmt.Sprintf("syntax error at byte %d", syntax.Offset)
	}
	return "invalid"
}
