package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Invoke the actual CLI, with a fake child executable only at its external
// Go-command seam. The full gate below still runs real normal and race suites.
func TestGLMReleaseCommandEnvelope(t *testing.T) {
	ev := run.Case(t, "P0-GLM13-release-command-envelope")
	dir := t.TempDir()
	binary := filepath.Join(dir, "releasegate")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	fake := filepath.Join(dir, "go")
	log := filepath.Join(dir, "commands")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GLM13_COMMAND_LOG\"\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctxFor(t), binary, "-out", filepath.Join(dir, "gate"))
	cmd.Dir = "../.."
	for _, key := range []string{"HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "GLM13_COMMAND_LOG="+log)
	output, err := cmd.CombinedOutput()
	raw, readErr := os.ReadFile(log)
	ev.Record("child-command-argv", string(raw))
	ev.Record("process", map[string]any{"rejected": err != nil, "panic": strings.Contains(string(output), "panic:")})
	ev.Check("incomplete commands fail closed", err != nil && readErr == nil && !strings.Contains(string(output), "panic:"), "err=%v read=%v", err, readErr)
	ev.Check("normal bounded package envelope preregistered", strings.Contains(string(raw), "test -count=1 -timeout=20m ./..."), "normal command=%s", raw)
	ev.Check("race bounded package envelope preregistered", strings.Contains(string(raw), "test -race -count=1 -timeout=75m ./..."), "race command=%s", raw)
}
