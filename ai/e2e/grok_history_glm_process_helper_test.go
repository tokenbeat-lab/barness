package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"
)

// These are process budgets, not public Client call deadlines. Leave room at
// every layer for startup and sealed manifest/audit writes: native 180s < full
// wrapper 240s < nested Go 300s < collector 360s < offline workflow 480s.
const glmProcessCleanup = 7 * time.Second

func glmProcessContext(t *testing.T, budget time.Duration) context.Context {
	t.Helper()
	deadline := time.Now().Add(budget)
	if parent, ok := t.Deadline(); ok && parent.Add(-glmProcessCleanup).Before(deadline) {
		deadline = parent.Add(-glmProcessCleanup)
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	t.Cleanup(cancel)
	return ctx
}

type glmProcessResult struct {
	Argv             []string `json:"argv"`
	EnvironmentNames []string `json:"environmentNames"`
	Status           string   `json:"status"`
	ExitCode         *int     `json:"exitCode"`
	Error            string   `json:"error,omitempty"`
	ContextError     string   `json:"contextError,omitempty"`
	TimeoutSeconds   float64  `json:"timeoutSeconds"`
	ElapsedSeconds   float64  `json:"elapsedSeconds"`
	Stdout           string   `json:"stdout"`
	Stderr           string   `json:"stderr"`
	VendorCalls      int      `json:"vendorCalls"`
}

func (r glmProcessResult) success() bool {
	return r.Status == "PASS" && r.ExitCode != nil && *r.ExitCode == 0 && r.ContextError == ""
}
func (r glmProcessResult) output() []byte { return []byte(r.Stdout + r.Stderr) }

// Reuse the collector's bounded group capture. A direct CommandContext Kill
// only kills `go`, leaving its test executable able to write partial evidence.
// The Python bridge owns that group; interrupt it and wait for cleanup before
// returning. Keep this adapter local to the two real GLM process workflows.
func glmCapture(t *testing.T, ctx context.Context, root string, env []string, argv ...string) glmProcessResult {
	t.Helper()
	result := glmProcessObservation(ctx, env, argv)
	if err := ctx.Err(); err != nil {
		result.Status, result.ContextError = "CANCELLED", err.Error()
		return result
	}
	out := t.TempDir()
	request, err := json.Marshal(map[string]any{"argv": argv, "root": root, "out": out, "timeout": result.TimeoutSeconds + 1})
	if err != nil {
		t.Fatal(err)
	}
	const bridge = `import json, os, sys
from pathlib import Path
sys.path.insert(0, sys.argv[1])
from run_process import capture
request = json.load(sys.stdin)
state = capture(request['argv'], Path(request['root']), dict(os.environ),
                Path(request['out']), request['timeout'], [], shutdown_grace=2)
print(json.dumps(state))
`
	cmd := exec.CommandContext(ctx, "python3", "-c", bridge, filepath.Join(root, "tools/verification"))
	cmd.Dir, cmd.Env, cmd.Stdin = root, env, bytes.NewReader(request)
	// Context cancellation must reach capture before its fallback deadline.
	// Its own one-second-later bound also fits within the cleanup allowance.
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = glmProcessCleanup - time.Second
	start := time.Now()
	raw, processErr := cmd.CombinedOutput()
	result.ElapsedSeconds = time.Since(start).Seconds()
	var state struct {
		Status   string
		ExitCode *int
		Error    string
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		result.Status, result.Error = "CAPTURE_FAILED", "group capture did not finalize: "+string(raw)
	} else {
		result.Status, result.ExitCode, result.Error = state.Status, state.ExitCode, state.Error
	}
	if processErr != nil && result.Error == "" {
		result.Error = processErr.Error()
	}
	if err := ctx.Err(); err != nil {
		result.ContextError = err.Error()
	}
	stdout, _ := os.ReadFile(filepath.Join(out, "stdout.jsonl"))
	stderr, _ := os.ReadFile(filepath.Join(out, "stderr.log"))
	result.Stdout, result.Stderr = string(stdout), string(stderr)
	return result
}

// The collector already owns group capture. Do not place it under a second
// capture group that could SIGKILL it before its own cleanup has completed.
func glmCollect(t *testing.T, ctx context.Context, root string, env []string, argv ...string) glmProcessResult {
	t.Helper()
	result := glmProcessObservation(ctx, env, argv)
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir, cmd.Env = root, env
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = glmProcessCleanup - time.Second
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	start := time.Now()
	stdout, err := cmd.Output()
	result.ElapsedSeconds = time.Since(start).Seconds()
	result.Stdout, result.Stderr = string(stdout), stderr.String()
	if cmd.ProcessState != nil {
		exit := cmd.ProcessState.ExitCode()
		result.ExitCode = &exit
	}
	if err == nil {
		result.Status = "PASS"
	} else {
		result.Status, result.Error = "FAIL", err.Error()
	}
	if err := ctx.Err(); err != nil {
		result.Status, result.ContextError = "CANCELLED", err.Error()
	}
	return result
}

func glmProcessObservation(ctx context.Context, env, argv []string) glmProcessResult {
	deadline, _ := ctx.Deadline()
	result := glmProcessResult{Argv: argv, Status: "START_FAILED", TimeoutSeconds: time.Until(deadline).Seconds()}
	for _, entry := range env {
		result.EnvironmentNames = append(result.EnvironmentNames, strings.SplitN(entry, "=", 2)[0])
	}
	sort.Strings(result.EnvironmentNames)
	return result
}
