package audit

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The redaction audit is tested in isolation (AGENTS.md 测试原则): these are
// the ways it could fail, enumerated before the implementation.
//
//  1. A registered secret (a test key, the live key) written verbatim into
//     any file goes unreported.
//  2. Key-shaped text (OpenAI/DeepSeek sk-, Anthropic sk-ant-, OpenAI
//     sk-proj-, Google AIza) that was never registered goes unreported.
//  3. Ordinary words that merely contain "sk-" (task-runner, disk-sync) are
//     reported, drowning real findings.
//  4. The redaction markers and test aliases themselves ([REDACTED],
//     [REDACTED-KEY], key:tenant-a@v1, [LIVE-KEY:combo]) are reported.
//  5. A request credential header (Authorization, x-api-key,
//     x-goog-api-key, api-key, cookie, proxy-authorization, any case, string
//     or list of strings) holding a real value goes unreported because the
//     value is not key-shaped.
//  6. A bearer token in free text (a Go map dump in an assertion detail)
//     goes unreported.
//  7. A key passed in a URL query (?key=, Gemini's form) goes unreported;
//     or the recorder's URL-encoded marker (key=%5BREDACTED%5D) is reported.
//  8. An observation record carrying content (text, arguments, error text:
//     any field outside the Observer's metadata) goes unreported.
//  9. A live capture keeping a response header value outside the allowlist
//     (organization, set-cookie) goes unreported.
// 10. A value from the auditing environment (a credential-looking variable,
//     the home directory, the host name) written into the bundle goes
//     unreported.
// 11. A finding repeats the secret it found, so the audit report leaks it.
// 12. A .json file that does not parse is silently skipped, so its JSON
//     rules never run.
// 13. A missing or empty bundle directory passes as "no findings".
// 14. Nested directories are not walked.

func writeBundle(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const manifest = `{"module":"barness-ai","cases":[]}`

func rules(fs []Finding) []Rule {
	var out []Rule
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return out
}

func mustAudit(t *testing.T, dir string, forbidden []Forbidden) []Finding {
	t.Helper()
	fs, err := Bundle(dir, forbidden)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	return fs
}

func TestRegisteredSecret(t *testing.T) { // 1, 11
	secret := "plainsecretvalue-0001"
	dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/result.txt": "body: " + secret})
	fs := mustAudit(t, dir, []Forbidden{{Value: secret, Label: "tenant-a key"}})
	if !slices.Contains(rules(fs), RuleSecret) {
		t.Fatalf("secret not reported: %+v", fs)
	}
	for _, f := range fs {
		if strings.Contains(f.Detail, secret) || strings.Contains(f.File, secret) {
			t.Fatalf("finding repeats the secret: %+v", f)
		}
	}
}

func TestKeyShapes(t *testing.T) { // 2, 11
	for _, key := range []string{
		"sk-test-tenant-a-0001",
		"sk-proj-AbCdEf0123456789xyz",
		"sk-ant-api03-AbCdEf0123456789",
		"AIzaSyTestTenantA0001geminiKey000000000",
		`he said "sk-abcdef1234567890"`,
	} {
		dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/x.json": `{"note":` + quote(key) + `}`})
		fs := mustAudit(t, dir, nil)
		if !slices.Contains(rules(fs), RuleKeyShaped) {
			t.Errorf("%q not reported: %+v", key, fs)
		}
		for _, f := range fs {
			if strings.Contains(f.Detail, "0123456789") || strings.Contains(f.Detail, "TenantA0001") || strings.Contains(f.Detail, "tenant-a-0001") {
				t.Errorf("finding repeats the key: %+v", f)
			}
		}
	}
}

func TestNoFalsePositives(t *testing.T) { // 3, 4, 7
	dir := writeBundle(t, map[string]string{
		"manifest.json": manifest,
		"c/failures.json": `{"body":"Rate limit reached for requests on task-runner and disk-sync",` +
			`"masked":"[REDACTED-KEY]","echo":"Incorrect API key provided: [REDACTED]",` +
			`"header":{"Authorization":"Bearer key:tenant-a@v1","X-Api-Key":"key:tenant-a-anthropic@v1",` +
			`"x-goog-api-key":["[REDACTED]"],"Cookie":""},"auth":"key:tenant-a@v1",` +
			`"live":{"authorization":"[REDACTED]","detail":"[LIVE-KEY:deepseek-chat]"},` +
			`"url":"https://generativelanguage.googleapis.com/v1beta/models/x:streamGenerateContent?alt=sse&key=%5BREDACTED%5D"}`,
		"c/assertions.json": `{"detail":"map[Authorization:Bearer key:tenant-a-deepseek@v1 Content-Type:application/json] bearer Authorization"}`,
	})
	if fs := mustAudit(t, dir, nil); len(fs) != 0 {
		t.Fatalf("clean bundle reported: %+v", fs)
	}
}

func TestCredentialHeaders(t *testing.T) { // 5
	for _, body := range []string{
		`{"headers":{"Authorization":"Bearer opaque.token.value"}}`,
		`{"headers":{"authorization":["Basic dXNlcjpwYXNz"]}}`,
		`{"header":{"X-Api-Key":"anthropic-raw"}}`,
		`{"requestHeaders":{"x-goog-api-key":"rawvalue"}}`,
		`{"h":{"Api-Key":"azure-style"}}`,
		`{"h":{"Cookie":"session=abc"}}`,
		`{"h":{"Proxy-Authorization":"Basic abc"}}`,
	} {
		dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/requests.json": body})
		fs := mustAudit(t, dir, nil)
		if !slices.Contains(rules(fs), RuleCredentialHeader) {
			t.Errorf("%s not reported: %+v", body, fs)
		}
	}
}

func TestBearerInText(t *testing.T) { // 6
	dir := writeBundle(t, map[string]string{"manifest.json": manifest,
		"c/assertions.json": `{"detail":"map[Authorization:Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig Content-Type:x]"}`})
	if fs := mustAudit(t, dir, nil); !slices.Contains(rules(fs), RuleBearer) {
		t.Fatalf("bearer token not reported: %+v", fs)
	}
}

func TestQueryKey(t *testing.T) { // 7
	dir := writeBundle(t, map[string]string{"manifest.json": manifest,
		"c/x.json": `{"url":"https://example.test/v1beta/models/m:generateContent?key=rawkeyvalue&alt=sse"}`})
	if fs := mustAudit(t, dir, nil); !slices.Contains(rules(fs), RuleCredentialQuery) {
		t.Fatalf("query key not reported: %+v", fs)
	}
}

func TestObservationContent(t *testing.T) { // 8
	clean := `[{"kind":"call_finished","time":"t","call":{"tenantId":"a","requestId":"r","bindingId":"b","resolved":true,` +
		`"attempts":[{"attemptId":"r#1","httpStatus":500,"code":"upstream_error","retryDelay":1,"usageReporting":"none",` +
		`"usage":{"input":1,"output":2,"cacheRead":0,"cacheWrite":0,"totalTokens":3,"cost":{"input":0,"output":0,"total":0}}}],` +
		`"nativeStateDowngrades":{"noEnvelope":1}},"duration":5,"stopReason":"error","error":{"code":"upstream_error","phase":"stream","httpStatus":500}}]`
	dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/observations-tenant-a.json": clean})
	if fs := mustAudit(t, dir, nil); len(fs) != 0 {
		t.Fatalf("clean observations reported: %+v", fs)
	}
	for _, leak := range []string{
		`[{"kind":"call_finished","error":{"code":"x","message":"provider said: secret prompt"}}]`,
		`[{"kind":"call_finished","text":"hello"}]`,
		`[{"kind":"attempt_finished","attempt":{"attemptId":"r#1","arguments":"{}"}}]`,
	} {
		dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/observations-x.json": leak})
		if fs := mustAudit(t, dir, nil); !slices.Contains(rules(fs), RuleObservationField) {
			t.Errorf("%s not reported: %+v", leak, fs)
		}
	}
}

func TestResponseHeaderAllowlist(t *testing.T) { // 9
	ok := `[{"responseHeaders":{"x-request-id":"req_1","content-type":"text/event-stream","retry-after":"1","x-ratelimit-remaining-requests":"9","cf-ray-trace-id":"t"}}]`
	dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/exchanges.json": ok})
	if fs := mustAudit(t, dir, nil); len(fs) != 0 {
		t.Fatalf("allowlisted headers reported: %+v", fs)
	}
	for _, name := range []string{"openai-organization", "set-cookie", "openai-project"} {
		dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/exchanges.json": `[{"responseHeaders":{"` + name + `":"v"}}]`})
		if fs := mustAudit(t, dir, nil); !slices.Contains(rules(fs), RuleResponseHeader) {
			t.Errorf("%s kept but not reported: %+v", name, fs)
		}
	}
}

func TestEnvironmentValues(t *testing.T) { // 10, 11
	t.Setenv("SOME_SERVICE_TOKEN", "envtokenvalue-777")
	t.Setenv("HARMLESS_SETTING", "envtokenvalue-888")
	env := Environment()
	if !slices.ContainsFunc(env, func(f Forbidden) bool { return f.Value == "envtokenvalue-777" }) {
		t.Fatalf("credential-looking variable not collected: %+v", labels(env))
	}
	if slices.ContainsFunc(env, func(f Forbidden) bool { return f.Value == "envtokenvalue-888" }) {
		t.Fatalf("ordinary variable collected")
	}
	home, _ := os.UserHomeDir()
	if home != "" && !slices.ContainsFunc(env, func(f Forbidden) bool { return f.Value == home }) {
		t.Fatalf("home directory not collected")
	}
	dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/x.txt": "token envtokenvalue-777 in " + home + "/project"})
	fs := mustAudit(t, dir, env)
	if n := len(slices.DeleteFunc(slices.Clone(fs), func(f Finding) bool { return f.Rule != RuleEnvironment })); n < 2 {
		t.Fatalf("environment values not reported: %+v", fs)
	}
	for _, f := range fs {
		if strings.Contains(f.Detail, "envtokenvalue-777") {
			t.Fatalf("finding repeats the value: %+v", f)
		}
	}
}

func TestUnparseableJSON(t *testing.T) { // 12
	dir := writeBundle(t, map[string]string{"manifest.json": manifest, "c/requests.json": `{"headers":{"Authorization":"Bearer raw`})
	if fs := mustAudit(t, dir, nil); !slices.Contains(rules(fs), RuleUnparseable) {
		t.Fatalf("broken JSON not reported: %+v", fs)
	}
}

func TestMissingBundle(t *testing.T) { // 13
	if _, err := Bundle(filepath.Join(t.TempDir(), "absent"), nil); err == nil {
		t.Fatal("missing bundle passed")
	}
	if _, err := Bundle(t.TempDir(), nil); err == nil {
		t.Fatal("empty bundle passed")
	}
	if _, err := Bundle(writeBundle(t, map[string]string{"c/x.json": `{}`}), nil); err == nil {
		t.Fatal("bundle without manifest passed")
	}
}

func TestNestedFiles(t *testing.T) { // 14
	dir := writeBundle(t, map[string]string{"manifest.json": manifest, "a/b/c/deep.json": `{"k":"sk-abcdef1234567890"}`})
	fs := mustAudit(t, dir, nil)
	if len(fs) != 1 || fs[0].File != filepath.Join("a", "b", "c", "deep.json") {
		t.Fatalf("nested file not audited or not named relative to the bundle: %+v", fs)
	}
}

func quote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func labels(fs []Forbidden) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Label)
	}
	return out
}
