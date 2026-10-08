// Package evidence writes the replayable, redacted evidence bundle that every
// barness-ai E2E run must leave behind (spec Testing Decisions §3, §6).
//
// A Run owns one bundle directory containing manifest.json plus one directory
// per case. Each case records its fixtures (with SHA-256), redacted request
// captures, response scripts, observed events and results, an assertion report
// and the command that replays exactly that scenario.
package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

// DirEnv overrides where bundles are written. By default bundles go to
// <repo>/.evidence/barness-ai/<run-id>, which is git-ignored.
const DirEnv = "BARNESS_AI_EVIDENCE_DIR"

// Run is one `go test` invocation's evidence bundle.
type Run struct {
	dir     string
	pkg     string
	started time.Time

	mu       sync.Mutex
	versions map[string]string
	cases    []*Case

	secretsMu sync.Mutex
	secrets   map[string]string
}

// NewRun creates the bundle directory. pkg is the go test package argument
// used in replay commands, with any build flags the package needs, e.g.
// "./ai/e2e" or "-tags live ./ai/live".
func NewRun(pkg string) (*Run, error) {
	started := time.Now().UTC()
	runID := started.Format("20060102T150405.000000000Z")
	base := os.Getenv(DirEnv)
	if base == "" {
		root, err := repoRoot()
		if err != nil {
			return nil, err
		}
		base = filepath.Join(root, ".evidence", "barness-ai")
	}
	dir := filepath.Join(base, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Run{dir: dir, pkg: pkg, started: started, versions: map[string]string{}, secrets: map[string]string{}}, nil
}

// Dir is the bundle directory.
func (r *Run) Dir() string { return r.dir }

// SetVersion records a version or hash (SDK version, catalog hash, ...) in the manifest.
func (r *Run) SetVersion(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.versions[key] = value
}

// RedactSecret registers a test secret that must never appear in the bundle.
// Every artifact written afterwards has it replaced by alias. This is a
// safety net; captures are expected to be redacted at the source already.
func (r *Run) RedactSecret(secret, alias string) {
	r.secretsMu.Lock()
	defer r.secretsMu.Unlock()
	r.secrets[secret] = alias
}

// Case starts the evidence for one scenario. id is the case identifier such as
// "P01-E01-text-stream-full"; the assertion report and summary are written when
// the test finishes.
func (r *Run) Case(t *testing.T, id string) *Case {
	t.Helper()
	c := &Case{run: r, t: t, id: id, dir: filepath.Join(r.dir, id), hashes: map[string]string{}}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		t.Fatalf("evidence: %v", err)
	}
	r.mu.Lock()
	r.cases = append(r.cases, c)
	r.mu.Unlock()
	t.Cleanup(c.finish)
	return c
}

// Record stores a bundle-level artifact as <name>.json next to the manifest,
// redacted like every other artifact.
func (r *Run) Record(name string, v any) error {
	return r.writeJSON(filepath.Join(r.dir, name+".json"), v)
}

// Finish writes manifest.json. It reports the bundle location on stderr so a
// maintainer can find it from CI logs.
func (r *Run) Finish() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cases := make([]map[string]any, 0, len(r.cases))
	for _, c := range r.cases {
		artifacts, err := caseArtifacts(c.dir)
		if err != nil {
			return err
		}
		cases = append(cases, map[string]any{
			"id":        c.id,
			"test":      c.t.Name(),
			"status":    c.status,
			"dir":       c.id,
			"replay":    c.replay(),
			"fixtures":  c.hashes,
			"artifacts": artifacts,
		})
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i]["id"].(string) < cases[j]["id"].(string) })
	manifest := map[string]any{
		"module":      "barness-ai",
		"package":     r.pkg,
		"started_at":  r.started.Format(time.RFC3339Nano),
		"finished_at": time.Now().UTC().Format(time.RFC3339Nano),
		"go_version":  runtime.Version(),
		"platform":    runtime.GOOS + "/" + runtime.GOARCH,
		"git_commit":  GitCommit(),
		"versions":    r.versions,
		"cases":       cases,
	}
	if err := r.writeJSON(filepath.Join(r.dir, "manifest.json"), manifest); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "barness-ai evidence bundle: %s\n", r.dir)
	return nil
}

// Audit runs the redaction audit over the finished bundle (spec Testing
// Decisions §6): every registered secret, the auditing environment's own
// credentials, home directory and host name, and the built-in key,
// credential header, observation and response header rules. It records the
// findings, which never repeat a value, as audit.json and returns them; a
// run with findings must fail. Call it after Finish.
func (r *Run) Audit() ([]audit.Finding, error) {
	r.secretsMu.Lock()
	forbidden := audit.Environment()
	for secret, alias := range r.secrets {
		forbidden = append(forbidden, audit.Forbidden{Value: secret, Label: "the secret aliased " + alias})
	}
	r.secretsMu.Unlock()
	findings, err := audit.Bundle(r.dir, forbidden)
	if err != nil {
		return nil, err
	}
	if findings == nil {
		findings = []audit.Finding{}
	}
	if err := r.Record("audit", map[string]any{"findings": findings}); err != nil {
		return nil, err
	}
	return findings, nil
}

func (r *Run) writeJSON(path string, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return r.writeFile(path, buf.Bytes())
}

func (r *Run) writeFile(path string, data []byte) error {
	r.secretsMu.Lock()
	for secret, alias := range r.secrets {
		data = bytes.ReplaceAll(data, []byte(secret), []byte(alias))
	}
	r.secretsMu.Unlock()
	return os.WriteFile(path, data, 0o644)
}

// Case is the evidence for one scenario.
type Case struct {
	run *Run
	t   *testing.T
	id  string
	dir string

	mu         sync.Mutex
	hashes     map[string]string
	assertions []assertion
	status     string
	replayEnv  string
	replayRun  string
	// unsupported is why the scenario's capability does not exist for its
	// target; set, a test that did not fail reports UNSUPPORTED.
	unsupported string
}

type assertion struct {
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

// ID is the case identifier.
func (c *Case) ID() string { return c.id }

// ReplayEnv prefixes the replay command with environment assignments the
// scenario needs to run at all, e.g. an opt-in switch.
func (c *Case) ReplayEnv(assignments string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.replayEnv = assignments
}

// ReplayRun selects a complete pipeline when a case requires earlier cases in
// the same process. It changes only the replay command, never what was executed.
func (c *Case) ReplayRun(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.replayRun = pattern
}

// Unsupported records that the scenario's capability does not exist for its
// target, e.g. a protocol that returns no reasoning to replay. A case that
// does not otherwise fail then reports UNSUPPORTED instead of PASS; it is
// neither a skip nor a pass.
func (c *Case) Unsupported(reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unsupported = reason
}

// T is the case's test.
func (c *Case) T() *testing.T { return c.t }

// Fixture stores an input fixture verbatim and records its SHA-256.
func (c *Case) Fixture(name string, data []byte) {
	c.t.Helper()
	sum := sha256.Sum256(data)
	c.mu.Lock()
	c.hashes[name] = hex.EncodeToString(sum[:])
	c.mu.Unlock()
	if err := c.run.writeFile(filepath.Join(c.dir, name), data); err != nil {
		c.t.Fatalf("evidence: %v", err)
	}
}

// Record stores an observed artifact (request capture, events, result) as JSON.
func (c *Case) Record(name string, v any) {
	c.t.Helper()
	if err := c.run.writeJSON(filepath.Join(c.dir, name+".json"), v); err != nil {
		c.t.Fatalf("evidence: %v", err)
	}
}

// Check records a named assertion and fails the test when ok is false. It
// returns ok so callers can stop dependent assertions.
func (c *Case) Check(name string, ok bool, detailFormat string, args ...any) bool {
	c.t.Helper()
	detail := fmt.Sprintf(detailFormat, args...)
	c.mu.Lock()
	c.assertions = append(c.assertions, assertion{Name: name, Pass: ok, Detail: detail})
	c.mu.Unlock()
	if !ok {
		c.t.Errorf("[%s] %s: %s", c.id, name, detail)
	}
	return ok
}

func (c *Case) finish() {
	c.mu.Lock()
	status := "PASS"
	switch {
	case c.t.Failed():
		status = "FAIL"
	case c.t.Skipped():
		// Not executed (e.g. an opt-in suite that was not requested); never
		// reported as a pass.
		status = "NOT_RUN"
	case c.unsupported != "":
		status = "UNSUPPORTED"
	}
	c.status = status
	report := map[string]any{
		"case_id":    c.id,
		"test":       c.t.Name(),
		"status":     status,
		"assertions": c.assertions,
		"replay":     c.replay(),
	}
	if c.unsupported != "" {
		report["unsupported"] = c.unsupported
	}
	c.mu.Unlock()
	if err := c.run.writeJSON(filepath.Join(c.dir, "assertions.json"), report); err != nil {
		c.t.Errorf("evidence: %v", err)
	}
}

// replay is the single-scenario replay command. Every subtest level is
// anchored so only this case runs.
func (c *Case) replay() string {
	parts := strings.Split(c.t.Name(), "/")
	for i, p := range parts {
		parts[i] = "^" + regexp.QuoteMeta(p) + "$"
	}
	pattern := strings.Join(parts, "/")
	if c.replayRun != "" {
		pattern = c.replayRun
	}
	quoted := "'" + strings.ReplaceAll(pattern, "'", "'\"'\"'") + "'"
	cmd := fmt.Sprintf("go test %s -count=1 -run %s", c.run.pkg, quoted)
	if c.replayEnv != "" {
		cmd = c.replayEnv + " " + cmd
	}
	return cmd
}

// ModuleVersion returns the version go.mod pins for module path. Test binaries
// carry no dependency build info, so go.mod is the source of truth.
func ModuleVersion(path string) string {
	root, err := repoRoot()
	if err != nil {
		return "unknown"
	}
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "require "))
		if len(fields) >= 2 && fields[0] == path {
			return fields[1]
		}
	}
	return "unknown"
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("evidence: go.mod not found above working directory")
		}
		dir = parent
	}
}

// GitCommit is the checked-out commit, or "unknown".
func GitCommit() string {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
