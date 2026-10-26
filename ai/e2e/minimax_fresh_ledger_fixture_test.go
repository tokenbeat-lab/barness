package e2e

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func miniMaxFreshLedgerCommand(t *testing.T, binary, route, approval, base string) *exec.Cmd {
	t.Helper()
	ns := miniMaxPrivateSourceNamespace(t)
	// Parent Verify and its child must resolve the same owned authority store.
	// These tests are sequential; t.Setenv restores each subtest's environment.
	for _, pair := range os.Environ() {
		name, _, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(name, "GIT_") && name != "GIT_DIR" && name != "GIT_WORK_TREE" {
			t.Setenv(name, "")
			if err := os.Unsetenv(name); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Setenv("GIT_DIR", ns.dir)
	t.Setenv("GIT_WORK_TREE", ns.root)
	cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestMiniMaxCandidate$", "-test.timeout=1800s")
	cmd.Dir, cmd.Env = "../..", append(xaiProcessEnv(t), "BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE="+route, "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-fresh-minimax-key", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_CONTROLLED=success-"+route, "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=synthetic-region", "BARNESS_MINIMAX_GRANT="+approval, "BARNESS_AI_EVIDENCE_DIR="+base)
	cmd.Env = append(cmd.Env, "GIT_DIR="+ns.dir, "GIT_WORK_TREE="+ns.root)
	// This read-only check fails before a native process can be started if a
	// future environment change drops or overrides the private namespace.
	if got := miniMaxGitProbe(t, cmd, "rev-parse", "--path-format=absolute", "--git-common-dir"); got != ns.dir {
		t.Fatalf("controlled native store escaped private namespace: %s", got)
	}
	return cmd
}

type miniMaxSourceNamespace struct{ dir, root, commit string }

// The private namespace shares only read access to genuine commit objects.
// Source bytes remain in this checkout; no synthetic commit or live proof is made.
func miniMaxPrivateSourceNamespace(t *testing.T) miniMaxSourceNamespace {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	env := xaiProcessEnv(t)
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.CommandContext(ctxFor(t), "git", args...)
		cmd.Dir, cmd.Env = root, env
		raw, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("private source Git %v: %v %s", args, err, raw)
		}
		return strings.TrimSpace(string(raw))
	}
	commit := git("rev-parse", "HEAD")
	objects := git("rev-parse", "--path-format=absolute", "--git-path", "objects")
	dir := filepath.Join(t.TempDir(), "source.git")
	git("init", "--bare", "--initial-branch=fixture", dir)
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "objects", "info", "alternates"), []byte(objects+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git("--git-dir", dir, "config", "core.bare", "false")
	git("--git-dir", dir, "config", "core.worktree", root)
	git("--git-dir", dir, "update-ref", "refs/heads/fixture", commit)
	return miniMaxSourceNamespace{dir, root, commit}
}

func miniMaxGitProbe(t *testing.T, cmd *exec.Cmd, args ...string) string {
	t.Helper()
	probe := exec.CommandContext(ctxFor(t), "git", args...)
	probe.Dir, probe.Env = cmd.Dir, cmd.Env
	raw, err := probe.CombinedOutput()
	if err != nil {
		t.Fatalf("read-only child Git probe: %v %s", err, raw)
	}
	return strings.TrimSpace(string(raw))
}

func miniMaxEnvValue(env []string, name string) string {
	for _, pair := range env {
		key, value, _ := strings.Cut(pair, "=")
		if key == name {
			return value
		}
	}
	return ""
}

func miniMaxRecordPrivateStore(t *testing.T, ev *evidence.Case, cmd *exec.Cmd, nativeDir, phase string) map[string]string {
	t.Helper()
	common := miniMaxGitProbe(t, cmd, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if common != miniMaxEnvValue(cmd.Env, "GIT_DIR") || common == "" {
		t.Fatal("refuse to inspect any unowned grant store")
	}
	grantRaw, err := minimaxlive.ReadFile(miniMaxEnvValue(cmd.Env, "BARNESS_MINIMAX_GRANT"), 16<<10)
	if err != nil {
		t.Fatal(err)
	}
	var grant livegrant.Grant
	if json.Unmarshal(grantRaw, &grant) != nil {
		t.Fatal("invalid controlled fixture grant")
	}
	receiptRaw, err := minimaxlive.ReadFile(filepath.Join(nativeDir, "process-authorization-receipt.json"), 16<<10)
	var receipt livegrant.Receipt
	if err != nil || json.Unmarshal(receiptRaw, &receipt) != nil || receipt.GrantID != grant.ID || receipt.GrantSHA256 != livegrant.Hash(grantRaw) {
		t.Fatal("private inventory attribution differs from verified native grant")
	}
	scope, err := hex.DecodeString(receipt.ScopeSHA256)
	if err != nil || len(scope) != 32 {
		t.Fatal("invalid native scope digest")
	}
	want := map[string]bool{
		"lock": true, "id-" + livegrant.Hash([]byte(grant.ID)) + ".json": true,
		"scope-" + receipt.ScopeSHA256 + ".json": true, "ledger-" + receipt.ScopeSHA256 + ".jsonl": true,
	}
	store := filepath.Join(common, "barness-live-process-grants")
	entries, err := os.ReadDir(store)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !want[entry.Name()] {
			t.Fatal("unexpected private grant-store entry")
		}
		raw, err := os.ReadFile(filepath.Join(store, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		snapshot[entry.Name()] = evidence.Digest(raw)
		if entry.Name() != "lock" {
			ev.Fixture("private-"+phase+"-"+entry.Name(), raw)
		}
	}
	ev.Check("exact native identity, scope, ledger and lock "+phase, len(snapshot) == len(want), "files=%d expected=%d", len(snapshot), len(want))
	ev.Record("private-store-"+phase, map[string]any{"commonGitDir": common, "sourceCommit": miniMaxGitProbe(t, cmd, "rev-parse", "HEAD"), "sourceRoot": miniMaxGitProbe(t, cmd, "rev-parse", "--show-toplevel"), "files": snapshot})
	return snapshot
}

func miniMaxCheckRefusedRestart(t *testing.T, ev *evidence.Case, cmd *exec.Cmd, base, originalDir string, originalPID int, output []byte, processErr error) {
	t.Helper()
	var exit *exec.ExitError
	started := errors.As(processErr, &exit) && cmd.ProcessState != nil && cmd.ProcessState.Pid() != originalPID
	var exitCode any
	if started {
		exitCode = exit.ExitCode()
	}
	ev.Record("restart", map[string]any{"argv": cmd.Args, "path": cmd.Path, "dir": cmd.Dir, "environmentSHA256": livegrant.Hash([]byte(strings.Join(cmd.Env, "\x00"))), "commonGitDir": miniMaxEnvValue(cmd.Env, "GIT_DIR"), "originalPID": originalPID, "processStarted": started, "exitCode": exitCode, "error": fmt.Sprint(processErr), "output": regionalSafeOutput(output), "vendorCalls": 0})
	if !ev.Check("same approval restart is an actual native exit one", started && exit.ExitCode() == 1 && cmd.ProcessState.ExitCode() == 1, "started=%v exit=%v error=%v", started, exitCode, processErr) {
		return
	}
	paths, err := filepath.Glob(filepath.Join(base, "*", minimaxlive.ReportFile))
	if !ev.Check("restart retained its own native run", err == nil && len(paths) == 2, "reports=%d error=%v", len(paths), err) {
		return
	}
	var restartedDir string
	for _, path := range paths {
		if filepath.Dir(path) != originalDir {
			restartedDir = filepath.Dir(path)
		}
	}
	raw, err := os.ReadFile(filepath.Join(restartedDir, minimaxlive.ReportFile))
	var report minimaxlive.Report
	if err != nil || json.Unmarshal(raw, &report) != nil {
		t.Fatal("unreadable restarted native report")
	}
	const refusal = "fresh process approval already consumed or unavailable"
	ev.Record("restart-native-pointer", map[string]any{"dir": restartedDir, "runID": report.Session, "pid": cmd.ProcessState.Pid(), "exitCode": exit.ExitCode(), "failure": report.Failure, "budget": report.Budget, "calls": len(report.Calls)})
	ev.Check("restart is refused by native single-use admission", report.Session == filepath.Base(restartedDir) && report.Session != filepath.Base(originalDir) && report.Mode == "controlled" && report.Outcome == "FAIL" && !report.Complete && report.Failure == refusal && strings.Contains(string(output), refusal), "failure=%q session=%s", report.Failure, report.Session)
	ev.Check("rejected native restart made zero reservations or HTTP attempts", report.Budget == (minimaxlive.Budget{}) && report.Cumulative == (minimaxlive.Budget{}) && len(report.Calls) == 0 && report.ProcessGrantSHA256 == "", "budget=%+v calls=%d", report.Budget, len(report.Calls))
	var manifest struct{ Cases []json.RawMessage }
	raw, err = os.ReadFile(filepath.Join(restartedDir, "manifest.json"))
	ev.Check("rejected restart has no native scene or request captures", err == nil && json.Unmarshal(raw, &manifest) == nil && len(manifest.Cases) == 0, "unexpected restarted native cases")
	for _, name := range []string{"process-grant.json", "process-authorization-receipt.json"} {
		_, err := os.Stat(filepath.Join(restartedDir, name))
		ev.Check("rejected restart acquired no "+name, os.IsNotExist(err), "unexpected restarted authorization artifact")
	}
}
