package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
	for _, mode := range []string{"default", "preflight", "missing-trace", "bad-snapshot", "pending-ledger", "conflicting-mode", "race-without-from"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-preflight-"+mode)
			work := t.TempDir()
			log := filepath.Join(work, "commands")
			fake := filepath.Join(work, "go")
			if err := os.WriteFile(fake, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$RELEASE_COMMAND_LOG\"\nexit 1\n"), 0700); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(work, "out")
			args := []string{"-out", out}
			switch mode {
			case "preflight":
				args = append(args, "-preflight")
			case "missing-trace":
				args = append(args, "-trace", filepath.Join(work, "missing.json"))
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
			cmd := exec.CommandContext(ctxFor(t), binary, args...)
			cmd.Dir = "../.."
			cmd.Env = append(releaseToolEnv(), "PATH="+work+string(os.PathListSeparator)+os.Getenv("PATH"), "RELEASE_COMMAND_LOG="+log)
			output, processErr := cmd.CombinedOutput()
			_, logErr := os.Stat(log)
			ev.Record("process", map[string]any{"rejected": processErr != nil, "childStarted": logErr == nil, "panic": strings.Contains(string(output), "panic:")})
			ev.Check("invalid inputs start no Go commands", processErr != nil && os.IsNotExist(logErr), "failed=%t childStarted=%t", processErr != nil, logErr == nil)
			if mode == "conflicting-mode" || mode == "race-without-from" {
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
