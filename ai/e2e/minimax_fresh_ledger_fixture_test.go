package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func miniMaxFreshLedgerCommand(t *testing.T, binary, route, approval, base string) *exec.Cmd {
	t.Helper()
	cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestMiniMaxCandidate$", "-test.timeout=1800s")
	cmd.Dir, cmd.Env = "../..", append(xaiProcessEnv(t), "BARNESS_AI_LIVE=1", "BARNESS_MINIMAX_ROUTE="+route, "BARNESS_MINIMAX_MODEL=MiniMax-M3", "BARNESS_MINIMAX_PURPOSE=standard", "BARNESS_MINIMAX_KEY=synthetic-fresh-minimax-key", "BARNESS_MINIMAX_ENTITLEMENT=confirmed", "BARNESS_MINIMAX_CONTROLLED=success-"+route, "BARNESS_MINIMAX_ACCOUNT_ALIAS=synthetic-account", "BARNESS_MINIMAX_REGION_ALIAS=synthetic-region", "BARNESS_MINIMAX_GRANT="+approval, "BARNESS_AI_EVIDENCE_DIR="+base)
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
