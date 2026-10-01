// Package pioracle is barness-ai's frozen pi-ai differential oracle (spec
// Testing Decisions §2, §5, §6). An E2E case drives the frozen pi-ai and
// barness-ai with the same logical input against the same response script;
// Compare reports every difference in what the local controlled Provider
// received and what each side emitted, and a maintained Ledger classifies
// each one. Any pending finding fails the protocol's differential gate.
//
// The frozen pi copy lives in node/ (see node/PROVENANCE.md): pi-ai 0.87.1
// pinned by lockfile, verified byte-identical to a source build of commit
// 898ab804 with its generated model data. It is rebuilt with `npm ci`; it never
// depends on a developer's research checkout.
package pioracle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnableEnv opts a test run into the differential. It needs Node.js and an
// installed node/ directory, so plain `go test ./...` reports it NOT_RUN.
const EnableEnv = "BARNESS_AI_PIDIFF"

// Enabled reports whether the differential was requested for this run.
func Enabled() bool { return os.Getenv(EnableEnv) == "1" }

// Provenance is node/provenance.json: where the frozen copy came from and the
// hashes a run must reproduce.
type Provenance struct {
	PiPackage        string `json:"pi_package"`
	PiVersion        string `json:"pi_version"`
	PiCommit         string `json:"pi_commit"`
	NpmGitHead       string `json:"npm_git_head"`
	TarballIntegrity string `json:"tarball_integrity"`
	ModelDataSHA256  string `json:"model_data_sha256"`
}

// Oracle runs the frozen pi copy.
type Oracle struct {
	dir  string
	node string
	prov Provenance
	sdks map[string]string
}

// Open locates the frozen copy and Node.js and checks the copy is installed.
// It does not install anything: rebuilding is an explicit maintainer step.
func Open() (*Oracle, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("pioracle: cannot locate package directory")
	}
	dir := filepath.Join(filepath.Dir(file), "node")
	raw, err := os.ReadFile(filepath.Join(dir, "provenance.json"))
	if err != nil {
		return nil, fmt.Errorf("pioracle: %w", err)
	}
	var prov Provenance
	if err := json.Unmarshal(raw, &prov); err != nil {
		return nil, fmt.Errorf("pioracle: provenance.json: %w", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "node_modules", "@earendil-works", "pi-ai", "package.json")); err != nil {
		return nil, fmt.Errorf("pioracle: frozen pi copy not installed; run `npm ci --ignore-scripts --prefix %s`", dir)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		return nil, fmt.Errorf("pioracle: Node.js not found: %w", err)
	}
	// The pi-side packages a case exercises; another protocol's SDK joins
	// when that protocol joins the oracle.
	sdks, err := lockedVersions(filepath.Join(dir, "package-lock.json"), "@earendil-works/pi-ai", "openai")
	if err != nil {
		return nil, err
	}
	return &Oracle{dir: dir, node: node, prov: prov, sdks: sdks}, nil
}

// Case is one runner invocation. Context is a pi Context (systemPrompt,
// messages, tools); Options are pi stream options without credentials or
// signal.
type Case struct {
	API      string          `json:"api"`
	Provider string          `json:"provider"`
	Model    string          `json:"model"`
	BaseURL  string          `json:"baseUrl"`
	APIKey   string          `json:"apiKey"`
	Entry    string          `json:"entry"`
	Context  json.RawMessage `json:"context"`
	Options  json.RawMessage `json:"options,omitempty"`
	// AbortAfterEvents > 0 aborts the call through its signal once that many
	// events were received.
	AbortAfterEvents int `json:"abortAfterEvents,omitempty"`
}

// Run is what pi observed for a case.
type Run struct {
	Pi struct {
		Version         string `json:"version"`
		ModelDataSHA256 string `json:"modelDataSha256"`
	} `json:"pi"`
	Node   string            `json:"node"`
	Events []json.RawMessage `json:"events"`
	Result json.RawMessage   `json:"result"`
}

// Run executes the case in a fresh Node.js process with an empty environment,
// so no credential, endpoint or proxy variable can reach pi. It fails when the
// installed copy no longer matches the recorded provenance.
func (o *Oracle) Run(ctx context.Context, c Case) (Run, error) {
	input, err := json.Marshal(c)
	if err != nil {
		return Run{}, err
	}
	cmd := exec.CommandContext(ctx, o.node, "runner.mjs")
	cmd.Dir = o.dir
	cmd.Env = []string{}
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		// The case input holds the test key; stderr is the runner's own
		// diagnostics and never echoes it.
		return Run{}, fmt.Errorf("pioracle: runner failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var out Run
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return Run{}, fmt.Errorf("pioracle: runner output: %w", err)
	}
	if out.Pi.Version != o.prov.PiVersion || out.Pi.ModelDataSHA256 != o.prov.ModelDataSHA256 {
		return Run{}, fmt.Errorf("pioracle: installed pi-ai %s (model data %s) does not match provenance %s (%s)",
			out.Pi.Version, out.Pi.ModelDataSHA256, o.prov.PiVersion, o.prov.ModelDataSHA256)
	}
	return out, nil
}

func lockedVersions(lockfile string, names ...string) (map[string]string, error) {
	raw, err := os.ReadFile(lockfile)
	if err != nil {
		return nil, fmt.Errorf("pioracle: %w", err)
	}
	var lock struct {
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(raw, &lock); err != nil {
		return nil, fmt.Errorf("pioracle: package-lock.json: %w", err)
	}
	out := map[string]string{}
	for _, n := range names {
		p, ok := lock.Packages["node_modules/"+n]
		if !ok {
			return nil, fmt.Errorf("pioracle: %s missing from package-lock.json", n)
		}
		out[n] = p.Version
	}
	return out, nil
}
