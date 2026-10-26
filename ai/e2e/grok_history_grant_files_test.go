package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func grokV6GrantBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "processgrant")
	build := exec.Command("go", "build", "-o", binary, "./ai/live/cmd/processgrant")
	build.Dir, build.Env = "../..", xaiProcessEnv(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	return binary
}

func TestGrokHistoryGrantSpecialFiles(t *testing.T) {
	binary := grokV6GrantBinary(t)
	for _, target := range []string{"grant", "binary"} {
		for _, kind := range []string{"fifo", "directory", "symlink"} {
			t.Run(target+"/"+kind, func(t *testing.T) {
				ev := run.Case(t, "GROK-HISTORY-V6-FILE-"+target+"-"+kind)
				path := filepath.Join(t.TempDir(), "special")
				var err error
				switch kind {
				case "fifo":
					err = syscall.Mkfifo(path, 0600)
				case "directory":
					err = os.Mkdir(path, 0700)
				case "symlink":
					err = os.Symlink(binary, path)
				}
				if err != nil {
					t.Fatal(err)
				}
				args := []string{"-template", "openai-chat", "-mode", "controlled", "-stage", "candidate", "-binary", path}
				if target == "grant" {
					args = []string{"-check", path, "-route", "openai-chat", "-mode", "controlled", "-stage", "candidate", "-binary", binary}
				}
				ctx, cancel := context.WithTimeout(ctxFor(t), 3*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, binary, args...)
				cmd.Dir, cmd.Env = "../..", xaiProcessEnv(t)
				out, err := cmd.CombinedOutput()
				exit := -1
				if cmd.ProcessState != nil {
					exit = cmd.ProcessState.ExitCode()
				}
				ev.Record("process", map[string]any{"argv": cmd.Args, "output": string(out), "exit": exit, "deadline": ctx.Err() != nil})
				ev.Check("special file rejected without waiting for a writer", err != nil && ctx.Err() == nil && exit == 1, "err=%v deadline=%v exit=%d", err, ctx.Err(), exit)
			})
		}
	}
}
