package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The approved seam is the public CLI process. Missing or invalid external
// inputs must stop expensive commands before any child Go tool is started.
func TestReleaseGatePreflightStopsCommands(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "releasegate")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = "../..", releaseToolEnv()
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	for _, mode := range []string{"default", "preflight", "fifo-trace", "fifo-trace-bundle", "missing-trace", "duplicate-matrix", "miscased-matrix", "unknown-matrix", "bad-snapshot", "pending-ledger", "collect-all", "conflicting-mode", "race-without-from"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-preflight-"+mode)
			work := t.TempDir()
			log := filepath.Join(work, "commands")
			fake := filepath.Join(work, "go")
			if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"${0%/*}/commands\"\nexit 1\n"), 0700); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(work, "out")
			args := []string{"-out", out}
			switch mode {
			case "preflight":
				args = append(args, "-preflight")
			case "collect-all":
				args = append(args, "-collect-all")
			case "missing-trace":
				args = append(args, "-trace", filepath.Join(work, "missing.json"))
			case "duplicate-matrix", "miscased-matrix", "unknown-matrix":
				raw, err := os.ReadFile("../live/support-matrix.json")
				if err != nil {
					t.Fatal(err)
				}
				field := map[string]string{"duplicate-matrix": `"schema": 1,`, "miscased-matrix": `"Schema": 1,`, "unknown-matrix": `"unexpected": 1,`}[mode]
				raw = []byte(strings.Replace(string(raw), "{", "{"+field, 1))
				p := filepath.Join(work, "matrix.json")
				if err := os.WriteFile(p, raw, 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-matrix", p)
			case "fifo-trace", "fifo-trace-bundle":
				p := filepath.Join(work, "trace.json")
				if err := syscall.Mkfifo(p, 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-trace", p)
				if mode == "fifo-trace-bundle" {
					args = append(args, "-bundle", work)
				}
			case "pending-ledger":
				raw, err := os.ReadFile("testdata/pidiff/ledger.json")
				if err != nil {
					t.Fatal(err)
				}
				var ledger map[string]any
				if err := json.Unmarshal(raw, &ledger); err != nil {
					t.Fatal(err)
				}
				ledger["decisions"].([]any)[0].(map[string]any)["classification"] = "pending"
				writeReleaseJSON(t, work, "ledger.json", ledger)
				args = append(args, "-ledger", filepath.Join(work, "ledger.json"))
			case "bad-snapshot":
				p := filepath.Join(work, "snapshot.json")
				if err := os.WriteFile(p, []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-snapshot", p)
			case "conflicting-mode":
				args = append(args, "-preflight", "-collect-all")
			case "race-without-from":
				args = append(args, "-stage", "race")
			}
			ctx, cancel := context.WithTimeout(ctxFor(t), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, args...)
			cmd.Dir = "../.."
			cmd.Env = append(releaseToolEnv(), "PATH="+work+string(os.PathListSeparator)+os.Getenv("PATH"), "RELEASE_COMMAND_LOG="+log)
			output, processErr := cmd.CombinedOutput()
			ev.Check("input refusal does not block on a file read", ctx.Err() == nil, "process exceeded bounded refusal time")
			_, logErr := os.Stat(log)
			ev.Record("process", map[string]any{"rejected": processErr != nil, "childStarted": logErr == nil, "panic": strings.Contains(string(output), "panic:")})
			if mode == "collect-all" {
				raw, _ := os.ReadFile(log)
				ev.Record("diagnostic-command-argv", string(raw))
				if data, err := os.ReadFile(filepath.Join(out, "execution-report.json")); err == nil {
					var report any
					if json.Unmarshal(data, &report) == nil {
						ev.Record("diagnostic-execution", report)
					}
				}
				ev.Check("explicit diagnostic continues only check commands", processErr != nil && logErr == nil && len(strings.Split(strings.TrimSpace(string(raw)), "\n")) == 3 && !strings.Contains(string(raw), "-race"), "diagnostic command sequence missing")
			} else {
				ev.Check("invalid inputs start no Go commands", processErr != nil && os.IsNotExist(logErr), "failed=%t childStarted=%t", processErr != nil, logErr == nil)
			}
			if mode == "conflicting-mode" || mode == "race-without-from" {
				return
			}
			if mode == "fifo-trace-bundle" {
				return
			}
			var preflight struct {
				Scope  string `json:"scope"`
				Ready  bool   `json:"ready"`
				Checks []struct {
					ID   string `json:"id"`
					Pass bool   `json:"pass"`
				} `json:"checks"`
			}
			raw, err := os.ReadFile(filepath.Join(out, "preflight-report.json"))
			decoded := err == nil && json.Unmarshal(raw, &preflight) == nil
			ev.Check("preflight rejection remains reviewable", decoded && preflight.Scope == "preflight" && !preflight.Ready && len(preflight.Checks) > 0, "report available=%t", decoded)
			if strings.HasSuffix(mode, "-matrix") {
				found := false
				for _, c := range preflight.Checks {
					found = found || c.ID == "input-loading" && !c.Pass
				}
				ev.Check("matrix retains strict native JSON contract", found, "invalid matrix crossed input boundary")
			}
			if mode == "pending-ledger" {
				found := false
				for _, c := range preflight.Checks {
					if c.ID == "difference-ledger" && !c.Pass {
						found = true
					}
				}
				ev.Check("pending difference rejected before tests", found, "missing ledger rejection")
			}
			var verdict struct {
				Pass bool `json:"pass"`
			}
			raw, err = os.ReadFile(filepath.Join(out, "release-report.json"))
			ev.Check("rejection never certifies release", err == nil && json.Unmarshal(raw, &verdict) == nil && !verdict.Pass, "release report available=%t", err == nil)
		})
	}
}

func releaseToolEnv() []string {
	var env []string
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+v)
		}
	}
	return env
}
