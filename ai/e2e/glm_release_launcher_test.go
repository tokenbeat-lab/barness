package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Actual launcher CLIs run in an owned minimal Git checkout with a no-network
// Go executable. Neither a review refusal nor script drift may reach execution.
func TestGLMReleaseLauncherGuards(t *testing.T) {
	for _, mode := range []string{"intact", "commands.py", "run.py", "glm_launcher.py", "review-denied"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "P0-GLM13-launcher-"+mode)
			root := t.TempDir()
			scripts := filepath.Join(root, ".scratch/barness-ai-glm/evidence/ticket-13")
			if err := os.MkdirAll(scripts, 0700); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"commands.py", "run.py", "glm_launcher.py", "rollback.py", "rollback-probe.txt"} {
				raw, err := os.ReadFile(filepath.Join("../../.scratch/barness-ai-glm/evidence/ticket-13", name))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(scripts, name), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll(filepath.Join(root, "ai"), 0700); err != nil {
				t.Fatal(err)
			}
			os.WriteFile(filepath.Join(root, "ai/fixture.go"), []byte("package fixture\n"), 0600)
			os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixture\n"), 0600)
			os.WriteFile(filepath.Join(root, "go.sum"), nil, 0600)
			fakebin := filepath.Join(root, "bin")
			os.Mkdir(fakebin, 0700)
			os.WriteFile(filepath.Join(fakebin, "go"), []byte("#!/bin/sh\nexit 0\n"), 0700)
			command := func(name string, args ...string) ([]byte, error) {
				cmd := exec.CommandContext(ctxFor(t), name, args...)
				cmd.Dir = root
				for _, key := range []string{"HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
					if v, ok := os.LookupEnv(key); ok {
						cmd.Env = append(cmd.Env, key+"="+v)
					}
				}
				cmd.Env = append(cmd.Env, "PATH="+fakebin+string(os.PathListSeparator)+os.Getenv("PATH"), "PYTHONDONTWRITEBYTECODE=1")
				return cmd.CombinedOutput()
			}
			for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
				if raw, err := command("git", args...); err != nil {
					t.Fatalf("git fixture: %v %s", err, raw)
				}
			}
			out := filepath.Join(root, ".evidence/round")
			os.MkdirAll(out, 0700)
			os.WriteFile(filepath.Join(out, "live.test"), []byte("frozen fixture binary"), 0600)
			// The fixture freezes actual launcher-owned file hashes through its normal
			// source recorder. Only the execution boundary is replaced by fake Go.
			setup := `import sys;sys.path.insert(0,'.scratch/barness-ai-glm/evidence/ticket-13');import commands as c;c.dump(c.ROOT/'.evidence/round/source-at-start.json',{'files':c.sources(),'gitCommit':'fixture-source'});c.dump(c.ROOT/'.evidence/round/binaries.json',{'live.test':c.digest(c.ROOT/'.evidence/round/live.test')});c.dump(c.ROOT/'.evidence/round/live-harness-accepted.json',{'exitCode':0});c.dump(c.ROOT/'.evidence/round/prerelease-source-review.json',{'sourceCommit':'fixture-source','decision':'DENIED'})`
			if raw, err := command("python3", "-c", setup); err != nil {
				t.Fatalf("freeze: %v %s", err, raw)
			}
			if mode != "intact" && mode != "review-denied" {
				path := filepath.Join(scripts, mode)
				f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				f.WriteString("\n# post-review mutation\n")
				f.Close()
			}
			action := "check"
			if mode == "review-denied" {
				action = "live"
			}
			raw, err := command("python3", filepath.Join(scripts, "run.py"), action, ".evidence/round")
			_, startedErr := os.Stat(filepath.Join(out, "paid-round-started.json"))
			_, executedErr := os.Stat(filepath.Join(out, "live-harness-command.json"))
			ev.Record("process", map[string]any{"success": err == nil, "paidIntentCreated": startedErr == nil, "childCommandCreated": executedErr == nil, "credentialsInherited": false})
			if mode == "intact" {
				ev.Check("intact committed execution set can run no-key check", err == nil && executedErr == nil, "process failed: %s", raw)
			} else {
				ev.Check("review or source guard stops before child/paid intent", err != nil && os.IsNotExist(startedErr) && os.IsNotExist(executedErr), "mode=%s unexpectedly proceeded", mode)
			}
			var source map[string]any
			data, readErr := os.ReadFile(filepath.Join(out, "source-at-start.json"))
			if readErr != nil || json.Unmarshal(data, &source) != nil {
				t.Fatal("source missing")
			}
			ev.Record("frozen-source-count", len(source["files"].(map[string]any)))
		})
	}
}
