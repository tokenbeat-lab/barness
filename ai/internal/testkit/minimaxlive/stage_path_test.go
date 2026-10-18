package minimaxlive_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Auxiliary durable-ledger boundary: a second worktree cannot select a new
// budget root or escape20calls before any credential acquisition/transport.
func TestCanonicalStageSurvivesWorktrees(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, e := cmd.CombinedOutput(); e != nil {
			t.Fatalf("git: %v %s", e, out)
		}
	}
	git("init", "-q")
	if e := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module stage-fixture\n"), 0600); e != nil {
		t.Fatal(e)
	}
	git("add", "go.mod")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "stage fixture")
	other := filepath.Join(t.TempDir(), "other")
	git("worktree", "add", "-qb", "codex/stage-fixture", other)
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	var paths []string
	for i, dir := range []string{root, other} {
		cmd := exec.Command(exe, "-test.run=^TestCanonicalStageChild$", "-test.v")
		cmd.Dir = dir
		cmd.Env = []string{"BARNESS_STAGE_CHILD=" + fmt.Sprint(i), "PATH=" + os.Getenv("PATH")}
		out, e := cmd.CombinedOutput()
		if e != nil {
			t.Fatalf("child%d: %v %s", i, e, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "STAGE_PATH ") {
				paths = append(paths, strings.TrimPrefix(line, "STAGE_PATH "))
			}
		}
	}
	expected := filepath.Join(root, ".evidence/minimax-stages/candidate.jsonl")
	// macOS temp paths can be canonicalized through /private.
	expected, e = filepath.EvalSymlinks(expected)
	if e != nil {
		t.Fatal(e)
	}
	if len(paths) != 2 || paths[0] != expected || paths[1] != expected {
		t.Fatalf("noncanonical stage paths: %v want %s", paths, expected)
	}
}
func TestCanonicalStageChild(t *testing.T) {
	n := os.Getenv("BARNESS_STAGE_CHILD")
	if n == "" {
		t.Skip("isolated child only")
	}
	path, e := minimaxlive.StageLedgerPath("candidate")
	if e != nil {
		t.Fatal(e)
	}
	fmt.Println("STAGE_PATH " + path)
	id, e := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	if e != nil {
		t.Fatal(e)
	}
	ledger := minimaxlive.Ledger{Path: path, Stage: "candidate", Session: "isolated", Identity: id}
	if n == "0" {
		for i := 0; i < 20; i++ {
			if _, e := ledger.Reserve(fmt.Sprintf("once-%d", i), 1000, 2048); e != nil {
				t.Fatal(e)
			}
		}
	} else if _, e := ledger.Reserve("restart-bypass", 1000, 2048); e == nil {
		t.Fatal("second worktree reset identity budget")
	}
}
