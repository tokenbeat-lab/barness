package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

func TestMiniMaxFullRollback(t *testing.T) {
	ev := run.Case(t, "MM08-full-software-rollback")
	rel := filepath.Join(".evidence", "minimax08-rollback-"+requestid.New())
	abs, err := filepath.Abs(filepath.Join("../..", rel))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(abs); err != nil {
			t.Error(err)
		}
	})
	seed := filepath.Join("../..", ".scratch", "minimax08-preserve-"+requestid.New())
	seedBytes := []byte("Owned rollback preservation fixture.\n")
	if err := os.WriteFile(seed, seedBytes, 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(seed); err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", ".scratch/barness-ai-minimax/evidence/ticket-08/rollback.py", rel)
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	output, executeErr := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"success": executeErr == nil, "panic": strings.Contains(string(output), "panic:"), "credentialsInherited": false})
	if !ev.Check("whole software rollback process succeeds", executeErr == nil, "%v %s", executeErr, output) {
		return
	}
	for _, name := range []string{"rollback-report.json", "restored-tree.json", "root-before.json", "root-after.json", "commands.json", "module-pins.json", "baseline-release-index.json", "public-no-minimax-probe.log", "baseline-old-provider-summary.json"} {
		raw, err := os.ReadFile(filepath.Join(abs, name))
		if err != nil {
			t.Fatal(err)
		}
		ev.Fixture(name, raw)
	}
	seedAfter, err := os.ReadFile(seed)
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("owned untracked caller file preserved", string(seedAfter) == string(seedBytes), "untracked file changed")
	var report struct {
		Pass                                                                                  bool   `json:"passVerdict"`
		Baseline                                                                              string `json:"baselineCommit"`
		Tree                                                                                  string `json:"baselineTree"`
		Files                                                                                 int    `json:"completeTrackedFiles"`
		RootUnchanged, RestoredTreeVerified, PublicNoMiniMaxRefusal, OldProviderControlledE2E bool
	}
	raw, err := os.ReadFile(filepath.Join(abs, "rollback-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &report)
	ev.Check("fixed complete baseline, old route and absent MiniMax verified", report.Pass && report.Baseline == "ecfd4b4d6ec54db30453f4b5112989e9fdbd0782" && report.Tree == "b9798a4e96dfb08cbfcce45c182915589b42eed1" && report.Files == 7039 && report.RootUnchanged && report.RestoredTreeVerified && report.PublicNoMiniMaxRefusal && report.OldProviderControlledE2E, "%+v", report)
	var commands []struct{ Argv []string }
	raw, err = os.ReadFile(filepath.Join(abs, "commands.json"))
	if err != nil {
		t.Fatal(err)
	}
	if json.Unmarshal(raw, &commands) != nil {
		t.Fatal("invalid commands")
	}
	for _, command := range commands {
		for _, arg := range command.Argv {
			ev.Check("portable redacted command receipt", !filepath.IsAbs(arg), "absolute argument leaked")
		}
	}
}
