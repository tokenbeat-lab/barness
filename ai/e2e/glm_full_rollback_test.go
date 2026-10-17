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

func TestGLMFullRollback(t *testing.T) {
	ev := run.Case(t, "P0-GLM13-full-rollback")
	rel := filepath.Join(".evidence", "glm13-rollback-"+requestid.New())
	abs, err := filepath.Abs(filepath.Join("../..", rel))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(abs); err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", ".scratch/barness-ai-glm/evidence/ticket-13/rollback.py", rel)
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	output, executeErr := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"success": executeErr == nil, "panic": strings.Contains(string(output), "panic:"), "credentialsInherited": false})
	if !ev.Check("whole public restoration process succeeds", executeErr == nil, "err=%v output=%s", executeErr, output) {
		return
	}
	for _, name := range []string{"rollback-report.json", "restored-tree.json", "root-before.json", "root-after.json", "commands.json", "module-pins.json", "baseline-release-index.json", "public-no-glm-probe.log"} {
		raw, err := os.ReadFile(filepath.Join(abs, name))
		if err != nil {
			t.Fatal(err)
		}
		ev.Fixture(name, raw)
	}
	var commands []struct {
		Argv []string `json:"argv"`
	}
	commandBytes, err := os.ReadFile(filepath.Join(abs, "commands.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, commandBytes, &commands)
	for _, command := range commands {
		for _, arg := range command.Argv {
			ev.Check("rollback receipt paths are portable and redacted", !filepath.IsAbs(arg), "absolute argument leaked")
		}
	}
	var report struct {
		Pass           bool `json:"passVerdict"`
		Baseline, Tree string
		Files          int  `json:"completeTrackedFiles"`
		RootUnchanged  bool `json:"rootUnchanged"`
	}
	raw, err := os.ReadFile(filepath.Join(abs, "rollback-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	ev.Check("complete fixed baseline restored", report.Pass && report.RootUnchanged && report.Files == 6682 && fields["baselineCommit"] == "f5cf5362610a8bdb8f3e170f408f3858c149e540" && fields["baselineTree"] == "e0e006866c606fdab3bf02be0322bde2eae02300", "report=%+v", report)
}
