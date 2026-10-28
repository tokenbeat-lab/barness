package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// The shell limit causes a real first receipt write failure in the public CLI.
// It changes no production hook and all synthetic guards remain consumed.
func TestGrokHistoryPartialClaimConsumesApproval(t *testing.T) {
	privateGrantStore(t)
	ev := run.Case(t, "GROK-HISTORY-V6-PARTIAL-CLAIM")
	binary := grokV6GrantBinary(t)
	probe := exec.CommandContext(ctxFor(t), binary)
	probe.Dir, probe.Env = "../..", xaiProcessEnv(t)
	if !privateGrantProcess(t, ev, probe) {
		return
	}
	call := func(limited bool, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctxFor(t), binary, args...)
		if limited {
			cmd = exec.CommandContext(ctxFor(t), "/bin/sh", append([]string{"-c", `ulimit -f 0; exec "$@"`, "claim-write-limit", binary}, args...)...)
		}
		cmd.Dir, cmd.Env = "../..", xaiProcessEnv(t)
		raw, err := cmd.CombinedOutput()
		exit := -1
		if cmd.ProcessState != nil {
			exit = cmd.ProcessState.ExitCode()
		}
		ev.Record("process-"+time.Now().Format("150405.000000000"), map[string]any{"argv": cmd.Args, "exit": exit, "output": string(raw), "zeroFileSizeLimit": limited})
		return raw, err
	}
	raw, err := call(false, "-template", "openai-chat", "-mode", "controlled", "-stage", "candidate", "-binary", binary)
	if err != nil {
		t.Fatalf("template: %v %s", err, raw)
	}
	var grant map[string]any
	mustUnmarshal(t, raw, &grant)
	grant["id"] = "synthetic-partial-" + time.Now().Format("20060102150405000000000")
	grant["accountAlias"], grant["regionAlias"] = "synthetic-account", "synthetic-region"
	expiry := time.Now().Add(time.Hour).UTC()
	grant["expiresAt"] = expiry.Format(time.RFC3339Nano)
	file := filepath.Join(t.TempDir(), "grant.json")
	check := func(limited bool) error {
		raw, _ := json.Marshal(grant)
		if err := os.WriteFile(file, raw, 0600); err != nil {
			t.Fatal(err)
		}
		_, err := call(limited, "-check", file, "-route", "openai-chat", "-mode", "controlled", "-stage", "candidate", "-binary", binary, "-claim")
		return err
	}
	ev.Check("real first receipt write fails", check(true) != nil, "write unexpectedly succeeded")
	grant["expiresAt"] = expiry.Add(time.Minute).Format(time.RFC3339Nano)
	ev.Check("failed write consumes approval ID across changed expiry", check(false) != nil, "same approval reclaimed with new expiry")
	grant["expiresAt"] = expiry.Format(time.RFC3339Nano)
	grant["id"] = grant["id"].(string) + "-renamed"
	ev.Check("failed write also consumes original scope across renamed ID", check(false) != nil, "same scope reclaimed with new ID")
}
