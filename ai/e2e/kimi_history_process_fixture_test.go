package e2e

import (
	"bytes"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Unlike ephemeral test namespaces, composition evidence retains its private
// authority and producer binary so it can be inspected again after the test.
func kimiHistoryPrivateStore(t *testing.T) {
	t.Helper()
	root, _ := filepath.Abs("../..")
	env := regionalChildEnvironment()
	clean := []string{}
	for _, kv := range env {
		if !strings.HasPrefix(kv, "GIT_") {
			clean = append(clean, kv)
		}
	}
	git := func(args ...string) string {
		t.Helper()
		c := exec.CommandContext(ctxFor(t), "git", args...)
		c.Dir, c.Env = root, clean
		b, e := c.CombinedOutput()
		if e != nil {
			t.Fatal(e, string(b))
		}
		return strings.TrimSpace(string(b))
	}
	commit := git("rev-parse", "HEAD")
	objects := git("rev-parse", "--path-format=absolute", "--git-path", "objects")
	dir := filepath.Join(filepath.Dir(run.Dir()), "kimi-private-authority", filepath.Base(run.Dir())+"-"+evidence.Digest([]byte(t.Name()))[:12], "source.git")
	if e := os.MkdirAll(filepath.Dir(dir), 0700); e != nil {
		t.Fatal(e)
	}
	git("init", "--bare", "--initial-branch=fixture", dir)
	if e := os.WriteFile(filepath.Join(dir, "objects/info/alternates"), []byte(objects+"\n"), 0600); e != nil {
		t.Fatal(e)
	}
	git("--git-dir", dir, "config", "core.bare", "false")
	git("--git-dir", dir, "config", "core.worktree", root)
	git("--git-dir", dir, "update-ref", "refs/heads/fixture", commit)
	t.Setenv("GIT_DIR", dir)
	t.Setenv("GIT_WORK_TREE", root)
}
func kimiRetainedBinary(t *testing.T, path, label string) string {
	t.Helper()
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(filepath.Dir(run.Dir()), "kimi-producer-binaries", filepath.Base(run.Dir()))
	if e := os.MkdirAll(dir, 0700); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(dir, label+"-"+evidence.Digest(raw)+".test")
	if old, e := os.ReadFile(out); e == nil {
		if !bytes.Equal(old, raw) {
			t.Fatal("binary collision")
		}
		return out
	}
	if e := os.WriteFile(out, raw, 0700); e != nil {
		t.Fatal(e)
	}
	return out
}
func kimiProducerReceipt(t *testing.T, cmd *exec.Cmd, dir string, stdout, stderr []byte, start, finish time.Time) {
	t.Helper()
	base := filepath.Dir(dir)
	source, e := supportmatrix.RegionalExecutionSource()
	if e != nil {
		t.Fatal(e)
	}
	files := map[string]string{}
	for n, b := range map[string][]byte{"stdout.jsonl": []byte(regionalSafeOutput(stdout)), "stderr.log": []byte(regionalSafeOutput(stderr)), "exit-code": []byte("0\n")} {
		if e := os.WriteFile(filepath.Join(base, n), b, 0600); e != nil {
			t.Fatal(e)
		}
		files[n] = evidence.Digest(b)
	}
	manifest, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var r supportmatrix.Report
	b, e := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalReportFile))
	if e != nil {
		t.Fatal(e)
	}
	mustUnmarshal(t, b, &r)
	value := map[string]any{"schema": 1, "stage": "e2e", "route": r.Combo, "status": "PASS", "exitCode": cmd.ProcessState.ExitCode(), "argv": cmd.Args, "source": map[string]any{"commit": evidence.GitCommit(), "files": source}, "startedAt": start, "finishedAt": finish, "files": files, "bundles": []any{map[string]any{"path": filepath.Base(dir), "manifestSHA256": evidence.Digest(manifest)}}}
	if e := os.WriteFile(filepath.Join(base, "producer-receipt.json"), mustMarshal(t, value), 0600); e != nil {
		t.Fatal(e)
	}
}
